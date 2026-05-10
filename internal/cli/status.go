package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/YumikoKawaii/angelix/internal/auth"
	"github.com/YumikoKawaii/angelix/pkg/apitypes"
)

// Status checks connectivity and prints a summary.
func Status() error {
	cfg, err := auth.LoadConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Health check
	fmt.Printf("Server: %s\n", cfg.ServerURL)
	if err := checkHealth(client, cfg.ServerURL); err != nil {
		fmt.Printf("  health: FAIL (%v)\n", err)
		return err
	}
	fmt.Printf("  health: ok\n")

	// Credential check — also reveals member identity
	member, err := checkCredentials(client, cfg.ServerURL, cfg.MemberToken)
	if err != nil {
		fmt.Printf("  token:  FAIL (%v)\n", err)
		return err
	}
	fmt.Printf("  token:  ok\n")
	_ = member // API key is intentionally not printed

	// Metrics summary
	summary, err := fetchMetrics(client, cfg.ServerURL, cfg.MemberToken)
	if err == nil {
		fmt.Printf("\nMetrics\n")
		fmt.Printf("  total spans : %d\n", summary.TotalSpans)
		fmt.Printf("  error spans : %d\n", summary.ErrorSpans)
		if len(summary.TopTools) > 0 {
			fmt.Printf("  top tools   :\n")
			for i, t := range summary.TopTools {
				if i >= 5 {
					break
				}
				fmt.Printf("    %-30s  count=%-5d  avg=%.1fms  err=%.0f%%\n",
					t.Name, t.Count, t.AvgMs, t.ErrorRate*100)
			}
		}
	}

	return nil
}

func checkHealth(c *http.Client, serverURL string) error {
	resp, err := c.Get(serverURL + "/health")
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func checkCredentials(c *http.Client, serverURL, token string) (*apitypes.CredentialResponse, error) {
	req, _ := http.NewRequest(http.MethodGet, serverURL+"/api/v1/credentials", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var cr apitypes.CredentialResponse
	return &cr, json.NewDecoder(resp.Body).Decode(&cr)
}

func fetchMetrics(c *http.Client, serverURL, token string) (*apitypes.MetricsSummaryResponse, error) {
	req, _ := http.NewRequest(http.MethodGet, serverURL+"/api/v1/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var s apitypes.MetricsSummaryResponse
	return &s, json.NewDecoder(resp.Body).Decode(&s)
}
