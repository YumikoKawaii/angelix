// Package patch applies binary patches to the Claude Code Bun SFE binary.
// It injects JS instrumentation for OTEL rate-limit telemetry using Claude's own tracer.
package patch

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnsurePatch returns a path to a patched copy of the claude binary at src.
// The result is cached under ~/.angelix/patched/<sha256>/claude and reused
// on subsequent calls as long as the source binary is unchanged.
func EnsurePatch(src string) (string, error) {
	hash, err := sha256File(src)
	if err != nil {
		return "", fmt.Errorf("hash source binary: %w", err)
	}

	cacheDir, err := cacheBase()
	if err != nil {
		return "", err
	}
	dest := filepath.Join(cacheDir, hash, "claude")

	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("read source binary: %w", err)
	}

	if err := applyPatches(data); err != nil {
		return "", fmt.Errorf("patch binary: %w", err)
	}

	if err := os.WriteFile(dest, data, 0755); err != nil {
		return "", fmt.Errorf("write patched binary: %w", err)
	}

	if runtime.GOOS == "darwin" {
		if err := codesign(dest); err != nil {
			_ = os.Remove(dest)
			return "", fmt.Errorf("codesign: %w", err)
		}
	}

	return dest, nil
}

// applyPatches modifies the Mach-O __bun section JS bundle in-place.
// It injects a JS IIFE at the end of the cli.js module so the code runs
// within the module's evaluation scope, then updates the module table and
// footer fields. The section grows by n bytes, consuming null padding.
func applyPatches(data []byte) error {
	sec, err := findBunSection(data)
	if err != nil {
		return err
	}

	// Section layout (confirmed from hex dump):
	//   [0:8]                  uint64 LE bundle_length (= section_size - 8)
	//   [8 : section_size-52]  JS + module data
	//   [section_size-52 : section_size]  52-byte Bun footer
	bundleLen := binary.LittleEndian.Uint64(data[sec.fileOff : sec.fileOff+8])
	footerStart := sec.fileOff + sec.size - 52
	footerEnd := sec.fileOff + sec.size
	jsStart := sec.fileOff + 8

	if footerEnd > uint64(len(data)) {
		return fmt.Errorf("section extends beyond file")
	}
	if string(data[footerStart+36:footerStart+52]) != "\n---- Bun! ----\n" {
		return fmt.Errorf("Bun footer magic not found — unsupported claude version")
	}

	maxN := sec.segFilesize - sec.size
	if maxN == 0 {
		return fmt.Errorf("no null padding in __BUN segment — cannot patch")
	}

	// Footer layout:
	//   [0:4]   padding/version
	//   [4:12]  uint64 field A — pointer within/near module table
	//   [12:16] uint32 field B — module table start offset (relative to jsStart)
	//   [16:24] uint64 field C — module table size in bytes (do NOT update)
	//   [24:32] uint64 field D — module table end offset (relative to jsStart)
	//   [32:36] padding
	//   [36:52] "\n---- Bun! ----\n" magic
	moduleTableOff := uint64(binary.LittleEndian.Uint32(data[footerStart+12 : footerStart+16]))
	moduleTableEnd := binary.LittleEndian.Uint64(data[footerStart+24 : footerStart+32])

	// Parse the cli.js module entry (first 16 bytes of module table):
	//   [0:4]  uint32 path_offset (relative to jsStart) — points to module path string
	//   [4:8]  uint32 path_length
	//   [8:12] uint32 code_offset (relative to jsStart) — start of JS code
	//   [12:16] uint32 code_size  — length of JS code
	mtAbs := jsStart + moduleTableOff
	if mtAbs+16 > uint64(len(data)) {
		return fmt.Errorf("module table truncated")
	}
	cliCodeOff := uint64(binary.LittleEndian.Uint32(data[mtAbs+8 : mtAbs+12]))
	cliCodeSize := uint64(binary.LittleEndian.Uint32(data[mtAbs+12 : mtAbs+16]))

	// The injection point is immediately after the cli.js JS source.
	// cliCodeOff and cliCodeSize are jsStart-relative.
	injPoint := cliCodeOff + cliCodeSize // jsStart-relative

	if injPoint > moduleTableOff {
		return fmt.Errorf("cli.js code (%d+%d) extends past module table (%d)", cliCodeOff, cliCodeSize, moduleTableOff)
	}

	// Build injection from the cli.js code region.
	cliCodeAbs := jsStart + cliCodeOff
	if cliCodeAbs+cliCodeSize > uint64(len(data)) {
		return fmt.Errorf("cli.js code region extends beyond file")
	}
	cliCode := data[cliCodeAbs : cliCodeAbs+cliCodeSize]

	newCli, err := patchJS(cliCode)
	if err != nil {
		return err
	}
	n := uint64(len(newCli)) - cliCodeSize
	if n == 0 {
		return nil
	}
	if n > maxN {
		return fmt.Errorf("injection (%d bytes) exceeds available padding (%d bytes)", n, maxN)
	}

	// Absolute file offset of the injection point.
	injAbs := jsStart + injPoint

	// Move the footer n bytes forward into the null padding (existing mechanism).
	copy(data[footerStart+n:footerEnd+n], data[footerStart:footerEnd])

	// Shift the region [injAbs, footerStart) forward by n bytes.
	// This region contains all modules after cli.js, the module table, etc.
	// We use a temporary copy to avoid overwrite corruption on a forward shift.
	shiftSize := footerStart - injAbs
	if shiftSize > 0 {
		tmp := make([]byte, shiftSize)
		copy(tmp, data[injAbs:footerStart])
		copy(data[injAbs+n:footerStart+n], tmp)
	}

	// Write the patched cli.js code (original + injection) at the original location.
	copy(data[cliCodeAbs:], newCli)

	// Update the module table in-place at its new position.
	// All uint32 jsStart-relative offsets in [injPoint, moduleTableOff) shift by +n.
	// cli.js code_size (mt[12:16]) increases by n.
	newMTAbs := jsStart + moduleTableOff + n
	mtSize := moduleTableEnd - moduleTableOff
	if newMTAbs+mtSize <= uint64(len(data)) {
		mt := data[newMTAbs : newMTAbs+mtSize]

		// Extend cli.js code_size.
		oldCodeSize := binary.LittleEndian.Uint32(mt[12:16])
		binary.LittleEndian.PutUint32(mt[12:16], oldCodeSize+uint32(n))

		// Shift all offsets pointing into the area after the injection.
		for i := 0; i+4 <= len(mt); i += 4 {
			v := uint64(binary.LittleEndian.Uint32(mt[i : i+4]))
			if v >= injPoint && v < moduleTableOff {
				binary.LittleEndian.PutUint32(mt[i:i+4], uint32(v+n))
			}
		}
	}

	// Update bundle_length.
	binary.LittleEndian.PutUint64(data[sec.fileOff:sec.fileOff+8], bundleLen+n)

	// Update footer fields A, B, D. All lie at or after the injection point.
	newFooter := footerStart + n
	oldA := binary.LittleEndian.Uint64(data[newFooter+4 : newFooter+12])
	binary.LittleEndian.PutUint64(data[newFooter+4:newFooter+12], oldA+n)
	oldB := binary.LittleEndian.Uint32(data[newFooter+12 : newFooter+16])
	binary.LittleEndian.PutUint32(data[newFooter+12:newFooter+16], oldB+uint32(n))
	oldD := binary.LittleEndian.Uint64(data[newFooter+24 : newFooter+32])
	binary.LittleEndian.PutUint64(data[newFooter+24:newFooter+32], oldD+n)

	// Update Mach-O section header size field.
	binary.LittleEndian.PutUint64(data[sec.sizeFieldOff:sec.sizeFieldOff+8], sec.size+n)

	return nil
}

// bunSection holds the parsed metadata of the __bun section.
type bunSection struct {
	fileOff      uint64
	size         uint64
	sizeFieldOff uint64
	segFilesize  uint64
}

// findBunSection locates the __bun section in a Mach-O binary and returns its layout.
func findBunSection(data []byte) (*bunSection, error) {
	needle := []byte("__bun\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
		"__BUN\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	idx := indexBytes(data, needle)
	if idx < 0 {
		return nil, fmt.Errorf("__bun section not found — not a Bun SFE or unsupported architecture")
	}

	// section64 layout (80 bytes total):
	//   [0:16]  sectname
	//   [16:32] segname
	//   [32:40] uint64 addr
	//   [40:48] uint64 size   ← update target
	//   [48:52] uint32 offset ← file offset of section data
	if idx+80 > len(data) {
		return nil, fmt.Errorf("section64 header truncated")
	}
	sizeFieldOff := uint64(idx + 40)
	secSize := binary.LittleEndian.Uint64(data[idx+40 : idx+48])
	secOff := uint64(binary.LittleEndian.Uint32(data[idx+48 : idx+52]))

	segFilesize, err := findBunSegmentFilesize(data)
	if err != nil {
		return nil, err
	}

	return &bunSection{
		fileOff:      secOff,
		size:         secSize,
		sizeFieldOff: sizeFieldOff,
		segFilesize:  segFilesize,
	}, nil
}

// findBunSegmentFilesize returns the filesize field of the LC_SEGMENT_64 for __BUN.
func findBunSegmentFilesize(data []byte) (uint64, error) {
	const lcSegment64 = 0x19
	if len(data) < 32 {
		return 0, fmt.Errorf("binary too small for Mach-O header")
	}
	ncmds := binary.LittleEndian.Uint32(data[16:20])
	off := uint32(32)
	for i := uint32(0); i < ncmds; i++ {
		if int(off+8) > len(data) {
			break
		}
		cmd := binary.LittleEndian.Uint32(data[off : off+4])
		cmdsize := binary.LittleEndian.Uint32(data[off+4 : off+8])
		if cmd == lcSegment64 && int(off+64) <= len(data) {
			segname := data[off+8 : off+24]
			if string(segname[:5]) == "__BUN" {
				filesize := binary.LittleEndian.Uint64(data[off+48 : off+56])
				return filesize, nil
			}
		}
		if cmdsize < 8 {
			return 0, fmt.Errorf("malformed load command at offset %d", off)
		}
		off += cmdsize
	}
	return 0, fmt.Errorf("__BUN segment not found in Mach-O load commands")
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func cacheBase() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".angelix", "patched"), nil
}

func codesign(path string) error {
	if err := runCmd("codesign", "--remove-signature", path); err != nil {
		return fmt.Errorf("remove-signature: %w", err)
	}
	return runCmd("codesign", "-s", "-", path)
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// indexBytes returns the index of the first occurrence of needle in haystack, or -1.
func indexBytes(haystack, needle []byte) int {
	if len(needle) == 0 {
		return 0
	}
	hlen := len(haystack) - len(needle)
	for i := 0; i <= hlen; i++ {
		if haystack[i] == needle[0] {
			match := true
			for j := 1; j < len(needle); j++ {
				if haystack[i+j] != needle[j] {
					match = false
					break
				}
			}
			if match {
				return i
			}
		}
	}
	return -1
}
