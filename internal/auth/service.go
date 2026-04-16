package auth

import (
	"log/slog"

	"github.com/dhamariT/dispatch/internal/db"
	"golang.org/x/oauth2"
)

// Service bundles the DB handle, config, and OAuth2 client. HTTP handlers
// and middleware hang off this type. Constructed once at startup.
type Service struct {
	db    *db.DB
	cfg   Config
	log   *slog.Logger
	oauth *oauth2.Config
}

func NewService(database *db.DB, cfg Config, log *slog.Logger) *Service {
	return &Service{
		db:  database,
		cfg: cfg,
		log: log,
		oauth: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubCallbackURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
			Scopes: []string{"read:user", "user:email"},
		},
	}
}
