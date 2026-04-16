package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dhamariT/dispatch/internal/db"
	"github.com/google/uuid"
)

const SessionCookieName = "dispatch_session"

var ErrInvalidToken = errors.New("invalid session token")

// A session token is "<base64url(sessionID)>.<base64url(secret)>".
// The id identifies the row; the secret is compared against the stored
// SHA-256 digest in constant time. Only the digest is ever persisted,
// so a compromised DB backup cannot be used to forge live cookies.

func encodeSessionToken(id uuid.UUID, secret []byte) string {
	return base64.RawURLEncoding.EncodeToString(id[:]) + "." +
		base64.RawURLEncoding.EncodeToString(secret)
}

func decodeSessionToken(token string) (uuid.UUID, []byte, error) {
	idPart, secPart, ok := strings.Cut(token, ".")
	if !ok {
		return uuid.Nil, nil, ErrInvalidToken
	}
	idBytes, err := base64.RawURLEncoding.DecodeString(idPart)
	if err != nil || len(idBytes) != 16 {
		return uuid.Nil, nil, ErrInvalidToken
	}
	secret, err := base64.RawURLEncoding.DecodeString(secPart)
	if err != nil || len(secret) == 0 {
		return uuid.Nil, nil, ErrInvalidToken
	}
	var id uuid.UUID
	copy(id[:], idBytes)
	return id, secret, nil
}

func hashSecret(secret []byte) []byte {
	h := sha256.Sum256(secret)
	return h[:]
}

func (s *Service) IssueSession(ctx context.Context, w http.ResponseWriter, userID uuid.UUID) (db.Session, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return db.Session{}, fmt.Errorf("read random: %w", err)
	}
	expires := time.Now().Add(s.cfg.sessionLifetime())
	session, err := s.db.CreateSession(ctx, userID, hashSecret(secret), expires)
	if err != nil {
		return db.Session{}, err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    encodeSessionToken(session.ID, secret),
		Path:     "/",
		Domain:   s.cfg.CookieDomain,
		Expires:  expires,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return session, nil
}

func (s *Service) ValidateSession(ctx context.Context, token string) (db.Session, error) {
	id, secret, err := decodeSessionToken(token)
	if err != nil {
		return db.Session{}, err
	}
	sess, err := s.db.GetSessionByID(ctx, id)
	if err != nil {
		return db.Session{}, err
	}
	if time.Now().After(sess.ExpiresAt) {
		// Best-effort cleanup; ignore failure since the session is
		// already invalid from the caller's perspective.
		_ = s.db.DeleteSession(ctx, sess.ID)
		return db.Session{}, ErrInvalidToken
	}
	if subtle.ConstantTimeCompare(sess.HashedSecret, hashSecret(secret)) != 1 {
		return db.Session{}, ErrInvalidToken
	}
	return sess, nil
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Domain:   s.cfg.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
