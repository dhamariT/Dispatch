package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dhamariT/dispatch/internal/db"
)

const (
	stateCookieName = "dispatch_oauth_state"
	githubProvider  = "github"
)

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (s *Service) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := make([]byte, 32)
		if _, err := rand.Read(state); err != nil {
			http.Error(w, "failed to generate state", http.StatusInternalServerError)
			return
		}
		stateStr := base64.RawURLEncoding.EncodeToString(state)
		// Scope the state cookie to /api/auth so it isn't sent on
		// unrelated API calls. The browser will present it to the
		// callback handler where we compare it constant-time.
		http.SetCookie(w, &http.Cookie{
			Name:     stateCookieName,
			Value:    stateStr,
			Path:     "/api/auth",
			Domain:   s.cfg.CookieDomain,
			Expires:  time.Now().Add(10 * time.Minute),
			HttpOnly: true,
			Secure:   s.cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, s.oauth.AuthCodeURL(stateStr), http.StatusFound)
	}
}

func (s *Service) CallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		stateCookie, err := r.Cookie(stateCookieName)
		if err != nil {
			http.Error(w, "missing state cookie", http.StatusBadRequest)
			return
		}
		// Clear the state cookie unconditionally so a replayed callback
		// can't reuse it even if the comparison below succeeds.
		s.clearStateCookie(w)

		stateParam := r.URL.Query().Get("state")
		if subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(stateParam)) != 1 {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		token, err := s.oauth.Exchange(ctx, code)
		if err != nil {
			s.log.Error("oauth exchange", "err", err)
			http.Error(w, "token exchange failed", http.StatusBadGateway)
			return
		}

		client := s.oauth.Client(ctx, token)
		gh, err := fetchGitHubUser(ctx, client)
		if err != nil {
			s.log.Error("fetch github user", "err", err)
			http.Error(w, "github user fetch failed", http.StatusBadGateway)
			return
		}
		email := gh.Email
		if email == "" {
			email, err = fetchPrimaryVerifiedEmail(ctx, client)
			if err != nil {
				s.log.Error("fetch github email", "err", err)
				http.Error(w, "github email fetch failed", http.StatusBadGateway)
				return
			}
		}

		user, err := s.upsertUserFromGitHub(ctx, gh, email)
		if err != nil {
			s.log.Error("upsert user from github", "err", err)
			http.Error(w, "user provisioning failed", http.StatusInternalServerError)
			return
		}

		if _, err := s.IssueSession(ctx, w, user.ID); err != nil {
			s.log.Error("issue session", "err", err)
			http.Error(w, "session issue failed", http.StatusInternalServerError)
			return
		}

		// We intentionally discard the GitHub access token here. Dispatch
		// needs GitHub only as an identity source, not to call the GitHub
		// API on behalf of the user afterward — so there's no refresh
		// logic to maintain and no long-lived third-party token to leak.
		http.Redirect(w, r, s.cfg.LoginRedirect, http.StatusFound)
	}
}

func (s *Service) clearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/api/auth",
		Domain:   s.cfg.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Service) upsertUserFromGitHub(ctx context.Context, gh githubUser, email string) (db.User, error) {
	providerID := strconv.FormatInt(gh.ID, 10)

	if u, err := s.db.FindUserByIdentity(ctx, githubProvider, providerID); err == nil {
		return s.db.UpdateUserProfile(ctx, u.ID, gh.Login, displayName(gh), gh.AvatarURL)
	} else if !errors.Is(err, db.ErrNotFound) {
		return db.User{}, err
	}

	// Fall back to email match so a user who signed up via another
	// identity provider can link their GitHub account without creating
	// a duplicate row.
	if u, err := s.db.GetUserByEmail(ctx, email); err == nil {
		if err := s.db.LinkUserIdentity(ctx, githubProvider, providerID, u.ID); err != nil {
			return db.User{}, err
		}
		return s.db.UpdateUserProfile(ctx, u.ID, gh.Login, displayName(gh), gh.AvatarURL)
	} else if !errors.Is(err, db.ErrNotFound) {
		return db.User{}, err
	}

	u, err := s.db.CreateUser(ctx, email, gh.Login, displayName(gh), gh.AvatarURL)
	if err != nil {
		return db.User{}, err
	}
	if err := s.db.LinkUserIdentity(ctx, githubProvider, providerID, u.ID); err != nil {
		return db.User{}, err
	}
	return u, nil
}

func displayName(gh githubUser) string {
	if gh.Name != "" {
		return gh.Name
	}
	return gh.Login
}

func fetchGitHubUser(ctx context.Context, client *http.Client) (githubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return githubUser{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return githubUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return githubUser{}, fmt.Errorf("github /user status %d", resp.StatusCode)
	}
	var u githubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return githubUser{}, err
	}
	return u, nil
}

func fetchPrimaryVerifiedEmail(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github /user/emails status %d", resp.StatusCode)
	}
	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", errors.New("no primary verified email on github account")
}
