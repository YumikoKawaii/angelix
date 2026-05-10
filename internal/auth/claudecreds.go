package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type claudeCredFile struct {
	ClaudeAiOauth claudeOAuth `json:"claudeAiOauth"`
}

type claudeOAuth struct {
	AccessToken string `json:"accessToken"`
}

// WriteClaudeCredentials writes the OAuth access token to ~/.claude/credentials.json,
// which is where Claude Code looks for authentication credentials.
func WriteClaudeCredentials(accessToken string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(claudeCredFile{
		ClaudeAiOauth: claudeOAuth{AccessToken: accessToken},
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600)
}
