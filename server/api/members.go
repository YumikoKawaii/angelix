package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/YumikoKawaii/angelix/pkg/apitypes"
	"github.com/YumikoKawaii/angelix/server/model"
)

func (s *Server) handleCreateMember(w http.ResponseWriter, r *http.Request) {
	var req apitypes.CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" || req.APIKey == "" {
		writeError(w, http.StatusBadRequest, "name, email and api_key are required")
		return
	}

	member := &model.Member{
		ID:        newToken(16),
		Name:      req.Name,
		Email:     req.Email,
		Token:     newToken(32),
		APIKey:    req.APIKey,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.CreateMember(member); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, apitypes.MemberResponse{
		ID:        member.ID,
		Name:      member.Name,
		Email:     member.Email,
		Token:     member.Token,
		CreatedAt: member.CreatedAt,
	})
}

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.store.ListMembers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	resp := apitypes.ListMembersResponse{Members: make([]apitypes.MemberResponse, 0, len(members))}
	for _, m := range members {
		resp.Members = append(resp.Members, apitypes.MemberResponse{
			ID: m.ID, Name: m.Name, Email: m.Email, Token: m.Token, CreatedAt: m.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleDeleteMember(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteMember(r.PathValue("id")); err != nil {
		writeError(w, http.StatusNotFound, "member not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAdminGetMetrics(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("id")
	summary, err := s.store.GetSpanSummary(memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load metrics")
		return
	}
	tools := make([]apitypes.ToolStat, 0, len(summary.TopTools))
	for _, t := range summary.TopTools {
		tools = append(tools, apitypes.ToolStat{
			Name: t.Name, Count: t.Count, AvgMs: t.AvgMs, ErrorRate: t.ErrorRate,
		})
	}
	writeJSON(w, http.StatusOK, apitypes.MetricsSummaryResponse{
		MemberID: summary.MemberID, TotalSpans: summary.TotalSpans,
		ErrorSpans: summary.ErrorSpans, TopTools: tools,
	})
}

func newToken(bytes int) string {
	b := make([]byte, bytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
