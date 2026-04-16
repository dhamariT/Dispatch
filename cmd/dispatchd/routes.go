package main

import (
	"net/http"

	"github.com/dhamariT/dispatch/internal/auth"
	"github.com/go-chi/chi/v5"
)

func (s *server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(s.cors)
	r.Use(s.auth.SessionMiddleware)

	// Public OAuth endpoints. SessionMiddleware still runs over them so
	// a pre-existing cookie populates the user context if present, but
	// neither is gated on RequireUser.
	r.Get("/api/auth/github/login", s.auth.LoginHandler())
	r.Get("/api/auth/github/callback", s.auth.CallbackHandler())

	// Metric ingest is intentionally unauthenticated for now. Dispatch
	// has no device-agent credential model yet. When it does, this
	// endpoint should require a per-device API key and assert that
	// the sample's device_id matches the credential's device_id.
	r.Post("/api/metrics", s.pushSample())

	r.Group(func(r chi.Router) {
		r.Use(s.auth.RequireUser)

		r.Get("/api/auth/me", s.me())
		r.Post("/api/auth/logout", s.logout())

		r.Get("/api/orgs", s.listMyOrgs())
		r.Post("/api/orgs", s.createOrg())

		// Invite acceptance requires a logged-in user; the raw token is
		// the authorization to join, not to authenticate.
		r.Get("/api/invites/{token}", s.getInviteByToken())
		r.Post("/api/invites/{token}/accept", s.acceptInvite())

		r.Route("/api/orgs/{orgSlug}", func(r chi.Router) {
			r.Use(s.auth.ExtractOrg)

			r.Get("/", s.getOrg())
			r.Get("/members", s.listMembers())
			r.Get("/invites", s.listInvites())

			r.Get("/experiments", s.listExperiments())
			r.Get("/experiments/{id}", s.getExperiment())
			r.Get("/simulation/scenarios", s.listScenarios())

			// Operator+ write endpoints.
			r.Group(func(r chi.Router) {
				r.Use(s.auth.RequireRole(auth.RoleAdmin, auth.RoleOperator))
				r.Post("/experiments", s.createExperiment())
				r.Post("/experiments/{id}/analyze", s.analyzeExperiment())
				r.Post("/experiments/{id}/promote", s.promoteExperiment())
				r.Post("/experiments/{id}/hold", s.holdExperiment())
				r.Post("/simulation/run", s.runScenario())
			})

			// Admin-only endpoints.
			r.Group(func(r chi.Router) {
				r.Use(s.auth.RequireRole(auth.RoleAdmin))
				r.Post("/invites", s.createInvite())
				r.Delete("/invites/{id}", s.deleteInvite())
				r.Delete("/members/{userID}", s.removeMember())
			})
		})
	})

	return r
}

// cors echoes the configured frontend origin and allows credentials so
// the session cookie is sent on cross-origin API calls from the Next.js
// dev server. We never use Access-Control-Allow-Origin: * here because
// that's incompatible with Allow-Credentials.
func (s *server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == s.corsOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
