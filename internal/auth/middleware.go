package auth

import (
	"errors"
	"net/http"
	"slices"

	"github.com/dhamariT/dispatch/internal/db"
	"github.com/go-chi/chi/v5"
)

// SessionMiddleware parses the session cookie if present and attaches the
// user to the request context. It does NOT reject unauthenticated requests
// — that's RequireUser's job — so that public endpoints like the OAuth
// callback can share this middleware chain.
func (s *Service) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		sess, err := s.ValidateSession(ctx, cookie.Value)
		if err != nil {
			// Bad or expired cookie: clear it so the browser stops
			// sending it and the frontend falls through to login.
			s.ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}
		user, err := s.db.GetUserByID(ctx, sess.UserID)
		if err != nil {
			s.ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}
		// Best-effort bump of last_used_at; failing this should not
		// fail the request.
		_ = s.db.TouchSession(ctx, sess.ID)

		ctx = WithUser(ctx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireUser rejects the request with 401 if no user is on the context.
// Always used *after* SessionMiddleware in the chain.
func (s *Service) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFromCtx(r.Context()); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ExtractOrg resolves {orgSlug} from the URL, checks membership, and
// attaches both the org and the membership row to the context.
//
// Non-members get 404 (not 403) so an outsider can't probe for the
// existence of orgs they don't belong to.
func (s *Service) ExtractOrg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := UserFromCtx(ctx)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		slug := chi.URLParam(r, "orgSlug")
		org, err := s.db.GetOrganizationBySlug(ctx, slug)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		membership, err := s.db.GetOrganizationMembership(ctx, org.ID, user.ID)
		if err != nil {
			// Non-member indistinguishable from non-existent org.
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		ctx = WithOrg(ctx, org)
		ctx = WithMembership(ctx, membership)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole rejects the request with 403 unless the current membership
// role is in the allowed set. Used on write endpoints (promote, hold,
// invite) to enforce the capability rules from roles.go.
func (s *Service) RequireRole(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m, ok := MembershipFromCtx(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !slices.Contains(allowed, m.Role) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
