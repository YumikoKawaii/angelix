package metrics

import (
	"sync"

	"github.com/YumikoKawaii/angelix/server/model"
)

// Memory is an in-memory MetricsStore used in tests.
type Memory struct {
	mu    sync.RWMutex
	spans map[string][]*model.Span // memberID → spans
}

func NewMemory() *Memory {
	return &Memory{spans: make(map[string][]*model.Span)}
}

func (m *Memory) RecordSpans(memberID string, spans []*model.Span) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spans[memberID] = append(m.spans[memberID], spans...)
	return nil
}

func (m *Memory) GetSpanSummary(memberID string) (*model.SpanSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return buildSummary(memberID, m.spans[memberID]), nil
}

func (m *Memory) Close() error { return nil }

func buildSummary(memberID string, spans []*model.Span) *model.SpanSummary {
	type agg struct {
		count               int
		errCount            int
		totalMs             int64
		inputTokens         int64
		outputTokens        int64
		cacheReadTokens     int64
		cacheCreationTokens int64
	}
	byName := make(map[string]*agg)

	summary := &model.SpanSummary{MemberID: memberID}
	for _, sp := range spans {
		summary.TotalSpans++
		summary.InputTokens += sp.InputTokens
		summary.OutputTokens += sp.OutputTokens
		summary.CacheReadTokens += sp.CacheReadTokens
		summary.CacheCreationTokens += sp.CacheCreationTokens
		if sp.IsError {
			summary.ErrorSpans++
		}
		a := byName[sp.Name]
		if a == nil {
			a = &agg{}
			byName[sp.Name] = a
		}
		a.count++
		a.totalMs += sp.DurationMs
		a.inputTokens += sp.InputTokens
		a.outputTokens += sp.OutputTokens
		a.cacheReadTokens += sp.CacheReadTokens
		a.cacheCreationTokens += sp.CacheCreationTokens
		if sp.IsError {
			a.errCount++
		}
	}
	for name, a := range byName {
		summary.TopTools = append(summary.TopTools, model.ToolStat{
			Name:                name,
			Count:               a.count,
			AvgMs:               float64(a.totalMs) / float64(a.count),
			ErrorRate:           float64(a.errCount) / float64(a.count),
			InputTokens:         a.inputTokens,
			OutputTokens:        a.outputTokens,
			CacheReadTokens:     a.cacheReadTokens,
			CacheCreationTokens: a.cacheCreationTokens,
		})
	}
	return summary
}
