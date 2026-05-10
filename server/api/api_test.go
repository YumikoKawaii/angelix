package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/YumikoKawaii/angelix/pkg/apitypes"
	"github.com/YumikoKawaii/angelix/server/api"
	"github.com/YumikoKawaii/angelix/server/model"
	"github.com/YumikoKawaii/angelix/server/store"
)

const adminTok = "admin-secret"

func newServer() *api.Server {
	return api.NewServer(store.NewMemory(), adminTok)
}

func adminReq(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("X-Admin-Token", adminTok)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func createMember(t *testing.T, srv *api.Server, name, email, apiKey string) apitypes.MemberResponse {
	t.Helper()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, adminReq("POST", "/api/v1/members", apitypes.CreateMemberRequest{
		Name: name, Email: email, APIKey: apiKey,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create member: want 201, got %d: %s", w.Code, w.Body)
	}
	var m apitypes.MemberResponse
	_ = json.NewDecoder(w.Body).Decode(&m)
	return m
}

func TestCreateMember(t *testing.T) {
	srv := newServer()
	m := createMember(t, srv, "Alice", "alice@test.com", "sk-test")
	if m.Token == "" || m.ID == "" {
		t.Fatal("expected non-empty token and id")
	}
}

func TestCreateMember_MissingFields(t *testing.T) {
	srv := newServer()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, adminReq("POST", "/api/v1/members", apitypes.CreateMemberRequest{Name: "Alice"}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestGetCredentials(t *testing.T) {
	srv := newServer()
	m := createMember(t, srv, "Bob", "bob@test.com", "sk-bob")

	req := httptest.NewRequest("GET", "/api/v1/credentials", nil)
	req.Header.Set("Authorization", "Bearer "+m.Token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body)
	}
	var creds apitypes.CredentialResponse
	_ = json.NewDecoder(w.Body).Decode(&creds)
	if creds.APIKey != "sk-bob" {
		t.Errorf("want sk-bob, got %s", creds.APIKey)
	}
}

func TestGetCredentials_InvalidToken(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("GET", "/api/v1/credentials", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

func TestListMembers_AdminOnly(t *testing.T) {
	srv := newServer()

	// no token → 401
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/members", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}

	// admin token → 200
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, adminReq("GET", "/api/v1/members", nil))
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

func TestDeleteMember(t *testing.T) {
	srv := newServer()
	m := createMember(t, srv, "Eve", "eve@test.com", "sk-eve")

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, adminReq("DELETE", "/api/v1/members/"+m.ID, nil))
	if w.Code != http.StatusNoContent {
		t.Errorf("want 204, got %d", w.Code)
	}

	// credentials should now be gone
	req := httptest.NewRequest("GET", "/api/v1/credentials", nil)
	req.Header.Set("Authorization", "Bearer "+m.Token)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401 after delete, got %d", w.Code)
	}
}

func TestOTELTraces(t *testing.T) {
	srv := newServer()
	m := createMember(t, srv, "Carol", "carol@test.com", "sk-carol")

	now := time.Now()
	payload := buildTracesPayload([]traceSpan{
		{name: "tool.bash", start: now, end: now.Add(100 * time.Millisecond), isError: false},
		{name: "tool.bash", start: now, end: now.Add(200 * time.Millisecond), isError: true},
		{name: "tool.read", start: now, end: now.Add(50 * time.Millisecond), isError: false},
	})

	req := httptest.NewRequest("POST", "/otel/v1/traces", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+m.Token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body)
	}

	// Check metrics summary
	req = httptest.NewRequest("GET", "/api/v1/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+m.Token)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	var summary apitypes.MetricsSummaryResponse
	_ = json.NewDecoder(w.Body).Decode(&summary)
	if summary.TotalSpans != 3 {
		t.Errorf("want 3 spans, got %d", summary.TotalSpans)
	}
	if summary.ErrorSpans != 1 {
		t.Errorf("want 1 error span, got %d", summary.ErrorSpans)
	}
}

// helpers for building OTLP JSON test payloads

type traceSpan struct {
	name    string
	start   time.Time
	end     time.Time
	isError bool
}

func buildTracesPayload(spans []traceSpan) []byte {
	type status struct {
		Code int `json:"code"`
	}
	type span struct {
		TraceID           string `json:"traceId"`
		SpanID            string `json:"spanId"`
		Name              string `json:"name"`
		StartTimeUnixNano string `json:"startTimeUnixNano"`
		EndTimeUnixNano   string `json:"endTimeUnixNano"`
		Status            status `json:"status"`
	}
	type scopeSpan struct {
		Spans []span `json:"spans"`
	}
	type resourceSpan struct {
		ScopeSpans []scopeSpan `json:"scopeSpans"`
	}
	type payload struct {
		ResourceSpans []resourceSpan `json:"resourceSpans"`
	}

	var ss []span
	for i, s := range spans {
		code := 0
		if s.isError {
			code = 2
		}
		ss = append(ss, span{
			TraceID:           fmt.Sprintf("trace%02d", i),
			SpanID:            fmt.Sprintf("span%02d", i),
			Name:              s.name,
			StartTimeUnixNano: fmt.Sprintf("%d", s.start.UnixNano()),
			EndTimeUnixNano:   fmt.Sprintf("%d", s.end.UnixNano()),
			Status:            status{Code: code},
		})
	}

	b, _ := json.Marshal(payload{
		ResourceSpans: []resourceSpan{{ScopeSpans: []scopeSpan{{Spans: ss}}}},
	})
	return b
}

func TestHealth(t *testing.T) {
	srv := newServer()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// Ensure memory store satisfies Store interface at compile time.
var _ interface {
	CreateMember(*model.Member) error
} = (*store.Memory)(nil)
