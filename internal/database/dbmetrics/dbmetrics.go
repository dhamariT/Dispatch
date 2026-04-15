// Package dbmetrics wraps a database.Store and records a count, an
// error count, and a fixed-size ring buffer of recent latencies for
// every method. The Snapshot method exposes the data as plain JSON,
// which the admin dashboard scrapes; we deliberately do not pull in
// the Prometheus client library so the simulation has zero external
// observability dependencies.
package dbmetrics

import (
	"cmp"
	"context"
	"slices"
	"sync"
	"time"

	"github.com/dhamariT/dispatch/internal/database"
)

const ringSize = 1024

type Store struct {
	inner   database.Store
	mu      sync.Mutex
	methods map[string]*methodStats
}

type methodStats struct {
	count   uint64
	errors  uint64
	ring    [ringSize]time.Duration
	ringPos int
	ringLen int
}

func New(inner database.Store) *Store {
	return &Store{
		inner:   inner,
		methods: make(map[string]*methodStats),
	}
}

// MethodSnapshot is the externally observable shape of a single
// method's stats.
type MethodSnapshot struct {
	Method   string  `json:"method"`
	Count    uint64  `json:"count"`
	Errors   uint64  `json:"errors"`
	P50Micro int64   `json:"p50_us"`
	P95Micro int64   `json:"p95_us"`
	P99Micro int64   `json:"p99_us"`
	MaxMicro int64   `json:"max_us"`
	ErrRate  float64 `json:"error_rate"`
}

// Snapshot returns a sorted slice of per-method stats, computed from
// the ring buffer at call time. Safe to call concurrently with
// Store mutations.
func (s *Store) Snapshot() []MethodSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]MethodSnapshot, 0, len(s.methods))
	for name, m := range s.methods {
		ms := MethodSnapshot{Method: name, Count: m.count, Errors: m.errors}
		if m.count > 0 {
			ms.ErrRate = float64(m.errors) / float64(m.count)
		}
		if m.ringLen > 0 {
			samples := slices.Clone(m.ring[:m.ringLen])
			slices.Sort(samples)
			ms.P50Micro = samples[percentileIdx(m.ringLen, 0.50)].Microseconds()
			ms.P95Micro = samples[percentileIdx(m.ringLen, 0.95)].Microseconds()
			ms.P99Micro = samples[percentileIdx(m.ringLen, 0.99)].Microseconds()
			ms.MaxMicro = samples[m.ringLen-1].Microseconds()
		}
		out = append(out, ms)
	}
	slices.SortFunc(out, func(a, b MethodSnapshot) int {
		return cmp.Compare(a.Method, b.Method)
	})
	return out
}

func percentileIdx(n int, p float64) int {
	idx := int(float64(n-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return idx
}

func (s *Store) record(method string, start time.Time, err error) {
	dur := time.Since(start)
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.methods[method]
	if !ok {
		m = &methodStats{}
		s.methods[method] = m
	}
	m.count++
	if err != nil {
		m.errors++
	}
	m.ring[m.ringPos] = dur
	m.ringPos = (m.ringPos + 1) % ringSize
	if m.ringLen < ringSize {
		m.ringLen++
	}
}

// --- Store passthroughs -------------------------------------------

func (s *Store) CreateExperiment(ctx context.Context, arg database.CreateExperimentParams) (database.Experiment, error) {
	start := time.Now()
	v, err := s.inner.CreateExperiment(ctx, arg)
	s.record("CreateExperiment", start, err)
	return v, err
}

func (s *Store) GetExperiment(ctx context.Context, id string) (database.Experiment, error) {
	start := time.Now()
	v, err := s.inner.GetExperiment(ctx, id)
	s.record("GetExperiment", start, err)
	return v, err
}

func (s *Store) ListExperiments(ctx context.Context) ([]database.Experiment, error) {
	start := time.Now()
	v, err := s.inner.ListExperiments(ctx)
	s.record("ListExperiments", start, err)
	return v, err
}

func (s *Store) AnalyzeExperiment(ctx context.Context, id string) (database.Experiment, error) {
	start := time.Now()
	v, err := s.inner.AnalyzeExperiment(ctx, id)
	s.record("AnalyzeExperiment", start, err)
	return v, err
}

func (s *Store) HoldExperiment(ctx context.Context, arg database.HoldExperimentParams) (database.Experiment, error) {
	start := time.Now()
	v, err := s.inner.HoldExperiment(ctx, arg)
	s.record("HoldExperiment", start, err)
	return v, err
}

func (s *Store) PromoteExperiment(ctx context.Context, id string) (database.Experiment, error) {
	start := time.Now()
	v, err := s.inner.PromoteExperiment(ctx, id)
	s.record("PromoteExperiment", start, err)
	return v, err
}

func (s *Store) SetMetricDirection(ctx context.Context, arg database.SetMetricDirectionParams) error {
	start := time.Now()
	err := s.inner.SetMetricDirection(ctx, arg)
	s.record("SetMetricDirection", start, err)
	return err
}

func (s *Store) InsertSample(ctx context.Context, arg database.InsertSampleParams) error {
	start := time.Now()
	err := s.inner.InsertSample(ctx, arg)
	s.record("InsertSample", start, err)
	return err
}

func (s *Store) CreateDevice(ctx context.Context, arg database.CreateDeviceParams) (database.Device, error) {
	start := time.Now()
	v, err := s.inner.CreateDevice(ctx, arg)
	s.record("CreateDevice", start, err)
	return v, err
}

func (s *Store) GetDevice(ctx context.Context, id string) (database.Device, error) {
	start := time.Now()
	v, err := s.inner.GetDevice(ctx, id)
	s.record("GetDevice", start, err)
	return v, err
}

func (s *Store) ListDevices(ctx context.Context) ([]database.Device, error) {
	start := time.Now()
	v, err := s.inner.ListDevices(ctx)
	s.record("ListDevices", start, err)
	return v, err
}

func (s *Store) CreateAPIKey(ctx context.Context, arg database.CreateAPIKeyParams) (database.APIKey, string, error) {
	start := time.Now()
	k, p, err := s.inner.CreateAPIKey(ctx, arg)
	s.record("CreateAPIKey", start, err)
	return k, p, err
}

func (s *Store) GetAPIKeyByPlaintext(ctx context.Context, plaintext string) (database.APIKey, error) {
	start := time.Now()
	v, err := s.inner.GetAPIKeyByPlaintext(ctx, plaintext)
	s.record("GetAPIKeyByPlaintext", start, err)
	return v, err
}

func (s *Store) RevokeAPIKey(ctx context.Context, id string) error {
	start := time.Now()
	err := s.inner.RevokeAPIKey(ctx, id)
	s.record("RevokeAPIKey", start, err)
	return err
}

func (s *Store) InsertAuditEntry(ctx context.Context, arg database.InsertAuditEntryParams) (database.AuditEntry, error) {
	start := time.Now()
	v, err := s.inner.InsertAuditEntry(ctx, arg)
	s.record("InsertAuditEntry", start, err)
	return v, err
}

func (s *Store) ListAuditEntries(ctx context.Context, filter database.AuditFilter) ([]database.AuditEntry, error) {
	start := time.Now()
	v, err := s.inner.ListAuditEntries(ctx, filter)
	s.record("ListAuditEntries", start, err)
	return v, err
}
