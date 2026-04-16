package experiment

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dhamariT/dispatch/internal/metric"
	"github.com/google/uuid"
)

// ErrNotFound is returned when an experiment does not exist for the
// requesting org. We use the same error for "no such experiment" and
// "experiment exists but belongs to another org" so handlers can always
// return 404 without leaking cross-tenant existence.
var ErrNotFound = errors.New("experiment not found")

type Store struct {
	mu          sync.RWMutex
	experiments map[string]*Experiment
}

func NewStore() *Store {
	return &Store{
		experiments: make(map[string]*Experiment),
	}
}

func (s *Store) Create(id string, orgID uuid.UUID, deployID string, canary, control []string, windowMinutes int) *Experiment {
	e := New(id, orgID, deployID, canary, control, windowMinutes)
	s.mu.Lock()
	s.experiments[id] = e
	s.mu.Unlock()
	return e
}

// Get returns the experiment only if it belongs to orgID. Cross-tenant
// requests come back as ErrNotFound so callers never leak the fact that
// the ID exists under a different org.
func (s *Store) Get(orgID uuid.UUID, id string) (*Experiment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.experiments[id]
	if !ok || e.OrgID != orgID {
		return nil, ErrNotFound
	}
	return e, nil
}

func (s *Store) List(orgID uuid.UUID) []Experiment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Experiment, 0, len(s.experiments))
	for _, e := range s.experiments {
		if e.OrgID == orgID {
			out = append(out, e.Snapshot())
		}
	}
	return out
}

// AddSample appends a sample if the experiment exists AND belongs to
// orgID. This is the operator-scoped ingest path.
func (s *Store) AddSample(orgID uuid.UUID, sample metric.Sample) error {
	e, err := s.Get(orgID, sample.ExperimentID)
	if err != nil {
		return fmt.Errorf("add sample: %w", err)
	}
	return e.AddSample(sample)
}

// AddSampleRaw adds a sample by experiment_id without any org check.
// It exists only for the unauthenticated /api/metrics ingest endpoint
// used by device agents. Once agents carry per-device credentials this
// method should be deleted and the handler should validate that the
// credential is authorized for the specific (experiment, device_id).
func (s *Store) AddSampleRaw(sample metric.Sample) error {
	s.mu.RLock()
	e, ok := s.experiments[sample.ExperimentID]
	s.mu.RUnlock()
	if !ok {
		return ErrNotFound
	}
	return e.AddSample(sample)
}
