package apitypes

import "time"

// Credential endpoints (wrapper-facing)

type CredentialResponse struct {
	APIKey string `json:"api_key"`
}

// Credential catalog (admin)

type CreateCredentialRequest struct {
	Name   string `json:"name"`
	APIKey string `json:"api_key"`
}

type CredentialItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ListCredentialsResponse struct {
	Credentials []CredentialItem `json:"credentials"`
}

type AssignCredentialRequest struct {
	CredentialID string `json:"credential_id"`
}

// Member management endpoints

type CreateMemberRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type MemberResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

type ListMembersResponse struct {
	Members []MemberResponse `json:"members"`
}

// Metrics

type ToolStat struct {
	Name                string  `json:"name"`
	Count               int     `json:"count"`
	AvgMs               float64 `json:"avg_ms"`
	ErrorRate           float64 `json:"error_rate"`
	InputTokens         int64   `json:"input_tokens"`
	OutputTokens        int64   `json:"output_tokens"`
	CacheReadTokens     int64   `json:"cache_read_tokens"`
	CacheCreationTokens int64   `json:"cache_creation_tokens"`
}

type MetricsSummaryResponse struct {
	MemberID            string     `json:"member_id"`
	TotalSpans          int        `json:"total_spans"`
	ErrorSpans          int        `json:"error_spans"`
	InputTokens         int64      `json:"input_tokens"`
	OutputTokens        int64      `json:"output_tokens"`
	CacheReadTokens     int64      `json:"cache_read_tokens"`
	CacheCreationTokens int64      `json:"cache_creation_tokens"`
	TopTools            []ToolStat `json:"top_tools"`
}

// Error

type ErrorResponse struct {
	Error string `json:"error"`
}
