package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/dhamariT/dispatch/internal/auth"
	"github.com/dhamariT/dispatch/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const inviteLifetime = 7 * 24 * time.Hour

type createInviteReq struct {
	Role  string `json:"role"`
	Email string `json:"email"`
}

type createInviteResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	ExpiresAt      time.Time `json:"expires_at"`
	// Token is returned to the caller exactly once. Only a hash is
	// stored server-side, so there is no way to retrieve it again.
	Token string `json:"token"`
}

func hashInviteToken(raw string) []byte {
	h := sha256.Sum256([]byte(raw))
	return h[:]
}

func (s *server) createInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		actor, _ := auth.UserFromCtx(r.Context())

		var req createInviteReq
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !auth.ValidRole(req.Role) {
			http.Error(w, "role must be admin, operator, or viewer", http.StatusBadRequest)
			return
		}

		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			s.log.Error("read random", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		token := base64.RawURLEncoding.EncodeToString(raw)

		inv, err := s.db.CreateInvite(
			r.Context(),
			org.ID,
			req.Email,
			req.Role,
			hashInviteToken(token),
			actor.ID,
			time.Now().Add(inviteLifetime),
		)
		if err != nil {
			s.log.Error("create invite", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, createInviteResponse{
			ID:             inv.ID,
			OrganizationID: inv.OrganizationID,
			Email:          inv.Email,
			Role:           inv.Role,
			ExpiresAt:      inv.ExpiresAt,
			Token:          token,
		})
	}
}

func (s *server) listInvites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		invites, err := s.db.ListInvitesForOrganization(r.Context(), org.ID)
		if err != nil {
			s.log.Error("list invites", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, invites)
	}
}

func (s *server) deleteInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "invalid invite id", http.StatusBadRequest)
			return
		}
		if err := s.db.DeleteInvite(r.Context(), id, org.ID); err != nil {
			s.log.Error("delete invite", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type inviteLookupResponse struct {
	OrganizationSlug string `json:"organization_slug"`
	OrganizationName string `json:"organization_name"`
	Role             string `json:"role"`
	Email            string `json:"email"`
}

func (s *server) getInviteByToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		inv, err := s.db.GetLiveInviteByHashedToken(r.Context(), hashInviteToken(token))
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, "invite not found or expired", http.StatusNotFound)
				return
			}
			s.log.Error("get invite", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		org, err := s.db.GetOrganizationByID(r.Context(), inv.OrganizationID)
		if err != nil {
			s.log.Error("get org for invite", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, inviteLookupResponse{
			OrganizationSlug: org.Slug,
			OrganizationName: org.Name,
			Role:             inv.Role,
			Email:            inv.Email,
		})
	}
}

func (s *server) acceptInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.UserFromCtx(r.Context())
		token := chi.URLParam(r, "token")

		inv, err := s.db.GetLiveInviteByHashedToken(r.Context(), hashInviteToken(token))
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, "invite not found or expired", http.StatusNotFound)
				return
			}
			s.log.Error("get invite", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if _, err := s.db.AddOrganizationMember(r.Context(), inv.OrganizationID, user.ID, inv.Role); err != nil {
			s.log.Error("accept invite: add member", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := s.db.MarkInviteAccepted(r.Context(), inv.ID, user.ID); err != nil {
			// Non-fatal — the membership row exists so the user is
			// effectively in. Log and continue.
			s.log.Warn("mark invite accepted", "err", err)
		}

		org, err := s.db.GetOrganizationByID(r.Context(), inv.OrganizationID)
		if err != nil {
			s.log.Error("get org after accept", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, orgResponse{Organization: org, Role: inv.Role})
	}
}
