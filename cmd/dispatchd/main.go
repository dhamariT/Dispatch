package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dhamariT/dispatch/internal/auth"
	"github.com/dhamariT/dispatch/internal/db"
	"github.com/dhamariT/dispatch/internal/experiment"
)

type server struct {
	db       *db.DB
	auth     *auth.Service
	expStore *experiment.Store
	log      *slog.Logger
	corsOrigin string
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbURL := os.Getenv("DISPATCH_DATABASE_URL")
	if dbURL == "" {
		log.Error("DISPATCH_DATABASE_URL is not set")
		os.Exit(1)
	}

	ctx := context.Background()
	database, err := db.Connect(ctx, dbURL)
	if err != nil {
		log.Error("connect database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	cfg := auth.Config{
		CookieSecure:       os.Getenv("DISPATCH_COOKIE_SECURE") == "true",
		SessionLifetime:    7 * 24 * time.Hour,
		LoginRedirect:      envOr("DISPATCH_LOGIN_REDIRECT", "http://localhost:3000"),
		GitHubClientID:     os.Getenv("DISPATCH_GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("DISPATCH_GITHUB_CLIENT_SECRET"),
		GitHubCallbackURL:  envOr("DISPATCH_GITHUB_CALLBACK_URL", "http://localhost:8080/api/auth/github/callback"),
	}
	authSvc := auth.NewService(database, cfg, log)

	srv := &server{
		db:         database,
		auth:       authSvc,
		expStore:   experiment.NewStore(),
		log:        log,
		corsOrigin: envOr("DISPATCH_CORS_ORIGIN", "http://localhost:3000"),
	}

	log.Info("dispatchd starting", "addr", ":8080")
	if err := http.ListenAndServe(":8080", srv.routes()); err != nil {
		log.Error("listen", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
