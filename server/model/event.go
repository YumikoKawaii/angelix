package model

// EventType tags non-trace OTLP signals for logging purposes.
type EventType string

const (
	EventMetrics EventType = "metrics"
	EventLogs    EventType = "logs"
)
