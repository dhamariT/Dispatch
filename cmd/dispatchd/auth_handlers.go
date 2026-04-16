package main

import (
	"net/http"

	"github.com/dhamariT/dispatch/internal/auth"
)

type meResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (s *server) me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.UserFromCtx(r.Context())
		writeJSON(w, http.StatusOK, meResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Login:     user.Login,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		})
	}
}

func (s *server) logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Best-effort delete of the row. Even if this fails, we clear
		// the cookie so the client stops presenting it. Session
		// cleanup also runs via DeleteExpiredSessions.
		if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
			if sess, err := s.auth.ValidateSession(r.Context(), cookie.Value); err == nil {
				_ = s.db.DeleteSession(r.Context(), sess.ID)
			}
		}
		s.auth.ClearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}
