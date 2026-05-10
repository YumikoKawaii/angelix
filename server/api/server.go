package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/YumikoKawaii/angelix/server/metrics"
	"github.com/YumikoKawaii/angelix/server/store"
)

type Server struct {
	store      store.Store    // member management (SQLite)
	metrics    metrics.Store  // span storage (ClickHouse / DuckDB)
	adminToken string
	mux        *http.ServeMux
}

func NewServer(st store.Store, ms metrics.Store, adminToken string) *Server {
	s := &Server{
		store:      st,
		metrics:    ms,
		adminToken: adminToken,
		mux:        http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Member-authenticated
	s.mux.Handle("GET /api/v1/credentials", s.memberAuth(http.HandlerFunc(s.handleGetCredentials)))
	s.mux.Handle("GET /api/v1/metrics", s.memberAuth(http.HandlerFunc(s.handleGetMetrics)))

	// Admin-authenticated
	s.mux.Handle("POST /api/v1/members", s.adminAuth(http.HandlerFunc(s.handleCreateMember)))
	s.mux.Handle("GET /api/v1/members", s.adminAuth(http.HandlerFunc(s.handleListMembers)))
	s.mux.Handle("DELETE /api/v1/members/{id}", s.adminAuth(http.HandlerFunc(s.handleDeleteMember)))
	s.mux.Handle("GET /api/v1/members/{id}/metrics", s.adminAuth(http.HandlerFunc(s.handleAdminGetMetrics)))

	// OTLP receiver — member-authenticated via Bearer token
	s.mux.Handle("POST /otel/v1/traces", s.memberAuth(http.HandlerFunc(s.handleOTELTraces)))
	s.mux.Handle("POST /otel/v1/metrics", s.memberAuth(http.HandlerFunc(s.handleOTELMetrics)))
	s.mux.Handle("POST /otel/v1/logs", s.memberAuth(http.HandlerFunc(s.handleOTELLogs)))

	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("request", "method", r.Method, "path", r.URL.Path)
	s.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func contextWithValue(ctx context.Context, key, val any) context.Context {
	return context.WithValue(ctx, key, val)
}

func memberIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(memberIDKey).(string)
	return v
}
