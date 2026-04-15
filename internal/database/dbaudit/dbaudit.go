// Package dbaudit wraps a database.Store and writes an audit_log
// entry for every mutating operator action. Three kinds of calls
// are deliberately NOT audited:
//   - Reads: too noisy to be useful.
//   - Sample inserts: too high-volume, and they're already
//     attributable via the agent's API key on the sample row.
//   - Calls made under the system subject (bootstrap path):
//     the audit log answers "which principal did this," and the
//     system subject is the absence of one. Bootstrap activity
//     is visible via slog entries in cmd/dispatchd.bootstrap.
//
// Audit entries are written AFTER the underlying operation succeeds,
// so a failed write doesn't leave a phantom audit row. If the audit
// insert itself fails, the underlying op still returns success: a
// missing audit row is bad, but failing the user's request because
// the audit log misbehaved is worse.
package dbaudit

import (
	"context"
	"log/slog"

	"github.com/dhamariT/dispatch/internal/database"
)

type Store struct {
	inner database.Store
	log   *slog.Logger
}

func New(inner database.Store, log *slog.Logger) *Store {
	return &Store{inner: inner, log: log}
}

// audit is best-effort: a failed insert is logged but does not
// propagate, since the user-visible operation already succeeded.
// The slog warning is the breadcrumb a future operator would use
// to detect that the audit log has gone silent.
//
// Calls made under the system subject (bootstrap path) are not
// audited. A row attributed to subject_id="system" is noise — the
// audit log is supposed to answer "which authenticated principal
// did this," and the system subject is specifically the absence
// of one. Bootstrap activity has its own visibility via the slog
// entries in cmd/dispatchd.bootstrap.
func (s *Store) audit(ctx context.Context, action, targetType, targetID string, metadata map[string]any) {
	subj, ok := database.SubjectFromContext(ctx)
	if !ok {
		s.log.Warn("dbaudit: no subject on context, skipping entry", "action", action)
		return
	}
	if subj.Type == database.SubjectSystem {
		return
	}
	_, err := s.inner.InsertAuditEntry(ctx, database.InsertAuditEntryParams{
		SubjectID:  subj.APIKeyID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
	})
	if err != nil {
		s.log.Warn("dbaudit: audit insert failed", "action", action, "err", err)
	}
}

// --- Mutating, audited --------------------------------------------

func (s *Store) CreateExperiment(ctx context.Context, arg database.CreateExperimentParams) (database.Experiment, error) {
	v, err := s.inner.CreateExperiment(ctx, arg)
	if err == nil {
		s.audit(ctx, "experiment.create", "experiment", v.ID, map[string]any{
			"deploy_id":       v.DeployID,
			"canary_devices":  v.CanaryDevices,
			"control_devices": v.ControlDevices,
		})
	}
	return v, err
}

func (s *Store) AnalyzeExperiment(ctx context.Context, id string) (database.Experiment, error) {
	v, err := s.inner.AnalyzeExperiment(ctx, id)
	if err == nil {
		s.audit(ctx, "experiment.analyze", "experiment", id, map[string]any{
			"decision":    v.Decision,
			"hold_reason": v.HoldReason,
		})
	}
	return v, err
}

func (s *Store) HoldExperiment(ctx context.Context, arg database.HoldExperimentParams) (database.Experiment, error) {
	v, err := s.inner.HoldExperiment(ctx, arg)
	if err == nil {
		s.audit(ctx, "experiment.hold", "experiment", arg.ID, map[string]any{
			"reason": arg.Reason,
		})
	}
	return v, err
}

func (s *Store) PromoteExperiment(ctx context.Context, id string) (database.Experiment, error) {
	v, err := s.inner.PromoteExperiment(ctx, id)
	if err == nil {
		s.audit(ctx, "experiment.promote", "experiment", id, nil)
	}
	return v, err
}

func (s *Store) SetMetricDirection(ctx context.Context, arg database.SetMetricDirectionParams) error {
	err := s.inner.SetMetricDirection(ctx, arg)
	if err == nil {
		s.audit(ctx, "experiment.set_direction", "experiment", arg.ExperimentID, map[string]any{
			"metric":    arg.MetricName,
			"direction": arg.Direction,
		})
	}
	return err
}

func (s *Store) CreateDevice(ctx context.Context, arg database.CreateDeviceParams) (database.Device, error) {
	v, err := s.inner.CreateDevice(ctx, arg)
	if err == nil {
		s.audit(ctx, "device.create", "device", v.ID, nil)
	}
	return v, err
}

func (s *Store) CreateAPIKey(ctx context.Context, arg database.CreateAPIKeyParams) (database.APIKey, string, error) {
	k, p, err := s.inner.CreateAPIKey(ctx, arg)
	if err == nil {
		s.audit(ctx, "api_key.create", "api_key", k.ID, map[string]any{
			"subject":   k.Subject,
			"device_id": k.DeviceID,
		})
	}
	return k, p, err
}

func (s *Store) RevokeAPIKey(ctx context.Context, id string) error {
	err := s.inner.RevokeAPIKey(ctx, id)
	if err == nil {
		s.audit(ctx, "api_key.revoke", "api_key", id, nil)
	}
	return err
}

// --- Pass-through, not audited ------------------------------------

func (s *Store) GetExperiment(ctx context.Context, id string) (database.Experiment, error) {
	return s.inner.GetExperiment(ctx, id)
}

func (s *Store) ListExperiments(ctx context.Context) ([]database.Experiment, error) {
	return s.inner.ListExperiments(ctx)
}

func (s *Store) InsertSample(ctx context.Context, arg database.InsertSampleParams) error {
	return s.inner.InsertSample(ctx, arg)
}

func (s *Store) GetDevice(ctx context.Context, id string) (database.Device, error) {
	return s.inner.GetDevice(ctx, id)
}

func (s *Store) ListDevices(ctx context.Context) ([]database.Device, error) {
	return s.inner.ListDevices(ctx)
}

func (s *Store) GetAPIKeyByPlaintext(ctx context.Context, plaintext string) (database.APIKey, error) {
	return s.inner.GetAPIKeyByPlaintext(ctx, plaintext)
}

func (s *Store) InsertAuditEntry(ctx context.Context, arg database.InsertAuditEntryParams) (database.AuditEntry, error) {
	// Auditing the audit insert itself would recurse forever.
	return s.inner.InsertAuditEntry(ctx, arg)
}

func (s *Store) ListAuditEntries(ctx context.Context, filter database.AuditFilter) ([]database.AuditEntry, error) {
	return s.inner.ListAuditEntries(ctx, filter)
}
