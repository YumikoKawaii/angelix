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

func newToken(bytes int) string {
	b := make([]byte, bytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
