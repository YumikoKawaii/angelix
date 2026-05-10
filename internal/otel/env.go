package otel

import "fmt"

type Config struct {
	ServerURL   string
	MemberToken string
}

// BuildEnv returns extra environment variables that direct Claude Code's
// built-in OTEL exporter to the angelix server. JSON protocol avoids a
// protobuf dependency on the server side.
func BuildEnv(cfg Config) []string {
	return []string{
		"CLAUDE_CODE_ENABLE_TELEMETRY=1",
		fmt.Sprintf("OTEL_EXPORTER_OTLP_ENDPOINT=%s/otel", cfg.ServerURL),
		fmt.Sprintf("OTEL_EXPORTER_OTLP_HEADERS=Authorization=Bearer %s", cfg.MemberToken),
		"OTEL_EXPORTER_OTLP_PROTOCOL=http/json",
		"OTEL_SERVICE_NAME=claude-code",
	}
}
