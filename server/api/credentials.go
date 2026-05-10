package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/YumikoKawaii/angelix/pkg/apitypes"
	"github.com/YumikoKawaii/angelix/server/model"
)

// handleGetCredentials is called by the angelix wrapper (member auth).
// Returns the first assigned credential's API key, or falls back to the
// member's direct api_key for backward compatibility.
func (s *Server) handleGetCredentials(w http.ResponseWriter, r *http.Request) {
	memberID := memberIDFromCtx(r.Context())

	creds, err := s.store.ListMemberCredentials(memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load credentials")
		return
	}

	if len(creds) > 0 {
		writeJSON(w, http.StatusOK, apitypes.CredentialResponse{APIKey: creds[0].APIKey})
		return
	}

	// Fallback: member's direct api_key (legacy)
	member, err := s.store.GetMemberByID(memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load member")
		return
	}
	writeJSON(w, http.StatusOK, apitypes.CredentialResponse{APIKey: member.APIKey})
}

// handleListCredentials lists all credentials in the catalog (admin).
func (s *Server) handleListCredentials(w http.ResponseWriter, r *http.Request) {
	creds, err := s.store.ListCredentials()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list credentials")
		return
	}
	resp := apitypes.ListCredentialsResponse{Credentials: make([]apitypes.CredentialItem, 0, len(creds))}
	for _, c := range creds {
		resp.Credentials = append(resp.Credentials, apitypes.CredentialItem{
			ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleCreateCredential creates a new credential in the catalog (admin).
func (s *Server) handleCreateCredential(w http.ResponseWriter, r *http.Request) {
	var req apitypes.CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.APIKey == "" {
		writeError(w, http.StatusBadRequest, "name and api_key are required")
		return
	}
	c := &model.Credential{
		ID:        newToken(16),
		Name:      req.Name,
		APIKey:    req.APIKey,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.CreateCredential(c); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, apitypes.CredentialItem{
		ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt,
	})
}

// handleDeleteCredential removes a credential and all its assignments (admin).
func (s *Server) handleDeleteCredential(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteCredential(r.PathValue("id")); err != nil {
		writeError(w, http.StatusNotFound, "credential not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleListMemberCredentials lists credentials assigned to a member (admin).
func (s *Server) handleListMemberCredentials(w http.ResponseWriter, r *http.Request) {
	creds, err := s.store.ListMemberCredentials(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list member credentials")
		return
	}
	resp := apitypes.ListCredentialsResponse{Credentials: make([]apitypes.CredentialItem, 0, len(creds))}
	for _, c := range creds {
		resp.Credentials = append(resp.Credentials, apitypes.CredentialItem{
			ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleAssignCredential assigns a credential to a member (admin).
func (s *Server) handleAssignCredential(w http.ResponseWriter, r *http.Request) {
	var req apitypes.AssignCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CredentialID == "" {
		writeError(w, http.StatusBadRequest, "credential_id is required")
		return
	}
	if err := s.store.AssignCredential(r.PathValue("id"), req.CredentialID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUnassignCredential removes a credential assignment from a member (admin).
func (s *Server) handleUnassignCredential(w http.ResponseWriter, r *http.Request) {
	if err := s.store.UnassignCredential(r.PathValue("id"), r.PathValue("cred_id")); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
