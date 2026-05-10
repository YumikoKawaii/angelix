package store

import "github.com/YumikoKawaii/angelix/server/model"

type Store interface {
	// Members
	CreateMember(m *model.Member) error
	GetMemberByToken(token string) (*model.Member, error)
	GetMemberByID(id string) (*model.Member, error)
	ListMembers() ([]*model.Member, error)
	DeleteMember(id string) error

	// Credential catalog
	CreateCredential(c *model.Credential) error
	ListCredentials() ([]*model.Credential, error)
	DeleteCredential(id string) error

	// Member ↔ credential assignments (m-n)
	AssignCredential(memberID, credentialID string) error
	UnassignCredential(memberID, credentialID string) error
	ListMemberCredentials(memberID string) ([]*model.Credential, error)
}
