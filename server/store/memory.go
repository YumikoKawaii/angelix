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
}

func NewMemory() *Memory {
	return &Memory{
		byToken: make(map[string]*model.Member),
		byID:    make(map[string]*model.Member),
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
	return nil
}
