// Package memstore is the in-memory backing implementation of
// database.Store. It is fast (microsecond-range per op) because
// every entity lives in a Go map or slice behind a single mutex,
// which is the right shape for the Dispatch simulation: we don't
// have a real fleet, we don't need durability, and we want the
// SRE story to be about the layered architecture, not about
// Postgres tuning. Swapping the backing for a real database is
// a separate concern that the Store interface already isolates.
package memstore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/dhamariT/dispatch/internal/analysis"
	"github.com/dhamariT/dispatch/internal/database"
)

type Store struct {
	mu sync.Mutex

	experiments map[string]*experimentRow
	samples     []sampleRow
	directions  []directionRow
	devices     map[string]database.Device
	apiKeys     map[string]apiKeyRow
	auditLog    []database.AuditEntry

	now func() time.Time
}

// experimentRow is the in-memory backing struct. It is private to
// memstore. Public callers only ever see a database.Experiment
// snapshot returned through the Store interface.
type experimentRow struct {
	database.Experiment
}

type sampleRow struct {
	database.Sample
}

type directionRow struct {
	ExperimentID string
	MetricName   string
	Direction    analysis.Direction
}

type apiKeyRow struct {
	database.APIKey
}

func New() *Store {
	return &Store{
		experiments: make(map[string]*experimentRow),
		devices:     make(map[string]database.Device),
		apiKeys:     make(map[string]apiKeyRow),
		now:         time.Now,
	}
}

// --- Experiments --------------------------------------------------

func (s *Store) CreateExperiment(_ context.Context, arg database.CreateExperimentParams) (database.Experiment, error) {
	if arg.DeployID == "" || len(arg.CanaryDevices) == 0 || len(arg.ControlDevices) == 0 {
		return database.Experiment{}, fmt.Errorf("%w: deploy_id, canary_devices, control_devices required", database.ErrInvalidArgument)
	}
	window := arg.WindowMinutes
	if window <= 0 {
		window = 5
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := "exp-" + arg.DeployID
	if _, exists := s.experiments[id]; exists {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s already exists", database.ErrConflict, id)
	}

	row := &experimentRow{
		Experiment: database.Experiment{
			ID:             id,
			DeployID:       arg.DeployID,
			Status:         database.StatusCollecting,
			CanaryDevices:  slices.Clone(arg.CanaryDevices),
			ControlDevices: slices.Clone(arg.ControlDevices),
			WindowMinutes:  window,
			StartedAt:      s.now(),
		},
	}
	s.experiments[id] = row
	return cloneExperiment(row.Experiment), nil
}

func (s *Store) GetExperiment(_ context.Context, id string) (database.Experiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.experiments[id]
	if !ok {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s", database.ErrNotFound, id)
	}
	return cloneExperiment(row.Experiment), nil
}

func (s *Store) ListExperiments(_ context.Context) ([]database.Experiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]database.Experiment, 0, len(s.experiments))
	for _, row := range s.experiments {
		out = append(out, cloneExperiment(row.Experiment))
	}
	return out, nil
}

func (s *Store) AnalyzeExperiment(_ context.Context, id string) (database.Experiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.experiments[id]
	if !ok {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s", database.ErrNotFound, id)
	}

	row.Status = database.StatusAnalyzing
	row.Results = nil

	// Pivot the flat sample slice into per-metric, per-group buckets.
	// O(n) over samples for this experiment; cheap at simulation scale.
	type bucket struct{ canary, control []float64 }
	buckets := make(map[string]*bucket)
	for _, sr := range s.samples {
		if sr.ExperimentID != id {
			continue
		}
		b, ok := buckets[sr.MetricName]
		if !ok {
			b = &bucket{}
			buckets[sr.MetricName] = b
		}
		switch sr.Group {
		case database.GroupCanary:
			b.canary = append(b.canary, sr.Value)
		case database.GroupControl:
			b.control = append(b.control, sr.Value)
		}
	}

	dirs := s.directionsFor(id)

	var regressions []string
	for name, b := range buckets {
		r := analysis.Analyze(name, dirs[name], b.canary, b.control, analysis.DefaultAlpha)
		row.Results = append(row.Results, r)
		if r.Verdict == analysis.Regression {
			regressions = append(regressions, fmt.Sprintf("%s (p=%.4f, d=%.2f)", name, r.PValue, r.EffectSize))
		}
	}

	row.Status = database.StatusDecided
	if len(regressions) > 0 {
		row.Decision = database.DecisionAutoHold
		row.HoldReason = fmt.Sprintf("regression detected: %s", regressions)
	} else {
		row.Decision = database.DecisionPromote
		row.HoldReason = ""
	}

	return cloneExperiment(row.Experiment), nil
}

func (s *Store) HoldExperiment(_ context.Context, arg database.HoldExperimentParams) (database.Experiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.experiments[arg.ID]
	if !ok {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s", database.ErrNotFound, arg.ID)
	}
	if row.Status != database.StatusDecided {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s is not decided yet", database.ErrConflict, arg.ID)
	}
	row.Decision = database.DecisionHold
	row.HoldReason = arg.Reason
	return cloneExperiment(row.Experiment), nil
}

func (s *Store) PromoteExperiment(_ context.Context, id string) (database.Experiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.experiments[id]
	if !ok {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s", database.ErrNotFound, id)
	}
	if row.Status != database.StatusDecided {
		return database.Experiment{}, fmt.Errorf("%w: experiment %s is not decided yet", database.ErrConflict, id)
	}
	row.Decision = database.DecisionPromote
	row.HoldReason = ""
	return cloneExperiment(row.Experiment), nil
}

func (s *Store) SetMetricDirection(_ context.Context, arg database.SetMetricDirectionParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.experiments[arg.ExperimentID]; !ok {
		return fmt.Errorf("%w: experiment %s", database.ErrNotFound, arg.ExperimentID)
	}
	for i := range s.directions {
		if s.directions[i].ExperimentID == arg.ExperimentID && s.directions[i].MetricName == arg.MetricName {
			s.directions[i].Direction = arg.Direction
			return nil
		}
	}
	s.directions = append(s.directions, directionRow{
		ExperimentID: arg.ExperimentID,
		MetricName:   arg.MetricName,
		Direction:    arg.Direction,
	})
	return nil
}

// --- Samples ------------------------------------------------------

func (s *Store) InsertSample(_ context.Context, arg database.InsertSampleParams) error {
	if arg.ExperimentID == "" || arg.DeviceID == "" || arg.MetricName == "" {
		return fmt.Errorf("%w: experiment_id, device_id, metric_name required", database.ErrInvalidArgument)
	}
	if arg.Group != database.GroupCanary && arg.Group != database.GroupControl {
		return fmt.Errorf(`%w: group must be "canary" or "control"`, database.ErrInvalidArgument)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.experiments[arg.ExperimentID]
	if !ok {
		return fmt.Errorf("%w: experiment %s", database.ErrNotFound, arg.ExperimentID)
	}
	if row.Status != database.StatusCollecting {
		return fmt.Errorf("%w: experiment %s is not collecting", database.ErrConflict, arg.ExperimentID)
	}
	if !deviceInGroup(row, arg.DeviceID, arg.Group) {
		return fmt.Errorf("%w: device %s is not in the %s group", database.ErrInvalidArgument, arg.DeviceID, arg.Group)
	}

	ts := arg.Timestamp
	if ts.IsZero() {
		ts = s.now()
	}

	s.samples = append(s.samples, sampleRow{
		Sample: database.Sample{
			ID:           newID("smp"),
			ExperimentID: arg.ExperimentID,
			DeviceID:     arg.DeviceID,
			MetricName:   arg.MetricName,
			Group:        arg.Group,
			Value:        arg.Value,
			Timestamp:    ts,
		},
	})
	return nil
}

// --- Devices ------------------------------------------------------

func (s *Store) CreateDevice(_ context.Context, arg database.CreateDeviceParams) (database.Device, error) {
	if arg.ID == "" {
		return database.Device{}, fmt.Errorf("%w: device id required", database.ErrInvalidArgument)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[arg.ID]; exists {
		return database.Device{}, fmt.Errorf("%w: device %s already exists", database.ErrConflict, arg.ID)
	}
	d := database.Device{
		ID:         arg.ID,
		Name:       arg.Name,
		Fleet:      arg.Fleet,
		EnrolledAt: s.now(),
	}
	s.devices[arg.ID] = d
	return d, nil
}

func (s *Store) GetDevice(_ context.Context, id string) (database.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.devices[id]
	if !ok {
		return database.Device{}, fmt.Errorf("%w: device %s", database.ErrNotFound, id)
	}
	return d, nil
}

func (s *Store) ListDevices(_ context.Context) ([]database.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]database.Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, d)
	}
	return out, nil
}

// --- API keys -----------------------------------------------------

func (s *Store) CreateAPIKey(_ context.Context, arg database.CreateAPIKeyParams) (database.APIKey, string, error) {
	if arg.Subject != database.SubjectOperator && arg.Subject != database.SubjectAgent {
		return database.APIKey{}, "", fmt.Errorf("%w: subject must be operator or agent", database.ErrInvalidArgument)
	}
	if arg.Subject == database.SubjectAgent && arg.DeviceID == "" {
		return database.APIKey{}, "", fmt.Errorf("%w: agent keys require a device_id", database.ErrInvalidArgument)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if arg.Subject == database.SubjectAgent {
		if _, ok := s.devices[arg.DeviceID]; !ok {
			return database.APIKey{}, "", fmt.Errorf("%w: device %s", database.ErrNotFound, arg.DeviceID)
		}
	}

	plaintext := newToken()
	hashed := hashToken(plaintext)
	key := database.APIKey{
		ID:        newID("key"),
		HashedKey: hashed,
		Subject:   arg.Subject,
		DeviceID:  arg.DeviceID,
		CreatedAt: s.now(),
	}
	s.apiKeys[key.ID] = apiKeyRow{APIKey: key}
	return key, plaintext, nil
}

func (s *Store) GetAPIKeyByPlaintext(_ context.Context, plaintext string) (database.APIKey, error) {
	hashed := hashToken(plaintext)

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, row := range s.apiKeys {
		if row.HashedKey != hashed {
			continue
		}
		if !row.RevokedAt.IsZero() {
			return database.APIKey{}, fmt.Errorf("%w: api key revoked", database.ErrUnauthorized)
		}
		// Update last-used timestamp in place; the map holds a value
		// so we have to reassign.
		row.LastUsedAt = s.now()
		s.apiKeys[id] = row
		return row.APIKey, nil
	}
	return database.APIKey{}, fmt.Errorf("%w: api key", database.ErrNotFound)
}

func (s *Store) RevokeAPIKey(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.apiKeys[id]
	if !ok {
		return fmt.Errorf("%w: api key %s", database.ErrNotFound, id)
	}
	if !row.RevokedAt.IsZero() {
		return fmt.Errorf("%w: api key %s already revoked", database.ErrConflict, id)
	}
	row.RevokedAt = s.now()
	s.apiKeys[id] = row
	return nil
}

// --- Audit log ----------------------------------------------------

func (s *Store) InsertAuditEntry(_ context.Context, arg database.InsertAuditEntryParams) (database.AuditEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := database.AuditEntry{
		ID:         newID("aud"),
		Timestamp:  s.now(),
		SubjectID:  arg.SubjectID,
		Action:     arg.Action,
		TargetType: arg.TargetType,
		TargetID:   arg.TargetID,
		Metadata:   arg.Metadata,
	}
	s.auditLog = append(s.auditLog, entry)
	return entry, nil
}

func (s *Store) ListAuditEntries(_ context.Context, filter database.AuditFilter) ([]database.AuditEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Walk newest-first so callers paginating with Limit get the
	// most recent activity, which is the only useful default for
	// a security log.
	out := make([]database.AuditEntry, 0, len(s.auditLog))
	for i := len(s.auditLog) - 1; i >= 0; i-- {
		e := s.auditLog[i]
		if filter.SubjectID != "" && e.SubjectID != filter.SubjectID {
			continue
		}
		if filter.Action != "" && e.Action != filter.Action {
			continue
		}
		out = append(out, e)
		if filter.Limit > 0 && len(out) >= filter.Limit {
			break
		}
	}
	return out, nil
}

// --- helpers ------------------------------------------------------

// directionsFor must be called with s.mu held.
func (s *Store) directionsFor(expID string) map[string]analysis.Direction {
	out := make(map[string]analysis.Direction)
	for _, d := range s.directions {
		if d.ExperimentID == expID {
			out[d.MetricName] = d.Direction
		}
	}
	return out
}

func deviceInGroup(row *experimentRow, deviceID string, group database.Group) bool {
	switch group {
	case database.GroupCanary:
		return slices.Contains(row.CanaryDevices, deviceID)
	case database.GroupControl:
		return slices.Contains(row.ControlDevices, deviceID)
	}
	return false
}

// cloneExperiment returns a deep-enough copy that the caller can
// safely hold and serialize it without racing against future
// mutations under the store's mutex.
func cloneExperiment(e database.Experiment) database.Experiment {
	e.CanaryDevices = slices.Clone(e.CanaryDevices)
	e.ControlDevices = slices.Clone(e.ControlDevices)
	e.Results = slices.Clone(e.Results)
	return e
}

func newID(prefix string) string {
	var b [6]byte
	rand.Read(b[:])
	return prefix + "-" + hex.EncodeToString(b[:])
}

func newToken() string {
	// 32 bytes = 256 bits of entropy, encoded as 64 hex chars.
	var b [32]byte
	rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func hashToken(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}
