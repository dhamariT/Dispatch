package experiment

import (
	"fmt"
	"sync"

	"github.com/dhamariT/dispatch/internal/metric"
)

type Store struct {
	mu          sync.RWMutex
	experiments map[string]*Experiment
}

func NewStore() *Store {
	return &Store{
		experiments: make(map[string]*Experiment),
	}
}

func (s *Store) Create(id, deployID string, canary, control []string, windowMinutes int) *Experiment {
	e := New(id, deployID, canary, control, windowMinutes)
	s.mu.Lock()
	s.experiments[id] = e
	s.mu.Unlock()
	return e
}

func (s *Store) Get(id string) (*Experiment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.experiments[id]
	if !ok {
		return nil, fmt.Errorf("experiment %s not found", id)
	}
	return e, nil
}

func (s *Store) List() []Experiment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Experiment, 0, len(s.experiments))
	for _, e := range s.experiments {
		out = append(out, e.Snapshot())
	}
	return out
}

func (s *Store) AddSample(sample metric.Sample) error {
	e, err := s.Get(sample.ExperimentID)
	if err != nil {
		return err
	}
	return e.AddSample(sample)
}
