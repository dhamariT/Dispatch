// Package dbauthz wraps a database.Store and enforces RBAC at the
// data-access boundary. Every Store method pulls the RBACSubject off
// the request context and rejects calls that the subject is not
// allowed to make. This is the load-bearing layer of the whole
// architecture: a future handler that forgets to check authorization
// has nowhere to fall through, because the underlying store is
// unreachable except through this wrapper.
//
// The authorization model has exactly three principals:
//
//   - SubjectOperator: human dashboard user. Allowed everything except
//     looking up API keys by plaintext (only the auth middleware does
//     that, via the system subject).
//   - SubjectAgent: device-side dispatch-agent. May only push samples
//     for its own device and read its own device row. Cannot list
//     other devices, mutate experiments, or read audit entries.
//   - SubjectSystem: a sentinel used by the auth middleware itself,
//     because resolving an API key from a plaintext token is the
//     bootstrap step that runs before a real subject is known. It
//     bypasses every check. It must never appear on a context that
//     reached a request handler.
//
// Every guard in this file has an inline comment explaining why it
// exists. Per AGENTS.md, removing one would silently change access,
// so the comment is the breadcrumb that keeps a future reader from
// "cleaning up" a check they don't understand.
package dbauthz

import (
	"context"
	"fmt"

	"github.com/dhamariT/dispatch/internal/database"
)

type Store struct {
	inner database.Store
}

func New(inner database.Store) *Store {
	return &Store{inner: inner}
}

// authorize is the single chokepoint every Store method calls
// before delegating. It pulls the subject from the context and
// dispatches to a per-method guard. The guard returns nil to
// allow, or an ErrUnauthorized-wrapped error to deny.
func (s *Store) authorize(ctx context.Context, action string, check func(database.Subject) error) error {
	subj, ok := database.SubjectFromContext(ctx)
	if !ok {
		// A Store call with no subject means a handler skipped the
		// auth middleware. That's a bug, not anonymous access, but
		// we return Unauthorized so the bug surfaces as a 401 rather
		// than silently allowing the call.
		return fmt.Errorf("%w: %s requires an authenticated subject", database.ErrUnauthorized, action)
	}
	// The system subject is the auth middleware bootstrap escape
	// hatch. It exists because looking up an API key by plaintext
	// is the very call that establishes the real subject; without
	// this bypass the chain would never start.
	if subj.Type == database.SubjectSystem {
		return nil
	}
	return check(subj)
}

func denied(action string) error {
	return fmt.Errorf("%w: subject not allowed to %s", database.ErrUnauthorized, action)
}

// --- Experiments --------------------------------------------------

func (s *Store) CreateExperiment(ctx context.Context, arg database.CreateExperimentParams) (database.Experiment, error) {
	if err := s.authorize(ctx, "CreateExperiment", func(subj database.Subject) error {
		// Only operators can start a deploy validation. Allowing
		// agents here would let a compromised device fabricate
		// experiments and poison the audit trail.
		if subj.Type != database.SubjectOperator {
			return denied("create experiments")
		}
		return nil
	}); err != nil {
		return database.Experiment{}, err
	}
	return s.inner.CreateExperiment(ctx, arg)
}

func (s *Store) GetExperiment(ctx context.Context, id string) (database.Experiment, error) {
	if err := s.authorize(ctx, "GetExperiment", func(subj database.Subject) error {
		// Agents do not see experiment metadata. They push samples
		// blind: the experiment id is told to them out-of-band, and
		// they have no need to enumerate or inspect experiments.
		if subj.Type != database.SubjectOperator {
			return denied("read experiments")
		}
		return nil
	}); err != nil {
		return database.Experiment{}, err
	}
	return s.inner.GetExperiment(ctx, id)
}

func (s *Store) ListExperiments(ctx context.Context) ([]database.Experiment, error) {
	if err := s.authorize(ctx, "ListExperiments", func(subj database.Subject) error {
		// Listing experiments is operator-only for the same reason
		// as GetExperiment: agents have no business enumerating
		// fleet-wide deploy state.
		if subj.Type != database.SubjectOperator {
			return denied("list experiments")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.inner.ListExperiments(ctx)
}

func (s *Store) AnalyzeExperiment(ctx context.Context, id string) (database.Experiment, error) {
	if err := s.authorize(ctx, "AnalyzeExperiment", func(subj database.Subject) error {
		// Triggering analysis decides whether a deploy promotes or
		// holds. That decision belongs to operators only; an agent
		// must never be able to declare its own deploy healthy.
		if subj.Type != database.SubjectOperator {
			return denied("analyze experiments")
		}
		return nil
	}); err != nil {
		return database.Experiment{}, err
	}
	return s.inner.AnalyzeExperiment(ctx, id)
}

func (s *Store) HoldExperiment(ctx context.Context, arg database.HoldExperimentParams) (database.Experiment, error) {
	if err := s.authorize(ctx, "HoldExperiment", func(subj database.Subject) error {
		// Manual hold is an operator override; agents have no
		// authority over rollout decisions.
		if subj.Type != database.SubjectOperator {
			return denied("hold experiments")
		}
		return nil
	}); err != nil {
		return database.Experiment{}, err
	}
	return s.inner.HoldExperiment(ctx, arg)
}

func (s *Store) PromoteExperiment(ctx context.Context, id string) (database.Experiment, error) {
	if err := s.authorize(ctx, "PromoteExperiment", func(subj database.Subject) error {
		// Manual promote overrides an auto-hold. Same reasoning as
		// HoldExperiment: rollout decisions are operator-only.
		if subj.Type != database.SubjectOperator {
			return denied("promote experiments")
		}
		return nil
	}); err != nil {
		return database.Experiment{}, err
	}
	return s.inner.PromoteExperiment(ctx, id)
}

func (s *Store) SetMetricDirection(ctx context.Context, arg database.SetMetricDirectionParams) error {
	if err := s.authorize(ctx, "SetMetricDirection", func(subj database.Subject) error {
		// Direction tells the analyzer which way is "good" for a
		// metric. An agent that could flip a HigherIsBetter metric
		// to LowerIsBetter could turn a regression into an
		// "improvement" and bypass auto-hold entirely.
		if subj.Type != database.SubjectOperator {
			return denied("set metric directions")
		}
		return nil
	}); err != nil {
		return err
	}
	return s.inner.SetMetricDirection(ctx, arg)
}

// --- Samples ------------------------------------------------------

func (s *Store) InsertSample(ctx context.Context, arg database.InsertSampleParams) error {
	if err := s.authorize(ctx, "InsertSample", func(subj database.Subject) error {
		// Operators cannot push samples directly. The simulation
		// runs as agents, not as the operator, so the audit trail
		// stays honest: every sample is attributable to a device
		// agent, never to a human operator who could fabricate data.
		if subj.Type != database.SubjectAgent {
			return denied("insert samples")
		}
		// Agents may only push samples for their own device. Without
		// this guard a compromised agent key could poison another
		// device's canary data and fake a regression (or hide one).
		if subj.DeviceID != arg.DeviceID {
			return denied("insert samples for another device")
		}
		return nil
	}); err != nil {
		return err
	}
	return s.inner.InsertSample(ctx, arg)
}

// --- Devices ------------------------------------------------------

func (s *Store) CreateDevice(ctx context.Context, arg database.CreateDeviceParams) (database.Device, error) {
	if err := s.authorize(ctx, "CreateDevice", func(subj database.Subject) error {
		// Only operators enroll devices. Letting an agent register
		// a new device would let a compromised key bootstrap
		// arbitrary new identities into the fleet.
		if subj.Type != database.SubjectOperator {
			return denied("create devices")
		}
		return nil
	}); err != nil {
		return database.Device{}, err
	}
	return s.inner.CreateDevice(ctx, arg)
}

func (s *Store) GetDevice(ctx context.Context, id string) (database.Device, error) {
	if err := s.authorize(ctx, "GetDevice", func(subj database.Subject) error {
		// Operators read any device. Agents read only their own
		// device row, which is the minimum they need to confirm
		// their identity and configuration on startup.
		if subj.Type == database.SubjectOperator {
			return nil
		}
		if subj.Type == database.SubjectAgent && subj.DeviceID == id {
			return nil
		}
		return denied("read other devices")
	}); err != nil {
		return database.Device{}, err
	}
	return s.inner.GetDevice(ctx, id)
}

func (s *Store) ListDevices(ctx context.Context) ([]database.Device, error) {
	if err := s.authorize(ctx, "ListDevices", func(subj database.Subject) error {
		// Agents must not enumerate the fleet. A compromised agent
		// that could list devices would learn the IDs needed to
		// target lateral attacks.
		if subj.Type != database.SubjectOperator {
			return denied("list devices")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.inner.ListDevices(ctx)
}

// --- API keys -----------------------------------------------------

func (s *Store) CreateAPIKey(ctx context.Context, arg database.CreateAPIKeyParams) (database.APIKey, string, error) {
	if err := s.authorize(ctx, "CreateAPIKey", func(subj database.Subject) error {
		// Issuing credentials is operator-only. Agents creating
		// keys would be a privilege escalation primitive.
		if subj.Type != database.SubjectOperator {
			return denied("create api keys")
		}
		return nil
	}); err != nil {
		return database.APIKey{}, "", err
	}
	return s.inner.CreateAPIKey(ctx, arg)
}

func (s *Store) GetAPIKeyByPlaintext(ctx context.Context, plaintext string) (database.APIKey, error) {
	if err := s.authorize(ctx, "GetAPIKeyByPlaintext", func(subj database.Subject) error {
		// Plaintext lookup is the auth bootstrap path and only the
		// system subject is allowed to call it. An operator that
		// could call this would be able to validate guessed tokens
		// against the live key table.
		return denied("look up api keys by plaintext")
	}); err != nil {
		return database.APIKey{}, err
	}
	return s.inner.GetAPIKeyByPlaintext(ctx, plaintext)
}

func (s *Store) RevokeAPIKey(ctx context.Context, id string) error {
	if err := s.authorize(ctx, "RevokeAPIKey", func(subj database.Subject) error {
		// Revocation is operator-only. An agent able to revoke
		// keys could lock the operator out of their own dashboard.
		if subj.Type != database.SubjectOperator {
			return denied("revoke api keys")
		}
		return nil
	}); err != nil {
		return err
	}
	return s.inner.RevokeAPIKey(ctx, id)
}

// --- Audit log ----------------------------------------------------

func (s *Store) InsertAuditEntry(ctx context.Context, arg database.InsertAuditEntryParams) (database.AuditEntry, error) {
	if err := s.authorize(ctx, "InsertAuditEntry", func(subj database.Subject) error {
		// Only the dbaudit wrapper writes audit rows, and it does
		// so on behalf of an authenticated operator. Restricting
		// this to operators keeps an agent from forging audit
		// entries that frame an operator for an action they didn't
		// take.
		if subj.Type != database.SubjectOperator {
			return denied("insert audit entries")
		}
		return nil
	}); err != nil {
		return database.AuditEntry{}, err
	}
	return s.inner.InsertAuditEntry(ctx, arg)
}

func (s *Store) ListAuditEntries(ctx context.Context, filter database.AuditFilter) ([]database.AuditEntry, error) {
	if err := s.authorize(ctx, "ListAuditEntries", func(subj database.Subject) error {
		// Audit visibility is operator-only. An agent reading the
		// audit log could learn operator activity patterns and
		// time attacks around quiet windows.
		if subj.Type != database.SubjectOperator {
			return denied("read audit entries")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.inner.ListAuditEntries(ctx, filter)
}
