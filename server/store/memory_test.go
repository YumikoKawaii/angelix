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

func TestMemory_DeleteMember(t *testing.T) {
	s := store.NewMemory()
	m := newMember("id1", "tok1")
	_ = s.CreateMember(m)
	_ = s.RecordSpans("id1", []*model.Span{{Name: "tool.bash", DurationMs: 100}})

	if err := s.DeleteMember("id1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetMemberByID("id1"); err == nil {
		t.Fatal("expected error after delete")
	}
	// spans should be gone too
	sum, _ := s.GetSpanSummary("id1")
	if sum.TotalSpans != 0 {
		t.Fatalf("expected 0 spans after delete, got %d", sum.TotalSpans)
	}
}

func TestMemory_SpanSummary(t *testing.T) {
	s := store.NewMemory()
	_ = s.CreateMember(newMember("m1", "t1"))

	spans := []*model.Span{
		{Name: "tool.bash", DurationMs: 100, IsError: false},
		{Name: "tool.bash", DurationMs: 200, IsError: true},
		{Name: "tool.read", DurationMs: 50, IsError: false},
	}
	_ = s.RecordSpans("m1", spans)

	sum, err := s.GetSpanSummary("m1")
	if err != nil {
		t.Fatal(err)
	}
	if sum.TotalSpans != 3 {
		t.Errorf("want 3 spans, got %d", sum.TotalSpans)
	}
	if sum.ErrorSpans != 1 {
		t.Errorf("want 1 error span, got %d", sum.ErrorSpans)
	}

	byName := make(map[string]model.ToolStat)
	for _, ts := range sum.TopTools {
		byName[ts.Name] = ts
	}
	bash := byName["tool.bash"]
	if bash.Count != 2 {
		t.Errorf("want bash count 2, got %d", bash.Count)
	}
	if bash.AvgMs != 150 {
		t.Errorf("want bash avg 150ms, got %.1f", bash.AvgMs)
	}
	if bash.ErrorRate != 0.5 {
		t.Errorf("want bash error rate 0.5, got %.2f", bash.ErrorRate)
	}
}
