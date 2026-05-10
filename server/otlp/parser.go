package otlp

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/YumikoKawaii/angelix/server/model"
)

// ParseTraces unmarshals an OTLP JSON traces payload and converts each span
// to a model.Span. Unknown fields and malformed timestamps are skipped.
func ParseTraces(memberID string, data []byte) ([]*model.Span, error) {
	var payload TracesPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var spans []*model.Span

	for _, rs := range payload.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			for _, s := range ss.Spans {
				start := nanoStringToTime(s.StartTimeUnixNano)
				end := nanoStringToTime(s.EndTimeUnixNano)
				if start.IsZero() || end.IsZero() {
					continue
				}
				attrs := indexAttrs(s.Attributes)
				spans = append(spans, &model.Span{
					MemberID:            memberID,
					TraceID:             s.TraceID,
					SpanID:              s.SpanID,
					Name:                s.Name,
					StartTime:           start,
					EndTime:             end,
					DurationMs:          end.Sub(start).Milliseconds(),
					IsError:             s.Status.Code == 2,
					RecordedAt:          now,
					InputTokens:         attrInt(attrs, "input_tokens"),
					OutputTokens:        attrInt(attrs, "output_tokens"),
					CacheReadTokens:     attrInt(attrs, "cache_read_tokens"),
					CacheCreationTokens: attrInt(attrs, "cache_creation_tokens"),
					Model:               attrStr(attrs, "model"),
					ToolName:            attrStr(attrs, "tool_name"),
					StopReason:          attrStr(attrs, "stop_reason"),
					UserPromptLength:    attrInt(attrs, "user_prompt_length"),
					FiveHourUtilization: attrDouble(attrs, "five_hour_utilization"),
					SevenDayUtilization: attrDouble(attrs, "seven_day_utilization"),
				})
			}
		}
	}
	return spans, nil
}

func indexAttrs(attrs []Attribute) map[string]AttributeValue {
	m := make(map[string]AttributeValue, len(attrs))
	for _, a := range attrs {
		m[a.Key] = a.Value
	}
	return m
}

func attrInt(attrs map[string]AttributeValue, key string) int64 {
	v, ok := attrs[key]
	if !ok || v.IntValue == "" {
		return 0
	}
	n, _ := strconv.ParseInt(string(v.IntValue), 10, 64)
	return n
}

func attrStr(attrs map[string]AttributeValue, key string) string {
	return attrs[key].StringValue
}

func attrDouble(attrs map[string]AttributeValue, key string) float64 {
	return attrs[key].DoubleValue
}

func nanoStringToTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	ns, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(0, ns).UTC()
}
