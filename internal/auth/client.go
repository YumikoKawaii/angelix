package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/YumikoKawaii/angelix/pkg/apitypes"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func FetchCredentials(serverURL, token string) (*apitypes.CredentialResponse, error) {
	req, err := http.NewRequest(http.MethodGet, serverURL+"/api/v1/credentials", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach server at %s: %w", serverURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid member token")
	}
	if resp.StatusCode != http.StatusOK {
		var errResp apitypes.ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp.Error)
	}

	var creds apitypes.CredentialResponse
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, fmt.Errorf("malformed credentials response: %w", err)
	}
	return &creds, nil
}
