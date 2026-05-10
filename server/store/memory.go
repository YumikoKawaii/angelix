package store

import (
	"fmt"
	"sync"

	"github.com/YumikoKawaii/angelix/server/model"
)

type Memory struct {
	mu      sync.RWMutex
	byToken map[string]*model.Member
	byID    map[string]*model.Member
	spans   map[string][]*model.Span // memberID → spans
}

func NewMemory() *Memory {
	return &Memory{
		byToken: make(map[string]*model.Member),
		byID:    make(map[string]*model.Member),
		spans:   make(map[string][]*model.Span),
	}
}

func (s *Memory) CreateMember(m *model.Member) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byToken[m.Token]; exists {
		return fmt.Errorf("token already in use")
	}
	s.byToken[m.Token] = m
	s.byID[m.ID] = m
	return nil
}

func (s *Memory) GetMemberByToken(token string) (*model.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.byToken[token]
	if !ok {
		return nil, fmt.Errorf("member not found")
	}
	return m, nil
}

func (s *Memory) GetMemberByID(id string) (*model.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("member not found")
	}
	return m, nil
}

func (s *Memory) ListMembers() ([]*model.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Member, 0, len(s.byID))
	for _, m := range s.byID {
		out = append(out, m)
	}
	return out, nil
}

func (s *Memory) DeleteMember(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.byID[id]
	if !ok {
		return fmt.Errorf("member not found")
	}
	delete(s.byToken, m.Token)
	delete(s.byID, id)
	delete(s.spans, id)
	return nil
}

func (s *Memory) RecordSpans(memberID string, spans []*model.Span) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spans[memberID] = append(s.spans[memberID], spans...)
	return nil
}

func (s *Memory) GetSpanSummary(memberID string) (*model.SpanSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	spans := s.spans[memberID]
	return buildSummary(memberID, spans), nil
}

// buildSummary computes SpanSummary from a flat span slice.
// Shared by both memory and SQLite stores (SQLite fetches rows then calls this).
func buildSummary(memberID string, spans []*model.Span) *model.SpanSummary {
	type agg struct {
		count      int
		errCount   int
		totalMs    int64
	}
	byName := make(map[string]*agg)

	summary := &model.SpanSummary{MemberID: memberID}
	for _, sp := range spans {
		summary.TotalSpans++
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
		if sp.IsError {
			a.errCount++
		}
	}

	for name, a := range byName {
		summary.TopTools = append(summary.TopTools, model.ToolStat{
			Name:      name,
			Count:     a.count,
			AvgMs:     float64(a.totalMs) / float64(a.count),
			ErrorRate: float64(a.errCount) / float64(a.count),
		})
	}
	return summary
}
