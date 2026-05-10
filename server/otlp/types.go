package otlp

// Minimal OTLP JSON types for traces.
// Numeric timestamps are sent as decimal strings per the OTLP JSON spec.

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
	StringValue string `json:"stringValue,omitempty"`
	IntValue    int64  `json:"intValue,omitempty"`
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
