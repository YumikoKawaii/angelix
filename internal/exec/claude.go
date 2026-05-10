package exec

import (
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"syscall"

	"github.com/YumikoKawaii/angelix/internal/patch"
)

// RunClaude replaces the current process with the patched claude binary,
// inheriting stdin/stdout/stderr. extraEnv entries take precedence over
// the current environment.
func RunClaude(args []string, extraEnv []string) error {
	claudePath, err := resolveClaude()
	if err != nil {
		return err
	}

	patched, err := patch.EnsurePatch(claudePath)
	if err != nil {
		// Non-fatal: fall back to unpatched binary so the user isn't blocked.
		fmt.Fprintf(os.Stderr, "angelix: patch skipped (%v), running unpatched claude\n", err)
		patched = claudePath
	}

	argv := append([]string{patched}, args...)
	env := append(os.Environ(), extraEnv...)
	return syscall.Exec(patched, argv, env)
}

// resolveClaude returns the path to the claude binary to use.
// It prefers the angelix-managed copy at ~/.angelix/bin/claude over whatever
// is on PATH, so the version bundled by angelix is always used when present.
func resolveClaude() (string, error) {
	home, err := os.UserHomeDir()
	if err == nil {
		managed := filepath.Join(home, ".angelix", "bin", "claude")
		if _, err := os.Stat(managed); err == nil {
			return managed, nil
		}
	}

	p, err := osexec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude not found — run 'angelix setup' or install Claude Code manually")
	}
	return p, nil
}
