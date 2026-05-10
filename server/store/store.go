package store

import "github.com/YumikoKawaii/angelix/server/model"

type Store interface {
	// Member management
	CreateMember(m *model.Member) error
	GetMemberByToken(token string) (*model.Member, error)
	GetMemberByID(id string) (*model.Member, error)
	ListMembers() ([]*model.Member, error)
	DeleteMember(id string) error

	// Telemetry
	RecordSpans(memberID string, spans []*model.Span) error
	GetSpanSummary(memberID string) (*model.SpanSummary, error)
}
