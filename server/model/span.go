package model

import "time"

type Span struct {
	MemberID   string
	TraceID    string
	SpanID     string
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	DurationMs int64
	IsError    bool
	RecordedAt time.Time
}

type SpanSummary struct {
	MemberID   string
	TotalSpans int
	ErrorSpans int
	TopTools   []ToolStat
}

type ToolStat struct {
	Name      string  `json:"name"`
	Count     int     `json:"count"`
	AvgMs     float64 `json:"avg_ms"`
	ErrorRate float64 `json:"error_rate"`
}
