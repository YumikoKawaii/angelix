package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/YumikoKawaii/angelix/internal/auth"
)

// Init interactively creates ~/.angelix/config.json.
func Init() error {
	fmt.Println("angelix init — set up your local config")
	fmt.Println()

	sc := bufio.NewScanner(os.Stdin)

	serverURL := prompt(sc, "Server URL (e.g. https://angelix.example.com): ")
	if serverURL == "" {
		return fmt.Errorf("server URL is required")
	}
	serverURL = strings.TrimRight(serverURL, "/")

	memberToken := prompt(sc, "Member token (provided by your admin): ")
	if memberToken == "" {
		return fmt.Errorf("member token is required")
	}

	fmt.Print("Verifying connection... ")
	if err := verify(serverURL, memberToken); err != nil {
		fmt.Println("failed")
		return fmt.Errorf("verification failed: %w", err)
	}
	fmt.Println("ok")

	cfg := auth.Config{
		ServerURL:   serverURL,
		MemberToken: memberToken,
	}
	if err := saveConfig(cfg); err != nil {
		return err
	}

	cfgPath := configPath()
	fmt.Printf("\nConfig saved to %s\n", cfgPath)
	fmt.Println("Run 'angelix' to start Claude Code.")
	return nil
}

func verify(serverURL, token string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, serverURL+"/api/v1/credentials", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid token")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

func saveConfig(cfg auth.Config) error {
	dir := filepath.Dir(configPath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(configPath(), data, 0600)
}

func configPath() string {
	if v := os.Getenv("ANGELIX_CONFIG"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".angelix", "config.json")
}

func prompt(sc *bufio.Scanner, label string) string {
	fmt.Print(label)
	sc.Scan()
	return strings.TrimSpace(sc.Text())
}
