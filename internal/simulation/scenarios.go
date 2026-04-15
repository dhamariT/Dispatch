package simulation

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/dhamariT/dispatch/internal/analysis"
	"github.com/dhamariT/dispatch/internal/database"
)

type Scenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var Scenarios = []Scenario{
	{Name: "lidar_regression", Description: "Canary's LiDAR accuracy drops significantly after deploy. Auto-hold."},
	{Name: "healthy_deploy", Description: "All metrics stable across canary and control. Promotes."},
	{Name: "noisy_but_safe", Description: "High variance in both groups, but no real difference. Promotes."},
	{Name: "slow_regression", Description: "Small regression in CPU — statistically borderline. Tests edge cases."},
}

const samplesPerDevice = 30

// Run creates an experiment, populates it with simulated metric
// streams for the chosen scenario, runs the analysis, and returns
// the decided experiment.
//
// The operator context creates the experiment, sets directions, and
// triggers analysis. Sample inserts switch to the per-device agent
// context so dbauthz checks behave exactly as they would for real
// device traffic — operator credentials cannot fabricate samples.
func Run(
	opCtx context.Context,
	store database.Store,
	scenarioName string,
	agentSubjects map[string]database.Subject,
) (database.Experiment, error) {
	rng := rand.New(rand.NewChaCha8([32]byte{}))

	devices := []string{"car-1", "car-2", "car-3"}
	canary := devices[:1]
	control := devices[1:]

	exp, err := store.CreateExperiment(opCtx, database.CreateExperimentParams{
		DeployID:       scenarioName + "-" + time.Now().Format("150405"),
		CanaryDevices:  canary,
		ControlDevices: control,
		WindowMinutes:  5,
	})
	if err != nil {
		return database.Experiment{}, err
	}

	// Higher accuracy is better; lower CPU/memory/latency is better.
	dirs := map[string]analysis.Direction{
		"lidar_accuracy":    analysis.HigherIsBetter,
		"cpu_usage":         analysis.LowerIsBetter,
		"memory_mb":         analysis.LowerIsBetter,
		"fusion_latency_ms": analysis.LowerIsBetter,
	}
	for name, dir := range dirs {
		if err := store.SetMetricDirection(opCtx, database.SetMetricDirectionParams{
			ExperimentID: exp.ID,
			MetricName:   name,
			Direction:    dir,
		}); err != nil {
			return database.Experiment{}, err
		}
	}

	switch scenarioName {
	case "lidar_regression":
		err = injectLidarRegression(store, agentSubjects, rng, exp.ID, canary, control)
	case "healthy_deploy", "":
		err = injectHealthy(store, agentSubjects, rng, exp.ID, canary, control)
	case "noisy_but_safe":
		err = injectNoisy(store, agentSubjects, rng, exp.ID, canary, control)
	case "slow_regression":
		err = injectSlowRegression(store, agentSubjects, rng, exp.ID, canary, control)
	default:
		return database.Experiment{}, fmt.Errorf("%w: unknown scenario %q", database.ErrInvalidArgument, scenarioName)
	}
	if err != nil {
		return database.Experiment{}, err
	}

	return store.AnalyzeExperiment(opCtx, exp.ID)
}

func injectLidarRegression(store database.Store, subs map[string]database.Subject, rng *rand.Rand, expID string, canary, control []string) error {
	// LiDAR accuracy: control stays high, canary drops.
	if err := injectMetric(store, subs, rng, expID, "lidar_accuracy", canary, 91.5, 1.2, control, 98.0, 0.3); err != nil {
		return err
	}
	// CPU: canary runs hotter from the bad code.
	if err := injectMetric(store, subs, rng, expID, "cpu_usage", canary, 48.0, 3.0, control, 33.0, 2.0); err != nil {
		return err
	}
	// Memory: similar across both groups.
	if err := injectMetric(store, subs, rng, expID, "memory_mb", canary, 415.0, 8.0, control, 412.0, 7.0); err != nil {
		return err
	}
	// Fusion latency: canary slightly worse.
	return injectMetric(store, subs, rng, expID, "fusion_latency_ms", canary, 14.0, 1.5, control, 12.0, 1.0)
}

func injectHealthy(store database.Store, subs map[string]database.Subject, rng *rand.Rand, expID string, canary, control []string) error {
	if err := injectMetric(store, subs, rng, expID, "lidar_accuracy", canary, 98.0, 0.4, control, 98.0, 0.4); err != nil {
		return err
	}
	if err := injectMetric(store, subs, rng, expID, "cpu_usage", canary, 33.0, 2.5, control, 33.0, 2.5); err != nil {
		return err
	}
	if err := injectMetric(store, subs, rng, expID, "memory_mb", canary, 412.0, 8.0, control, 412.0, 8.0); err != nil {
		return err
	}
	return injectMetric(store, subs, rng, expID, "fusion_latency_ms", canary, 12.0, 1.2, control, 12.0, 1.2)
}

func injectNoisy(store database.Store, subs map[string]database.Subject, rng *rand.Rand, expID string, canary, control []string) error {
	// High stddev on everything — means are identical, data is wild.
	if err := injectMetric(store, subs, rng, expID, "lidar_accuracy", canary, 98.0, 3.0, control, 98.0, 3.0); err != nil {
		return err
	}
	if err := injectMetric(store, subs, rng, expID, "cpu_usage", canary, 33.0, 8.0, control, 33.0, 8.0); err != nil {
		return err
	}
	if err := injectMetric(store, subs, rng, expID, "memory_mb", canary, 412.0, 20.0, control, 412.0, 20.0); err != nil {
		return err
	}
	return injectMetric(store, subs, rng, expID, "fusion_latency_ms", canary, 12.0, 4.0, control, 12.0, 4.0)
}

func injectSlowRegression(store database.Store, subs map[string]database.Subject, rng *rand.Rand, expID string, canary, control []string) error {
	if err := injectMetric(store, subs, rng, expID, "lidar_accuracy", canary, 98.0, 0.4, control, 98.0, 0.4); err != nil {
		return err
	}
	if err := injectMetric(store, subs, rng, expID, "memory_mb", canary, 412.0, 8.0, control, 412.0, 8.0); err != nil {
		return err
	}
	// CPU: small regression, borderline significance with moderate variance.
	if err := injectMetric(store, subs, rng, expID, "cpu_usage", canary, 35.5, 3.0, control, 33.0, 3.0); err != nil {
		return err
	}
	return injectMetric(store, subs, rng, expID, "fusion_latency_ms", canary, 12.0, 1.2, control, 12.0, 1.2)
}

func injectMetric(
	store database.Store,
	subs map[string]database.Subject,
	rng *rand.Rand,
	expID string,
	name string,
	canary []string, canaryMean, canarySD float64,
	control []string, controlMean, controlSD float64,
) error {
	now := time.Now()
	for _, d := range canary {
		ctx, err := agentCtx(subs, d)
		if err != nil {
			return err
		}
		for _, v := range Stream(rng, canaryMean, canarySD, samplesPerDevice) {
			if err := store.InsertSample(ctx, database.InsertSampleParams{
				ExperimentID: expID,
				DeviceID:     d,
				MetricName:   name,
				Value:        v,
				Timestamp:    now,
				Group:        database.GroupCanary,
			}); err != nil {
				return err
			}
		}
	}
	for _, d := range control {
		ctx, err := agentCtx(subs, d)
		if err != nil {
			return err
		}
		for _, v := range Stream(rng, controlMean, controlSD, samplesPerDevice) {
			if err := store.InsertSample(ctx, database.InsertSampleParams{
				ExperimentID: expID,
				DeviceID:     d,
				MetricName:   name,
				Value:        v,
				Timestamp:    now,
				Group:        database.GroupControl,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// agentCtx returns a context whose subject is the agent for the
// given device. The simulation handler must have bootstrapped agent
// subjects for every device referenced by a scenario; a missing
// entry is a programming error, not a runtime user error.
func agentCtx(subs map[string]database.Subject, deviceID string) (context.Context, error) {
	subj, ok := subs[deviceID]
	if !ok {
		return nil, fmt.Errorf("simulation: no agent subject for device %s", deviceID)
	}
	return database.WithSubject(context.Background(), subj), nil
}
