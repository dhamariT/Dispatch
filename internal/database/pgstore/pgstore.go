// Package pgstore implements database.Store backed by PostgreSQL
// via pgx. It is a drop-in replacement for memstore: the decorator
// stack (dbauthz → dbaudit → dbmetrics → pgstore) works identically.
//
// ID generation mirrors memstore's scheme (prefix + random hex) so
// existing API consumers and the dashboard don't need changes.
package pgstore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dhamariT/dispatch/internal/analysis"
	"github.com/dhamariT/dispatch/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time check that Store satisfies database.Store.
var _ database.Store = (*Store)(nil)

type Store struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, now: time.Now}
}

// Close shuts down the connection pool.
func (s *Store) Close() { s.pool.Close() }

// --- Experiments --------------------------------------------------

func (s *Store) CreateExperiment(_ context.Context, arg database.CreateExperimentParams) (database.Experiment, error) {
	if arg.DeployID == "" || len(arg.CanaryDevices) == 0 || len(arg.ControlDevices) == 0 {
		return database.Experiment{}, fmt.Errorf("%w: deploy_id, canary_devices, control_devices required", database.ErrInvalidArgument)
	}
	window := arg.WindowMinutes
	if window <= 0 {
		window = 5
	}

	ctx := context.Background()
	id := "exp-" + arg.DeployID
	now := s.now()

	_, err := s.pool.Exec(ctx, `
		INSERT INTO experiments (id, deploy_id, status, canary_devices, control_devices, window_minutes, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, arg.DeployID, string(database.StatusCollecting),
		arg.CanaryDevices, arg.ControlDevices, window, now,
	)
	if err != nil {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s already exists", database.ErrConflict, id)
	}

	return database.Experiment{
		ID:             id,
		DeployID:       arg.DeployID,
		Status:         database.StatusCollecting,
		CanaryDevices:  arg.CanaryDevices,
		ControlDevices: arg.ControlDevices,
		WindowMinutes:  window,
		StartedAt:      now,
	}, nil
}

func (s *Store) GetExperiment(_ context.Context, id string) (database.Experiment, error) {
	return s.scanExperiment(context.Background(), id)
}

func (s *Store) ListExperiments(_ context.Context) ([]database.Experiment, error) {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `
		SELECT id, deploy_id, status, decision, hold_reason,
		       canary_devices, control_devices, window_minutes,
		       started_at, results
		FROM experiments
		ORDER BY started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []database.Experiment
	for rows.Next() {
		e, err := scanExpRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) AnalyzeExperiment(_ context.Context, id string) (database.Experiment, error) {
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return database.Experiment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	exp, err := s.scanExperimentTx(ctx, tx, id)
	if err != nil {
		return database.Experiment{}, err
	}

	// Pivot samples into per-metric, per-group buckets.
	sampleRows, err := tx.Query(ctx, `
		SELECT metric_name, "group", value
		FROM samples
		WHERE experiment_id = $1`, id)
	if err != nil {
		return database.Experiment{}, err
	}
	defer sampleRows.Close()

	type bucket struct{ canary, control []float64 }
	buckets := make(map[string]*bucket)
	for sampleRows.Next() {
		var name, group string
		var val float64
		if err := sampleRows.Scan(&name, &group, &val); err != nil {
			return database.Experiment{}, err
		}
		b, ok := buckets[name]
		if !ok {
			b = &bucket{}
			buckets[name] = b
		}
		switch database.Group(group) {
		case database.GroupCanary:
			b.canary = append(b.canary, val)
		case database.GroupControl:
			b.control = append(b.control, val)
		}
	}
	if err := sampleRows.Err(); err != nil {
		return database.Experiment{}, err
	}

	dirs, err := s.directionsForTx(ctx, tx, id)
	if err != nil {
		return database.Experiment{}, err
	}

	var results []analysis.Result
	var regressions []string
	for name, b := range buckets {
		r := analysis.Analyze(name, dirs[name], b.canary, b.control, analysis.DefaultAlpha)
		results = append(results, r)
		if r.Verdict == analysis.Regression {
			regressions = append(regressions, fmt.Sprintf("%s (p=%.4f, d=%.2f)", name, r.PValue, r.EffectSize))
		}
	}

	decision := database.DecisionPromote
	holdReason := ""
	if len(regressions) > 0 {
		decision = database.DecisionAutoHold
		holdReason = fmt.Sprintf("regression detected: %s", regressions)
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return database.Experiment{}, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE experiments
		SET status = $1, decision = $2, hold_reason = $3, results = $4
		WHERE id = $5`,
		string(database.StatusDecided), string(decision), holdReason, resultsJSON, id,
	)
	if err != nil {
		return database.Experiment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return database.Experiment{}, err
	}

	exp.Status = database.StatusDecided
	exp.Decision = decision
	exp.HoldReason = holdReason
	exp.Results = results
	return exp, nil
}

func (s *Store) HoldExperiment(_ context.Context, arg database.HoldExperimentParams) (database.Experiment, error) {
	ctx := context.Background()

	tag, err := s.pool.Exec(ctx, `
		UPDATE experiments
		SET decision = $1, hold_reason = $2
		WHERE id = $3 AND status = 'decided'`,
		string(database.DecisionHold), arg.Reason, arg.ID,
	)
	if err != nil {
		return database.Experiment{}, err
	}
	if tag.RowsAffected() == 0 {
		// Distinguish not-found from wrong-state.
		if _, err := s.scanExperiment(ctx, arg.ID); err != nil {
			return database.Experiment{}, err
		}
		return database.Experiment{}, fmt.Errorf("%w: experiment %s is not decided yet", database.ErrConflict, arg.ID)
	}
	return s.scanExperiment(ctx, arg.ID)
}

func (s *Store) PromoteExperiment(_ context.Context, id string) (database.Experiment, error) {
	ctx := context.Background()

	tag, err := s.pool.Exec(ctx, `
		UPDATE experiments
		SET decision = $1, hold_reason = ''
		WHERE id = $2 AND status = 'decided'`,
		string(database.DecisionPromote), id,
	)
	if err != nil {
		return database.Experiment{}, err
	}
	if tag.RowsAffected() == 0 {
		if _, err := s.scanExperiment(ctx, id); err != nil {
			return database.Experiment{}, err
		}
		return database.Experiment{}, fmt.Errorf("%w: experiment %s is not decided yet", database.ErrConflict, id)
	}
	return s.scanExperiment(ctx, id)
}

func (s *Store) SetMetricDirection(_ context.Context, arg database.SetMetricDirectionParams) error {
	ctx := context.Background()

	// Verify experiment exists.
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM experiments WHERE id = $1)`, arg.ExperimentID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: experiment %s", database.ErrNotFound, arg.ExperimentID)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO metric_directions (experiment_id, metric_name, direction)
		VALUES ($1, $2, $3)
		ON CONFLICT (experiment_id, metric_name) DO UPDATE SET direction = EXCLUDED.direction`,
		arg.ExperimentID, arg.MetricName, int(arg.Direction),
	)
	return err
}

// --- Samples ------------------------------------------------------

func (s *Store) InsertSample(_ context.Context, arg database.InsertSampleParams) error {
	if arg.ExperimentID == "" || arg.DeviceID == "" || arg.MetricName == "" {
		return fmt.Errorf("%w: experiment_id, device_id, metric_name required", database.ErrInvalidArgument)
	}
	if arg.Group != database.GroupCanary && arg.Group != database.GroupControl {
		return fmt.Errorf(`%w: group must be "canary" or "control"`, database.ErrInvalidArgument)
	}

	ctx := context.Background()

	// Verify experiment exists and is collecting, and that the
	// device belongs to the right group.
	var status string
	var canary, control []string
	err := s.pool.QueryRow(ctx, `
		SELECT status, canary_devices, control_devices
		FROM experiments WHERE id = $1`, arg.ExperimentID,
	).Scan(&status, &canary, &control)
	if err != nil {
		return fmt.Errorf("%w: experiment %s", database.ErrNotFound, arg.ExperimentID)
	}
	if database.Status(status) != database.StatusCollecting {
		return fmt.Errorf("%w: experiment %s is not collecting", database.ErrConflict, arg.ExperimentID)
	}
	if !deviceInGroup(canary, control, arg.DeviceID, arg.Group) {
		return fmt.Errorf("%w: device %s is not in the %s group", database.ErrInvalidArgument, arg.DeviceID, arg.Group)
	}

	ts := arg.Timestamp
	if ts.IsZero() {
		ts = s.now()
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO samples (id, experiment_id, device_id, metric_name, "group", value, ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		newID("smp"), arg.ExperimentID, arg.DeviceID, arg.MetricName,
		string(arg.Group), arg.Value, ts,
	)
	return err
}

// --- Devices ------------------------------------------------------

func (s *Store) CreateDevice(_ context.Context, arg database.CreateDeviceParams) (database.Device, error) {
	if arg.ID == "" {
		return database.Device{}, fmt.Errorf("%w: device id required", database.ErrInvalidArgument)
	}

	ctx := context.Background()
	now := s.now()

	_, err := s.pool.Exec(ctx, `
		INSERT INTO devices (id, name, fleet, enrolled_at)
		VALUES ($1, $2, $3, $4)`,
		arg.ID, arg.Name, arg.Fleet, now,
	)
	if err != nil {
		return database.Device{}, fmt.Errorf("%w: device %s already exists", database.ErrConflict, arg.ID)
	}

	return database.Device{
		ID:         arg.ID,
		Name:       arg.Name,
		Fleet:      arg.Fleet,
		EnrolledAt: now,
	}, nil
}

func (s *Store) GetDevice(_ context.Context, id string) (database.Device, error) {
	ctx := context.Background()
	var d database.Device
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, fleet, enrolled_at
		FROM devices WHERE id = $1`, id,
	).Scan(&d.ID, &d.Name, &d.Fleet, &d.EnrolledAt)
	if err != nil {
		return database.Device{}, fmt.Errorf("%w: device %s", database.ErrNotFound, id)
	}
	return d, nil
}

func (s *Store) ListDevices(_ context.Context) ([]database.Device, error) {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id, name, fleet, enrolled_at FROM devices ORDER BY enrolled_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []database.Device
	for rows.Next() {
		var d database.Device
		if err := rows.Scan(&d.ID, &d.Name, &d.Fleet, &d.EnrolledAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// --- API keys -----------------------------------------------------

func (s *Store) CreateAPIKey(_ context.Context, arg database.CreateAPIKeyParams) (database.APIKey, string, error) {
	if arg.Subject != database.SubjectOperator && arg.Subject != database.SubjectAgent {
		return database.APIKey{}, "", fmt.Errorf("%w: subject must be operator or agent", database.ErrInvalidArgument)
	}
	if arg.Subject == database.SubjectAgent && arg.DeviceID == "" {
		return database.APIKey{}, "", fmt.Errorf("%w: agent keys require a device_id", database.ErrInvalidArgument)
	}

	ctx := context.Background()

	if arg.Subject == database.SubjectAgent {
		var exists bool
		err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM devices WHERE id = $1)`, arg.DeviceID).Scan(&exists)
		if err != nil {
			return database.APIKey{}, "", err
		}
		if !exists {
			return database.APIKey{}, "", fmt.Errorf("%w: device %s", database.ErrNotFound, arg.DeviceID)
		}
	}

	plaintext := newToken()
	hashed := hashToken(plaintext)
	now := s.now()
	key := database.APIKey{
		ID:        newID("key"),
		HashedKey: hashed,
		Subject:   arg.Subject,
		DeviceID:  arg.DeviceID,
		CreatedAt: now,
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO api_keys (id, hashed_key, subject, device_id, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		key.ID, key.HashedKey, string(key.Subject), key.DeviceID, key.CreatedAt,
	)
	if err != nil {
		return database.APIKey{}, "", err
	}

	return key, plaintext, nil
}

func (s *Store) GetAPIKeyByPlaintext(_ context.Context, plaintext string) (database.APIKey, error) {
	ctx := context.Background()
	hashed := hashToken(plaintext)

	var key database.APIKey
	var revokedAt, lastUsed *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT id, hashed_key, subject, device_id, created_at, last_used_at, revoked_at
		FROM api_keys WHERE hashed_key = $1`, hashed,
	).Scan(&key.ID, &key.HashedKey, &key.Subject, &key.DeviceID,
		&key.CreatedAt, &lastUsed, &revokedAt)
	if err != nil {
		return database.APIKey{}, fmt.Errorf("%w: api key", database.ErrNotFound)
	}
	if revokedAt != nil {
		return database.APIKey{}, fmt.Errorf("%w: api key revoked", database.ErrUnauthorized)
	}
	if lastUsed != nil {
		key.LastUsedAt = *lastUsed
	}

	// Update last-used timestamp.
	now := s.now()
	_, _ = s.pool.Exec(ctx, `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`, now, key.ID)
	key.LastUsedAt = now

	return key, nil
}

func (s *Store) RevokeAPIKey(_ context.Context, id string) error {
	ctx := context.Background()

	var revokedAt *time.Time
	err := s.pool.QueryRow(ctx, `SELECT revoked_at FROM api_keys WHERE id = $1`, id).Scan(&revokedAt)
	if err != nil {
		return fmt.Errorf("%w: api key %s", database.ErrNotFound, id)
	}
	if revokedAt != nil {
		return fmt.Errorf("%w: api key %s already revoked", database.ErrConflict, id)
	}

	_, err = s.pool.Exec(ctx, `UPDATE api_keys SET revoked_at = $1 WHERE id = $2`, s.now(), id)
	return err
}

// --- Audit log ----------------------------------------------------

func (s *Store) InsertAuditEntry(_ context.Context, arg database.InsertAuditEntryParams) (database.AuditEntry, error) {
	ctx := context.Background()
	now := s.now()

	meta, err := json.Marshal(arg.Metadata)
	if err != nil {
		meta = []byte("{}")
	}

	entry := database.AuditEntry{
		ID:         newID("aud"),
		Timestamp:  now,
		SubjectID:  arg.SubjectID,
		Action:     arg.Action,
		TargetType: arg.TargetType,
		TargetID:   arg.TargetID,
		Metadata:   arg.Metadata,
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_log (id, ts, subject_id, action, target_type, target_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.ID, entry.Timestamp, entry.SubjectID,
		entry.Action, entry.TargetType, entry.TargetID, meta,
	)
	if err != nil {
		return database.AuditEntry{}, err
	}
	return entry, nil
}

func (s *Store) ListAuditEntries(_ context.Context, filter database.AuditFilter) ([]database.AuditEntry, error) {
	ctx := context.Background()

	query := `SELECT id, ts, subject_id, action, target_type, target_id, metadata FROM audit_log WHERE 1=1`
	args := []any{}
	argN := 1

	if filter.SubjectID != "" {
		query += fmt.Sprintf(` AND subject_id = $%d`, argN)
		args = append(args, filter.SubjectID)
		argN++
	}
	if filter.Action != "" {
		query += fmt.Sprintf(` AND action = $%d`, argN)
		args = append(args, filter.Action)
		argN++
	}

	query += ` ORDER BY ts DESC`

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	query += fmt.Sprintf(` LIMIT $%d`, argN)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []database.AuditEntry
	for rows.Next() {
		var e database.AuditEntry
		var meta []byte
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.SubjectID, &e.Action, &e.TargetType, &e.TargetID, &meta); err != nil {
			return nil, err
		}
		if len(meta) > 0 {
			_ = json.Unmarshal(meta, &e.Metadata)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// --- helpers ------------------------------------------------------

func (s *Store) scanExperiment(ctx context.Context, id string) (database.Experiment, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, deploy_id, status, decision, hold_reason,
		       canary_devices, control_devices, window_minutes,
		       started_at, results
		FROM experiments WHERE id = $1`, id)

	e, err := scanExpFromRow(row)
	if err != nil {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s", database.ErrNotFound, id)
	}
	return e, nil
}

func (s *Store) scanExperimentTx(ctx context.Context, tx pgx.Tx, id string) (database.Experiment, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, deploy_id, status, decision, hold_reason,
		       canary_devices, control_devices, window_minutes,
		       started_at, results
		FROM experiments WHERE id = $1
		FOR UPDATE`, id)

	e, err := scanExpFromRow(row)
	if err != nil {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s", database.ErrNotFound, id)
	}
	return e, nil
}

func (s *Store) directionsForTx(ctx context.Context, tx pgx.Tx, expID string) (map[string]analysis.Direction, error) {
	rows, err := tx.Query(ctx, `
		SELECT metric_name, direction
		FROM metric_directions
		WHERE experiment_id = $1`, expID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]analysis.Direction)
	for rows.Next() {
		var name string
		var dir int
		if err := rows.Scan(&name, &dir); err != nil {
			return nil, err
		}
		out[name] = analysis.Direction(dir)
	}
	return out, rows.Err()
}

// scannable is satisfied by both pgx.Row and pgx.Rows.
type scannable interface {
	Scan(dest ...any) error
}

func scanExpFromRow(row scannable) (database.Experiment, error) {
	var e database.Experiment
	var status, decision string
	var resultsJSON []byte

	err := row.Scan(
		&e.ID, &e.DeployID, &status, &decision, &e.HoldReason,
		&e.CanaryDevices, &e.ControlDevices, &e.WindowMinutes,
		&e.StartedAt, &resultsJSON,
	)
	if err != nil {
		return database.Experiment{}, err
	}

	e.Status = database.Status(status)
	e.Decision = database.Decision(decision)
	if len(resultsJSON) > 0 && string(resultsJSON) != "[]" {
		_ = json.Unmarshal(resultsJSON, &e.Results)
	}
	return e, nil
}

func scanExpRow(rows pgx.Rows) (database.Experiment, error) {
	return scanExpFromRow(rows)
}

func deviceInGroup(canary, control []string, deviceID string, group database.Group) bool {
	devices := canary
	if group == database.GroupControl {
		devices = control
	}
	for _, d := range devices {
		if d == deviceID {
			return true
		}
	}
	return false
}

func newID(prefix string) string {
	var b [6]byte
	rand.Read(b[:])
	return prefix + "-" + hex.EncodeToString(b[:])
}

func newToken() string {
	var b [32]byte
	rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func hashToken(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}
