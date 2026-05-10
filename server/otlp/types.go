package otlp

import "encoding/json"

// Minimal OTLP JSON types for traces.
// Numeric timestamps are sent as decimal strings per the OTLP JSON spec.

// FlexInt unmarshals an OTLP intValue that may arrive as either a quoted
// decimal string ("123") or a bare JSON number (123).
type FlexInt string

func (f *FlexInt) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexInt(s)
		return nil
	}
	// bare number — keep raw digits as a string
	*f = FlexInt(data)
	return nil
}

type TracesPayload struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans"`
}

type ResourceSpan struct {
	ScopeSpans []ScopeSpan `json:"scopeSpans"`
}

type ScopeSpan struct {
	Spans []Span `json:"spans"`
}

type Span struct {
	TraceID           string      `json:"traceId"`
	SpanID            string      `json:"spanId"`
	Name              string      `json:"name"`
	StartTimeUnixNano string      `json:"startTimeUnixNano"`
	EndTimeUnixNano   string      `json:"endTimeUnixNano"`
	Status            SpanStatus  `json:"status"`
	Attributes        []Attribute `json:"attributes"`
}

type SpanStatus struct {
	// 0 = unset, 1 = ok, 2 = error
	Code int `json:"code"`
}

type Attribute struct {
	Key   string         `json:"key"`
	Value AttributeValue `json:"value"`
}

type AttributeValue struct {
	StringValue string  `json:"stringValue,omitempty"`
	IntValue    FlexInt `json:"intValue,omitempty"`
	DoubleValue float64 `json:"doubleValue,omitempty"`
}

// MetricsPayload is a minimal wrapper; full parsing deferred.
type MetricsPayload struct {
	ResourceMetrics []ResourceMetric `json:"resourceMetrics"`
}

type ResourceMetric struct {
	ScopeMetrics []ScopeMetric `json:"scopeMetrics"`
}

type ScopeMetric struct {
	Metrics []Metric `json:"metrics"`
}

type Metric struct {
	Name string `json:"name"`
}
