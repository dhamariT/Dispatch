package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
)

// store holds everything in memory. Throwaway — just proving the loop.
type store struct {
	mu        sync.RWMutex
	metrics   map[string]map[string]float64            // device_id → metric_name → value
	snapshots map[string]map[string]map[string]float64 // deploy_id:phase → device_id → metric_name → value
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	s := &store{
		metrics:   make(map[string]map[string]float64),
		snapshots: make(map[string]map[string]map[string]float64),
	}

	r := chi.NewRouter()
	r.Post("/api/metrics", s.pushMetrics)
	r.Post("/api/snapshots", s.takeSnapshot)
	r.Get("/api/snapshots/{deployID}/diff", s.diff)

	log.Info("dispatchd starting", "addr", ":8080")
	http.ListenAndServe(":8080", r)
}

type metricsReq struct {
	DeviceID string             `json:"device_id"`
	Metrics  map[string]float64 `json:"metrics"`
}

// pushMetrics accepts metrics from a device (or curl).
func (s *store) pushMetrics(w http.ResponseWriter, r *http.Request) {
	var req metricsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.DeviceID == "" || len(req.Metrics) == 0 {
		http.Error(w, "device_id and metrics required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.metrics[req.DeviceID] = req.Metrics
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

type snapshotReq struct {
	DeployID string `json:"deploy_id"`
	Phase    string `json:"phase"` // "before" or "after"
}

// takeSnapshot freezes current metrics for all devices into a snapshot.
func (s *store) takeSnapshot(w http.ResponseWriter, r *http.Request) {
	var req snapshotReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.DeployID == "" || (req.Phase != "before" && req.Phase != "after") {
		http.Error(w, `deploy_id and phase ("before"/"after") required`, http.StatusBadRequest)
		return
	}

	key := req.DeployID + ":" + req.Phase

	s.mu.Lock()
	snap := make(map[string]map[string]float64, len(s.metrics))
	for device, metrics := range s.metrics {
		copy := make(map[string]float64, len(metrics))
		for k, v := range metrics {
			copy[k] = v
		}
		snap[device] = copy
	}
	s.snapshots[key] = snap
	s.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]any{
		"deploy_id": req.DeployID,
		"phase":     req.Phase,
		"devices":   len(snap),
	})
}

type metricDiff struct {
	Before float64 `json:"before"`
	After  float64 `json:"after"`
	Delta  float64 `json:"delta"`
}

// diff compares before and after snapshots for a deploy.
func (s *store) diff(w http.ResponseWriter, r *http.Request) {
	deployID := chi.URLParam(r, "deployID")

	s.mu.RLock()
	before := s.snapshots[deployID+":before"]
	after := s.snapshots[deployID+":after"]
	s.mu.RUnlock()

	if before == nil {
		http.Error(w, "no before snapshot for "+deployID, http.StatusNotFound)
		return
	}
	if after == nil {
		http.Error(w, "no after snapshot for "+deployID, http.StatusNotFound)
		return
	}

	// Per-device diff.
	result := make(map[string]map[string]metricDiff)
	for device, afterMetrics := range after {
		beforeMetrics := before[device]
		if beforeMetrics == nil {
			continue
		}
		diffs := make(map[string]metricDiff)
		for name, afterVal := range afterMetrics {
			beforeVal := beforeMetrics[name]
			diffs[name] = metricDiff{
				Before: beforeVal,
				After:  afterVal,
				Delta:  afterVal - beforeVal,
			}
		}
		result[device] = diffs
	}

	json.NewEncoder(w).Encode(result)
}
