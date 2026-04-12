package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/dhamariT/dispatch/internal/experiment"
	"github.com/dhamariT/dispatch/internal/metric"
	"github.com/dhamariT/dispatch/internal/simulation"
	"github.com/go-chi/chi/v5"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := experiment.NewStore()

	r := chi.NewRouter()
	r.Use(cors)

	r.Post("/api/experiments", createExperiment(store))
	r.Get("/api/experiments", listExperiments(store))
	r.Get("/api/experiments/{id}", getExperiment(store))
	r.Post("/api/experiments/{id}/analyze", analyzeExperiment(store))
	r.Post("/api/experiments/{id}/promote", promoteExperiment(store))
	r.Post("/api/experiments/{id}/hold", holdExperiment(store))
	r.Post("/api/metrics", pushSample(store))

	r.Get("/api/simulation/scenarios", listScenarios())
	r.Post("/api/simulation/run", runScenario(store))

	log.Info("dispatchd starting", "addr", ":8080")
	http.ListenAndServe(":8080", r)
}

type createReq struct {
	DeployID       string   `json:"deploy_id"`
	CanaryDevices  []string `json:"canary_devices"`
	ControlDevices []string `json:"control_devices"`
	WindowMinutes  int      `json:"window_minutes"`
}

func createExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.DeployID == "" || len(req.CanaryDevices) == 0 || len(req.ControlDevices) == 0 {
			http.Error(w, "deploy_id, canary_devices, and control_devices required", http.StatusBadRequest)
			return
		}
		if req.WindowMinutes <= 0 {
			req.WindowMinutes = 5
		}

		id := "exp-" + req.DeployID
		e := store.Create(id, req.DeployID, req.CanaryDevices, req.ControlDevices, req.WindowMinutes)
		json.NewEncoder(w).Encode(e.Snapshot())
	}
}

func listExperiments(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(store.List())
	}
}

func getExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e, err := store.Get(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(e.Snapshot())
	}
}

func analyzeExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e, err := store.Get(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		e.AnalyzeDefault()
		json.NewEncoder(w).Encode(e.Snapshot())
	}
}

func promoteExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e, err := store.Get(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err := e.ManualPromote(); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		json.NewEncoder(w).Encode(e.Snapshot())
	}
}

type holdReq struct {
	Reason string `json:"reason"`
}

func holdExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req holdReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		e, err := store.Get(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err := e.ManualHold(req.Reason); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		json.NewEncoder(w).Encode(e.Snapshot())
	}
}

type sampleReq struct {
	ExperimentID string  `json:"experiment_id"`
	DeviceID     string  `json:"device_id"`
	MetricName   string  `json:"metric_name"`
	Value        float64 `json:"value"`
	Group        string  `json:"group"`
}

func pushSample(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req sampleReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.ExperimentID == "" || req.DeviceID == "" || req.MetricName == "" {
			http.Error(w, "experiment_id, device_id, metric_name required", http.StatusBadRequest)
			return
		}
		g := metric.Group(req.Group)
		if g != metric.Canary && g != metric.Control {
			http.Error(w, `group must be "canary" or "control"`, http.StatusBadRequest)
			return
		}

		err := store.AddSample(metric.Sample{
			ExperimentID: req.ExperimentID,
			DeviceID:     req.DeviceID,
			MetricName:   req.MetricName,
			Value:        req.Value,
			Group:        g,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// cors allows the Next.js dev server to reach the API.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func listScenarios() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(simulation.Scenarios)
	}
}

type runReq struct {
	Scenario string `json:"scenario"`
}

func runScenario(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req runReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		e, err := simulation.Run(store, req.Scenario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(e.Snapshot())
	}
}
