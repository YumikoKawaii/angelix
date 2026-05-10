package store

import "github.com/YumikoKawaii/angelix/server/model"

// Store handles member management only. Span/metrics storage lives in metrics.Store.
type Store interface {
	CreateMember(m *model.Member) error
	GetMemberByToken(token string) (*model.Member, error)
	GetMemberByID(id string) (*model.Member, error)
	ListMembers() ([]*model.Member, error)
	DeleteMember(id string) error
}
