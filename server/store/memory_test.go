package store_test

import (
	"testing"
	"time"

	"github.com/YumikoKawaii/angelix/server/model"
	"github.com/YumikoKawaii/angelix/server/store"
)

func newMember(id, token string) *model.Member {
	return &model.Member{
		ID: id, Name: "Test", Email: id + "@test.com",
		Token: token, APIKey: "sk-test", CreatedAt: time.Now(),
	}
}

func TestMemory_CreateAndGet(t *testing.T) {
	s := store.NewMemory()
	m := newMember("id1", "tok1")

	if err := s.CreateMember(m); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetMemberByToken("tok1")
	if err != nil || got.ID != "id1" {
		t.Fatalf("expected id1, got %v err %v", got, err)
	}
	got, err = s.GetMemberByID("id1")
	if err != nil || got.Token != "tok1" {
		t.Fatalf("expected tok1, got %v err %v", got, err)
	}
}

func TestMemory_DuplicateToken(t *testing.T) {
	s := store.NewMemory()
	_ = s.CreateMember(newMember("id1", "sametoken"))
	if err := s.CreateMember(newMember("id2", "sametoken")); err == nil {
		t.Fatal("expected error on duplicate token")
	}
}

func TestMemory_Delete(t *testing.T) {
	s := store.NewMemory()
	_ = s.CreateMember(newMember("id1", "tok1"))
	if err := s.DeleteMember("id1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetMemberByID("id1"); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestMemory_ListMembers(t *testing.T) {
	s := store.NewMemory()
	_ = s.CreateMember(newMember("a", "ta"))
	_ = s.CreateMember(newMember("b", "tb"))
	members, err := s.ListMembers()
	if err != nil || len(members) != 2 {
		t.Fatalf("want 2 members, got %d err %v", len(members), err)
	}
}
