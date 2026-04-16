package auth

import "time"

type Config struct {
	// CookieDomain is the domain set on session and state cookies. Leave
	// empty to bind the cookie to the exact host that served the response.
	CookieDomain string

	// CookieSecure controls the Secure flag on cookies. Set true in prod,
	// false for plain-HTTP local dev against the Next.js dev server.
	CookieSecure bool

	// SessionLifetime is how long a newly issued session stays valid.
	SessionLifetime time.Duration

	// LoginRedirect is where the browser is sent after a successful
	// GitHub callback. Points at the frontend origin, not the API.
	LoginRedirect string

	// GitHub OAuth app credentials.
	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
}

func (c Config) sessionLifetime() time.Duration {
	if c.SessionLifetime <= 0 {
		return 7 * 24 * time.Hour
	}
	return c.SessionLifetime
}
