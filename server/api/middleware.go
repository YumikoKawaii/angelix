package api

import (
	"net/http"
	"strings"
)

type contextKey string

const memberIDKey contextKey = "member_id"

// memberAuth validates the Bearer token and injects the member into the request context.
func (s *Server) memberAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		member, err := s.store.GetMemberByToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := r.Context()
		ctx = contextWithValue(ctx, memberIDKey, member.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// adminAuth checks the X-Admin-Token header against the configured admin token.
func (s *Server) adminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Admin-Token")
		if token == "" || token != s.adminToken {
			writeError(w, http.StatusUnauthorized, "invalid admin token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) string {
	v := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(v, "Bearer "); ok {
		return after
	}
	return ""
}
