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
					InputTokens:         attrInt(attrs, "gen_ai.usage.input_tokens"),
					OutputTokens:        attrInt(attrs, "gen_ai.usage.output_tokens"),
					CacheReadTokens:     attrInt(attrs, "gen_ai.usage.cache_read_input_tokens"),
					CacheCreationTokens: attrInt(attrs, "gen_ai.usage.cache_creation_input_tokens"),
					Model:               attrStr(attrs, "gen_ai.request.model"),
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
	n, _ := strconv.ParseInt(v.IntValue, 10, 64)
	return n
}

func attrStr(attrs map[string]AttributeValue, key string) string {
	return attrs[key].StringValue
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
