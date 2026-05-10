package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Setup downloads the claude binary from npm into ~/.angelix/bin/claude.
// It uses the latest published version unless version is non-empty.
func Setup(version string) error {
	suffix, err := platformSuffix()
	if err != nil {
		return err
	}

	if version == "" {
		version, err = latestClaudeVersion()
		if err != nil {
			return err
		}
	}

	pkgName := "@anthropic-ai/claude-code-" + suffix
	// npm registry tarball URL: the scope is stripped from the filename part.
	filename := strings.TrimPrefix(pkgName, "@anthropic-ai/") + "-" + version + ".tgz"
	url := "https://registry.npmjs.org/" + pkgName + "/-/" + filename

	fmt.Printf("Downloading claude %s for %s...\n", version, suffix)

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("npm registry returned %d for %s", resp.StatusCode, url)
	}

	binary, err := extractClaudeFromTarball(resp.Body)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	destDir := filepath.Join(home, ".angelix", "bin")
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	dest := filepath.Join(destDir, "claude")

	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, binary, 0755); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("install binary: %w", err)
	}

	fmt.Printf("claude %s installed to %s\n", version, dest)
	return nil
}

func platformSuffix() (string, error) {
	os_ := runtime.GOOS
	arch := runtime.GOARCH
	switch os_ {
	case "darwin":
		switch arch {
		case "arm64":
			return "darwin-arm64", nil
		case "amd64":
			return "darwin-x64", nil
		}
	case "linux":
		musl := isMusl()
		switch arch {
		case "arm64":
			if musl {
				return "linux-arm64-musl", nil
			}
			return "linux-arm64", nil
		case "amd64":
			if musl {
				return "linux-x64-musl", nil
			}
			return "linux-x64", nil
		}
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", os_, arch)
}

func isMusl() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "musl")
}

func latestClaudeVersion() (string, error) {
	resp, err := http.Get("https://registry.npmjs.org/@anthropic-ai/claude-code/latest") //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("fetch latest version: %w", err)
	}
	defer resp.Body.Close()
	var meta struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", fmt.Errorf("parse npm metadata: %w", err)
	}
	if meta.Version == "" {
		return "", fmt.Errorf("empty version in npm metadata")
	}
	return meta.Version, nil
}

func extractClaudeFromTarball(r io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("decompress tarball: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tarball: %w", err)
		}
		if hdr.Name == "package/claude" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("claude binary not found in tarball")
}
