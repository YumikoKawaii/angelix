package apitypes

import "time"

// Credential endpoints

type CredentialResponse struct {
	APIKey string `json:"api_key"`
}

// Member management endpoints

type CreateMemberRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	APIKey string `json:"api_key"`
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
	Name      string  `json:"name"`
	Count     int     `json:"count"`
	AvgMs     float64 `json:"avg_ms"`
	ErrorRate float64 `json:"error_rate"`
}

type MetricsSummaryResponse struct {
	MemberID   string     `json:"member_id"`
	TotalSpans int        `json:"total_spans"`
	ErrorSpans int        `json:"error_spans"`
	TopTools   []ToolStat `json:"top_tools"`
}

// Error

type ErrorResponse struct {
	Error string `json:"error"`
}
