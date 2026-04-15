package database

import (
	"context"
	"fmt"
)

// SubjectType identifies which kind of caller is making a Store
// request. The dbauthz wrapper switches on this value to decide
// which methods the caller is allowed to invoke and which rows
// they're allowed to touch.
type SubjectType string

const (
	// SubjectOperator is a human operator authenticated against the
	// dashboard. Operators can do anything except impersonate an
	// agent for sample inserts (which would defeat the audit trail).
	SubjectOperator SubjectType = "operator"

	// SubjectAgent is a device-side dispatch-agent process. Agents
	// can only push samples for their own device and read their
	// own device row. They cannot list other devices, create
	// experiments, or read audit logs.
	SubjectAgent SubjectType = "agent"

	// SubjectSystem is a sentinel used by the auth middleware
	// itself to look up an API key by its plaintext, since that
	// lookup must happen before a real subject is known. It must
	// never appear on a context that came from a request handler.
	SubjectSystem SubjectType = "system"
)

// Subject is the RBAC principal carried on every request context.
// It is constructed by the auth middleware after validating an
// API key, and consumed by the dbauthz wrapper on every Store call.
type Subject struct {
	// APIKeyID is the database ID of the key used to authenticate.
	// Audit entries reference this so revoked keys can still be
	// traced back through the log.
	APIKeyID string

	// Type is what the subject is allowed to do, broadly.
	Type SubjectType

	// DeviceID is non-empty only for SubjectAgent. It is the one
	// device this agent is allowed to write samples for.
	DeviceID string
}

// SystemSubject is the well-known principal that bypasses all
// authorization checks. It exists for the auth middleware bootstrap
// path: looking up an API key by its plaintext requires reading
// the api_keys table, which itself requires a subject. The system
// subject breaks that cycle. It must never escape the auth
// middleware into a request handler.
var SystemSubject = Subject{
	APIKeyID: "system",
	Type:     SubjectSystem,
}

type subjectKey struct{}

// WithSubject returns a context carrying the given RBAC subject.
// Auth middleware calls this once per request after validating
// credentials; every downstream Store call reads it back via
// SubjectFromContext.
func WithSubject(ctx context.Context, s Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

// SubjectFromContext returns the subject attached to ctx by
// WithSubject. The second return value is false if no subject is
// present, which the dbauthz wrapper treats as an unauthorized
// call: a request that reached the Store layer without going
// through auth middleware is a bug, not an anonymous user.
func SubjectFromContext(ctx context.Context) (Subject, bool) {
	s, ok := ctx.Value(subjectKey{}).(Subject)
	return s, ok
}

// RequireOperator is the explicit guard for HTTP endpoints that do
// not route through the Store interface and therefore cannot be
// gated by the dbauthz wrapper. The canonical case is observability
// endpoints (e.g. the dbmetrics snapshot) whose data lives on a
// concrete wrapper type, not on a Store row.
//
// Every such endpoint MUST call RequireOperator before doing any
// work. Without it, the auth middleware admits any valid API key —
// including agents — and the handler leaks operator-only data. This
// helper exists precisely so that the guard is one line and looks
// the same everywhere, instead of being reinvented (and forgotten)
// per handler.
func RequireOperator(ctx context.Context) error {
	subj, ok := SubjectFromContext(ctx)
	if !ok {
		// No subject means the auth middleware was skipped, which
		// is a routing bug. Return Unauthorized so it surfaces as
		// a 401 instead of silently succeeding.
		return fmt.Errorf("%w: missing subject", ErrUnauthorized)
	}
	if subj.Type == SubjectSystem {
		// The system subject is the bootstrap escape hatch and must
		// never reach a request handler. Treating it as denied here
		// turns a "system principal leaked into an HTTP path" bug
		// into a loud 401 instead of a silent privilege escalation.
		return fmt.Errorf("%w: system subject not allowed in handlers", ErrUnauthorized)
	}
	if subj.Type != SubjectOperator {
		return fmt.Errorf("%w: requires operator", ErrUnauthorized)
	}
	return nil
}
