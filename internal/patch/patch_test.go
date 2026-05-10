package patch

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"testing"
)

func claudePath(t *testing.T) string {
	t.Helper()
	u, err := user.Current()
	if err != nil {
		t.Fatalf("get current user: %v", err)
	}
	p := filepath.Join(u.HomeDir, ".local", "share", "claude", "versions", "2.1.138")
	if _, err := os.Stat(p); err != nil {
		t.Skipf("claude binary not found at %s: %v", p, err)
	}
	return p
}

func loadBunInfo(t *testing.T) (data []byte, sec *bunSection, jsStart, footerStart uint64) {
	t.Helper()
	data, err := os.ReadFile(claudePath(t))
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	sec, err = findBunSection(data)
	if err != nil {
		t.Fatalf("findBunSection: %v", err)
	}
	jsStart = sec.fileOff + 8
	footerStart = sec.fileOff + sec.size - 52
	return
}

// TestFindBunSection verifies we can locate the __bun section and parse its header.
func TestFindBunSection(t *testing.T) {
	_, sec, _, _ := loadBunInfo(t)
	t.Logf("fileOff=%d size=%d sizeFieldOff=%d segFilesize=%d",
		sec.fileOff, sec.size, sec.sizeFieldOff, sec.segFilesize)
	t.Logf("null padding available: %d bytes", sec.segFilesize-sec.size)
	if sec.size == 0 {
		t.Fatal("section size is zero")
	}
	if sec.segFilesize < sec.size {
		t.Fatalf("segFilesize %d < section size %d", sec.segFilesize, sec.size)
	}
}

// TestBunFooter verifies the footer magic and dumps the footer fields.
func TestBunFooter(t *testing.T) {
	data, sec, _, _ := loadBunInfo(t)
	footerStart := sec.fileOff + sec.size - 52

	magic := string(data[footerStart+36 : footerStart+52])
	if magic != "\n---- Bun! ----\n" {
		t.Fatalf("bad footer magic: %q", magic)
	}

	fieldA := binary.LittleEndian.Uint64(data[footerStart+4 : footerStart+12])
	fieldB := binary.LittleEndian.Uint32(data[footerStart+12 : footerStart+16])
	fieldC := binary.LittleEndian.Uint64(data[footerStart+16 : footerStart+24])
	fieldD := binary.LittleEndian.Uint64(data[footerStart+24 : footerStart+32])
	bundleLen := binary.LittleEndian.Uint64(data[sec.fileOff : sec.fileOff+8])

	t.Logf("bundleLen=%d  A=%d  B=%d  C=%d  D=%d", bundleLen, fieldA, fieldB, fieldC, fieldD)
	t.Logf("moduleTableOff=%d  moduleTableSize=%d", fieldB, fieldD-uint64(fieldB))
}

// TestCliModule verifies we can parse the cli.js module entry from the module table.
func TestCliModule(t *testing.T) {
	data, sec, jsStart, _ := loadBunInfo(t)
	footerStart := sec.fileOff + sec.size - 52

	moduleTableOff := uint64(binary.LittleEndian.Uint32(data[footerStart+12 : footerStart+16]))
	mtAbs := jsStart + moduleTableOff

	cliCodeOff := uint64(binary.LittleEndian.Uint32(data[mtAbs+8 : mtAbs+12]))
	cliCodeSize := uint64(binary.LittleEndian.Uint32(data[mtAbs+12 : mtAbs+16]))
	cliPathOff := uint64(binary.LittleEndian.Uint32(data[mtAbs : mtAbs+4]))
	cliPathLen := uint64(binary.LittleEndian.Uint32(data[mtAbs+4 : mtAbs+8]))

	pathStr := string(data[jsStart+cliPathOff : jsStart+cliPathOff+cliPathLen])
	t.Logf("cli.js path: %q", pathStr)
	t.Logf("cli.js code: offset=%d size=%d end=%d", cliCodeOff, cliCodeSize, cliCodeOff+cliCodeSize)
	t.Logf("injection point (jsStart-relative): %d", cliCodeOff+cliCodeSize)

	if pathStr != "/$bunfs/root/src/entrypoints/cli.js" {
		t.Errorf("unexpected module path: %q", pathStr)
	}

	// Verify L48 is within cli.js code range
	cliCodeAbs := jsStart + cliCodeOff
	cliCode := data[cliCodeAbs : cliCodeAbs+cliCodeSize]
	l48Pattern := []byte("function L48(")
	if idx := indexBytes(cliCode, l48Pattern); idx < 0 {
		t.Error("L48 not found within cli.js code range")
	} else {
		t.Logf("L48 at offset %d within cli.js code", idx)
	}
}

// TestExtractNames runs name extraction against cli.js code.
func TestExtractNames(t *testing.T) {
	data, sec, jsStart, _ := loadBunInfo(t)
	footerStart := sec.fileOff + sec.size - 52
	moduleTableOff := uint64(binary.LittleEndian.Uint32(data[footerStart+12 : footerStart+16]))
	mtAbs := jsStart + moduleTableOff

	cliCodeOff := uint64(binary.LittleEndian.Uint32(data[mtAbs+8 : mtAbs+12]))
	cliCodeSize := uint64(binary.LittleEndian.Uint32(data[mtAbs+12 : mtAbs+16]))
	cliCode := data[jsStart+cliCodeOff : jsStart+cliCodeOff+cliCodeSize]

	claimsVar, err := extractClaimsVar(cliCode)
	if err != nil {
		t.Fatalf("extractClaimsVar: %v", err)
	}
	t.Logf("claimsVar = %q", claimsVar)

	rlFunc, err := extractRlFunc(cliCode)
	if err != nil {
		t.Fatalf("extractRlFunc: %v", err)
	}
	t.Logf("rlFunc = %q", rlFunc)

	successHandler, err := extractSuccessHandler(cliCode, rlFunc)
	if err != nil {
		t.Fatalf("extractSuccessHandler: %v", err)
	}
	t.Logf("successHandler = %q", successHandler)

	showContext(t, cliCode, []byte(rateLimitAnchor), "rateLimitAnchor", 80)
	showContext(t, cliCode, []byte("function "+rlFunc+"("), "rlFunc declaration", 120)
	showContext(t, cliCode, []byte("function "+successHandler+"("), "successHandler declaration", 120)
}

// TestPatchJS verifies the injection is built and contains the expected marker.
func TestPatchJS(t *testing.T) {
	data, sec, jsStart, _ := loadBunInfo(t)
	footerStart := sec.fileOff + sec.size - 52
	moduleTableOff := uint64(binary.LittleEndian.Uint32(data[footerStart+12 : footerStart+16]))
	mtAbs := jsStart + moduleTableOff

	cliCodeOff := uint64(binary.LittleEndian.Uint32(data[mtAbs+8 : mtAbs+12]))
	cliCodeSize := uint64(binary.LittleEndian.Uint32(data[mtAbs+12 : mtAbs+16]))
	cliCode := data[jsStart+cliCodeOff : jsStart+cliCodeOff+cliCodeSize]

	t.Logf("cli.js code: offset=%d size=%d", cliCodeOff, cliCodeSize)

	newCode, err := patchJS(cliCode)
	if err != nil {
		t.Fatalf("patchJS: %v", err)
	}

	added := len(newCode) - len(cliCode)
	t.Logf("patchJS: injected %d bytes at end of cli.js", added)

	marker := "[angelix]"
	found := false
	for i := len(cliCode); i < len(newCode); i++ {
		if i+len(marker) <= len(newCode) && string(newCode[i:i+len(marker)]) == marker {
			t.Logf("injection marker found at offset %d within newCode", i)
			found = true
			break
		}
	}
	if !found {
		t.Errorf("injection marker %q not found in appended injection", marker)
	}

	// Verify the injection is syntactically appended after the last JS byte
	t.Logf("last byte of original code: 0x%02x %q", cliCode[len(cliCode)-1], string(cliCode[len(cliCode)-5:]))
	t.Logf("first bytes of injection: %q", string(newCode[len(cliCode):len(cliCode)+20]))
}

// TestApplyPatches runs the full binary patch.
func TestApplyPatches(t *testing.T) {
	data, err := os.ReadFile(claudePath(t))
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}

	if err := applyPatches(data); err != nil {
		t.Fatalf("applyPatches: %v", err)
	}
	t.Logf("applyPatches succeeded, data len=%d", len(data))

	// Verify footer magic is still intact
	sec, _ := findBunSection(data)
	footerStart := sec.fileOff + sec.size - 52
	magic := string(data[footerStart+36 : footerStart+52])
	if magic != "\n---- Bun! ----\n" {
		t.Errorf("footer magic corrupted after patch: %q", magic)
	}
	t.Logf("footer magic OK at new position %d", footerStart)
}

// TestPatchedBinaryRuns writes the patched binary to disk and runs it.
// It verifies the [angelix] startup message appears on stderr.
func TestPatchedBinaryRuns(t *testing.T) {
	data, err := os.ReadFile(claudePath(t))
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}

	if err := applyPatches(data); err != nil {
		t.Fatalf("applyPatches: %v", err)
	}

	tmp, err := os.CreateTemp("", "claude-patched-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("write patched binary: %v", err)
	}
	tmp.Close()
	os.Chmod(tmp.Name(), 0755)

	// Re-sign so macOS can execute it
	if err := codesign(tmp.Name()); err != nil {
		t.Fatalf("codesign: %v", err)
	}

	cmd := exec.Command(tmp.Name(), "--version")
	out, _ := cmd.CombinedOutput()
	t.Logf("patched binary output:\n%s", out)

	if !contains(out, []byte("[angelix]")) {
		t.Errorf("[angelix] startup marker not found in output — IIFE not running")
	}
}

func contains(haystack, needle []byte) bool {
	return indexBytes(haystack, needle) >= 0
}

func showContext(t *testing.T, js []byte, needle []byte, label string, window int) {
	t.Helper()
	idx := 0
	for i := 0; i <= len(js)-len(needle); i++ {
		if string(js[i:i+len(needle)]) == string(needle) {
			idx = i
			break
		}
	}
	start := max(0, idx-window)
	end := min(len(js), idx+len(needle)+window)
	t.Logf("%s context:\n%s", label, fmt.Sprintf("%q", string(js[start:end])))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
