package exec

import (
	"fmt"
	osexec "os/exec"
	"os"
	"syscall"
)

// RunClaude replaces the current process with claude, inheriting stdin/stdout/stderr.
// extraEnv entries are appended after the current environment so they take precedence.
func RunClaude(args []string, extraEnv []string) error {
	claudePath, err := osexec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude binary not found in PATH — is Claude Code installed?")
	}

	argv := append([]string{"claude"}, args...)
	env := append(os.Environ(), extraEnv...)

	// syscall.Exec replaces the process in-place: PID stays the same,
	// signals work naturally, and no zombie process is left behind.
	return syscall.Exec(claudePath, argv, env)
}
