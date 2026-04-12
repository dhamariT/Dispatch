package simulation

import (
	"math/rand/v2"
	"time"

	"github.com/dhamariT/dispatch/internal/experiment"
	"github.com/dhamariT/dispatch/internal/metric"
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
func Run(store *experiment.Store, scenarioName string) (*experiment.Experiment, error) {
	rng := rand.New(rand.NewChaCha8([32]byte{}))

	devices := []string{"car-1", "car-2", "car-3"}
	canary := devices[:1]
	control := devices[1:]

	id := scenarioName + "-" + time.Now().Format("150405")
	e := store.Create(id, "deploy-"+scenarioName, canary, control, 5)

	// Register metric directions: higher accuracy is better,
	// lower CPU/memory/latency is better.
	e.SetDirection("lidar_accuracy", experiment.HigherIsBetter)
	e.SetDirection("cpu_usage", experiment.LowerIsBetter)
	e.SetDirection("memory_mb", experiment.LowerIsBetter)
	e.SetDirection("fusion_latency_ms", experiment.LowerIsBetter)

	switch scenarioName {
	case "lidar_regression":
		injectLidarRegression(e, rng, canary, control)
	case "healthy_deploy":
		injectHealthy(e, rng, canary, control)
	case "noisy_but_safe":
		injectNoisy(e, rng, canary, control)
	case "slow_regression":
		injectSlowRegression(e, rng, canary, control)
	default:
		injectHealthy(e, rng, canary, control)
	}

	e.AnalyzeDefault()
	return e, nil
}

func injectLidarRegression(e *experiment.Experiment, rng *rand.Rand, canary, control []string) {
	// LiDAR accuracy: control stays high, canary drops.
	injectMetric(e, rng, "lidar_accuracy", canary, 91.5, 1.2, control, 98.0, 0.3)
	// CPU: canary runs hotter from the bad code.
	injectMetric(e, rng, "cpu_usage", canary, 48.0, 3.0, control, 33.0, 2.0)
	// Memory: similar across both groups.
	injectMetric(e, rng, "memory_mb", canary, 415.0, 8.0, control, 412.0, 7.0)
	// Fusion latency: canary slightly worse.
	injectMetric(e, rng, "fusion_latency_ms", canary, 14.0, 1.5, control, 12.0, 1.0)
}

func injectHealthy(e *experiment.Experiment, rng *rand.Rand, canary, control []string) {
	injectMetric(e, rng, "lidar_accuracy", canary, 98.0, 0.4, control, 98.0, 0.4)
	injectMetric(e, rng, "cpu_usage", canary, 33.0, 2.5, control, 33.0, 2.5)
	injectMetric(e, rng, "memory_mb", canary, 412.0, 8.0, control, 412.0, 8.0)
	injectMetric(e, rng, "fusion_latency_ms", canary, 12.0, 1.2, control, 12.0, 1.2)
}

func injectNoisy(e *experiment.Experiment, rng *rand.Rand, canary, control []string) {
	// High stddev on everything — means are identical, data is wild.
	injectMetric(e, rng, "lidar_accuracy", canary, 98.0, 3.0, control, 98.0, 3.0)
	injectMetric(e, rng, "cpu_usage", canary, 33.0, 8.0, control, 33.0, 8.0)
	injectMetric(e, rng, "memory_mb", canary, 412.0, 20.0, control, 412.0, 20.0)
	injectMetric(e, rng, "fusion_latency_ms", canary, 12.0, 4.0, control, 12.0, 4.0)
}

func injectSlowRegression(e *experiment.Experiment, rng *rand.Rand, canary, control []string) {
	// LiDAR and memory are fine — same distributions.
	injectMetric(e, rng, "lidar_accuracy", canary, 98.0, 0.4, control, 98.0, 0.4)
	injectMetric(e, rng, "memory_mb", canary, 412.0, 8.0, control, 412.0, 8.0)
	// CPU: small regression, borderline significance with moderate variance.
	injectMetric(e, rng, "cpu_usage", canary, 35.5, 3.0, control, 33.0, 3.0)
	injectMetric(e, rng, "fusion_latency_ms", canary, 12.0, 1.2, control, 12.0, 1.2)
}

func injectMetric(
	e *experiment.Experiment, rng *rand.Rand,
	name string,
	canary []string, canaryMean, canarySD float64,
	control []string, controlMean, controlSD float64,
) {
	now := time.Now()
	for _, d := range canary {
		for _, v := range Stream(rng, canaryMean, canarySD, samplesPerDevice) {
			e.AddSample(metric.Sample{
				ExperimentID: e.ID,
				DeviceID:     d,
				MetricName:   name,
				Value:        v,
				Timestamp:    now,
				Group:        metric.Canary,
			})
		}
	}
	for _, d := range control {
		for _, v := range Stream(rng, controlMean, controlSD, samplesPerDevice) {
			e.AddSample(metric.Sample{
				ExperimentID: e.ID,
				DeviceID:     d,
				MetricName:   name,
				Value:        v,
				Timestamp:    now,
				Group:        metric.Control,
			})
		}
	}
}
