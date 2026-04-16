package main

import (
	"errors"
	"net/http"

	"github.com/dhamariT/dispatch/internal/auth"
	"github.com/dhamariT/dispatch/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type createOrgReq struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (s *server) createOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.UserFromCtx(r.Context())

		var req createOrgReq
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !validSlug(req.Slug) {
			http.Error(w, "slug must be 3-40 chars, lowercase alphanumeric or hyphen", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		org, err := s.db.CreateOrganization(r.Context(), req.Slug, req.Name)
		if err != nil {
			// Unique violation on slug → 409.
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				http.Error(w, "slug already taken", http.StatusConflict)
				return
			}
			s.log.Error("create organization", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// The creator becomes the org's first admin. This is the only
		// path that grants admin without going through an invite.
		if _, err := s.db.AddOrganizationMember(r.Context(), org.ID, user.ID, auth.RoleAdmin); err != nil {
			s.log.Error("add creator membership", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, org)
	}
}

func (s *server) listMyOrgs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.UserFromCtx(r.Context())
		orgs, err := s.db.ListOrganizationsForUser(r.Context(), user.ID)
		if err != nil {
			s.log.Error("list orgs for user", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, orgs)
	}
}

type orgResponse struct {
	db.Organization
	Role string `json:"role"`
}

func (s *server) getOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		m, _ := auth.MembershipFromCtx(r.Context())
		writeJSON(w, http.StatusOK, orgResponse{Organization: org, Role: m.Role})
	}
}

func (s *server) listMembers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		members, err := s.db.ListOrganizationMembers(r.Context(), org.ID)
		if err != nil {
			s.log.Error("list members", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, members)
	}
}

func (s *server) removeMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		actor, _ := auth.UserFromCtx(r.Context())

		targetID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}
		// An admin cannot remove themselves. Use transfer-then-leave
		// once that feature exists; for now, just refuse so an org
		// can't be orphaned by accident.
		if targetID == actor.ID {
			http.Error(w, "cannot remove yourself", http.StatusConflict)
			return
		}
		if err := s.db.RemoveOrganizationMember(r.Context(), org.ID, targetID); err != nil {
			s.log.Error("remove member", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
