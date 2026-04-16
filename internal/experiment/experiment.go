package experiment

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/dhamariT/dispatch/internal/metric"
	"github.com/google/uuid"
)

type Status string

const (
	Collecting Status = "collecting"
	Analyzing  Status = "analyzing"
	Decided    Status = "decided"
)

type Decision string

const (
	Promote  Decision = "promote"
	Hold     Decision = "hold"
	AutoHold Decision = "auto_hold"
)

type Experiment struct {
	mu sync.RWMutex

	ID             string           `json:"id"`
	OrgID          uuid.UUID        `json:"org_id"`
	DeployID       string           `json:"deploy_id"`
	Status         Status           `json:"status"`
	Decision       Decision         `json:"decision,omitzero"`
	HoldReason     string           `json:"hold_reason,omitzero"`
	CanaryDevices  []string         `json:"canary_devices"`
	ControlDevices []string         `json:"control_devices"`
	WindowMinutes  int              `json:"window_minutes"`
	StartedAt      time.Time        `json:"started_at"`
	Results        []AnalysisResult `json:"results,omitzero"`

	// samples maps metric_name → group → values.
	samples    map[string]map[metric.Group][]float64
	directions map[string]Direction
}

func New(id string, orgID uuid.UUID, deployID string, canary, control []string, windowMinutes int) *Experiment {
	return &Experiment{
		ID:             id,
		OrgID:          orgID,
		DeployID:       deployID,
		Status:         Collecting,
		CanaryDevices:  canary,
		ControlDevices: control,
		WindowMinutes:  windowMinutes,
		StartedAt:      time.Now(),
		samples:        make(map[string]map[metric.Group][]float64),
		directions:     make(map[string]Direction),
	}
}

// SetDirection registers which way is "good" for a metric.
func (e *Experiment) SetDirection(metricName string, dir Direction) {
	e.mu.Lock()
	e.directions[metricName] = dir
	e.mu.Unlock()
}

func (e *Experiment) AddSample(s metric.Sample) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Status != Collecting {
		return fmt.Errorf("experiment %s is not collecting", e.ID)
	}

	if !e.deviceBelongs(s.DeviceID, s.Group) {
		return fmt.Errorf("device %s is not in the %s group", s.DeviceID, s.Group)
	}

	byGroup, ok := e.samples[s.MetricName]
	if !ok {
		byGroup = make(map[metric.Group][]float64)
		e.samples[s.MetricName] = byGroup
	}
	byGroup[s.Group] = append(byGroup[s.Group], s.Value)
	return nil
}

// Analyze runs the statistical comparison on all collected metrics.
// Transitions the experiment from collecting → analyzing → decided.
func (e *Experiment) Analyze(alpha float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.Status = Analyzing
	e.Results = nil

	var regressions []string
	for name, byGroup := range e.samples {
		canary := byGroup[metric.Canary]
		control := byGroup[metric.Control]
		dir := e.directions[name]
		r := Analyze(name, dir, canary, control, alpha)
		e.Results = append(e.Results, r)
		if r.Verdict == Regression {
			regressions = append(regressions, fmt.Sprintf(
				"%s (p=%.4f, d=%.2f)", name, r.PValue, r.EffectSize,
			))
		}
	}

	e.Status = Decided
	if len(regressions) > 0 {
		e.Decision = AutoHold
		e.HoldReason = fmt.Sprintf("regression detected: %s", regressions)
		return
	}
	e.Decision = Promote
}

// AnalyzeDefault uses the default alpha of 0.05.
func (e *Experiment) AnalyzeDefault() {
	e.Analyze(defaultAlpha)
}

// ManualHold lets an operator override the decision.
func (e *Experiment) ManualHold(reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Status != Decided {
		return fmt.Errorf("experiment %s is not decided yet", e.ID)
	}
	e.Decision = Hold
	e.HoldReason = reason
	return nil
}

// ManualPromote lets an operator override an auto-hold.
func (e *Experiment) ManualPromote() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Status != Decided {
		return fmt.Errorf("experiment %s is not decided yet", e.ID)
	}
	e.Decision = Promote
	e.HoldReason = ""
	return nil
}

// CollectionProgress returns elapsed and total seconds.
func (e *Experiment) CollectionProgress() (elapsed, total int) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	total = e.WindowMinutes * 60
	elapsed = int(time.Since(e.StartedAt).Seconds())
	return min(elapsed, total), total
}

// Snapshot returns a read-safe copy of the experiment state.
func (e *Experiment) Snapshot() Experiment {
	e.mu.RLock()
	defer e.mu.RUnlock()

	snap := *e
	snap.CanaryDevices = slices.Clone(e.CanaryDevices)
	snap.ControlDevices = slices.Clone(e.ControlDevices)
	snap.Results = slices.Clone(e.Results)
	// Don't copy samples — callers only need results.
	snap.samples = nil
	return snap
}

func (e *Experiment) deviceBelongs(deviceID string, group metric.Group) bool {
	switch group {
	case metric.Canary:
		return slices.Contains(e.CanaryDevices, deviceID)
	case metric.Control:
		return slices.Contains(e.ControlDevices, deviceID)
	}
	return false
}
