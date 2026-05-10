package api

import (
	"net/http"

	"github.com/YumikoKawaii/angelix/pkg/apitypes"
)

func (s *Server) handleGetCredentials(w http.ResponseWriter, r *http.Request) {
	memberID := memberIDFromCtx(r.Context())

	member, err := s.store.GetMemberByID(memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load member")
		return
	}

	writeJSON(w, http.StatusOK, apitypes.CredentialResponse{
		APIKey: member.APIKey,
	})
}
