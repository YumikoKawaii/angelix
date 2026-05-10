package model

import "time"

type Span struct {
	MemberID            string
	TraceID             string
	SpanID              string
	Name                string
	StartTime           time.Time
	EndTime             time.Time
	DurationMs          int64
	IsError             bool
	RecordedAt          time.Time
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
	Model               string
}

type SpanSummary struct {
	MemberID            string
	TotalSpans          int
	ErrorSpans          int
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
	TopTools            []ToolStat
}

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
