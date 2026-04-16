package main

import (
	"errors"
	"net/http"

	"github.com/dhamariT/dispatch/internal/auth"
	"github.com/dhamariT/dispatch/internal/experiment"
	"github.com/dhamariT/dispatch/internal/metric"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createExperimentReq struct {
	DeployID       string   `json:"deploy_id"`
	CanaryDevices  []string `json:"canary_devices"`
	ControlDevices []string `json:"control_devices"`
	WindowMinutes  int      `json:"window_minutes"`
}

func (s *server) createExperiment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())

		var req createExperimentReq
		if err := decodeJSON(r, &req); err != nil {
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

		// Experiment IDs are globally unique, opaque strings. We pair
		// them with the org scope at lookup time rather than in the
		// ID format, so no caller ever has to parse the ID.
		id := "exp-" + uuid.NewString()
		e := s.expStore.Create(id, org.ID, req.DeployID, req.CanaryDevices, req.ControlDevices, req.WindowMinutes)
		writeJSON(w, http.StatusCreated, e.Snapshot())
	}
}

func (s *server) listExperiments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		writeJSON(w, http.StatusOK, s.expStore.List(org.ID))
	}
}

func (s *server) getExperiment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		e, err := s.expStore.Get(org.ID, chi.URLParam(r, "id"))
		if err != nil {
			writeExperimentError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, e.Snapshot())
	}
}

func (s *server) analyzeExperiment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		e, err := s.expStore.Get(org.ID, chi.URLParam(r, "id"))
		if err != nil {
			writeExperimentError(w, err)
			return
		}
		e.AnalyzeDefault()
		writeJSON(w, http.StatusOK, e.Snapshot())
	}
}

func (s *server) promoteExperiment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		e, err := s.expStore.Get(org.ID, chi.URLParam(r, "id"))
		if err != nil {
			writeExperimentError(w, err)
			return
		}
		if err := e.ManualPromote(); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		writeJSON(w, http.StatusOK, e.Snapshot())
	}
}

type holdReq struct {
	Reason string `json:"reason"`
}

func (s *server) holdExperiment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		var req holdReq
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		e, err := s.expStore.Get(org.ID, chi.URLParam(r, "id"))
		if err != nil {
			writeExperimentError(w, err)
			return
		}
		if err := e.ManualHold(req.Reason); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		writeJSON(w, http.StatusOK, e.Snapshot())
	}
}

type sampleReq struct {
	ExperimentID string  `json:"experiment_id"`
	DeviceID     string  `json:"device_id"`
	MetricName   string  `json:"metric_name"`
	Value        float64 `json:"value"`
	Group        string  `json:"group"`
}

func (s *server) pushSample() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req sampleReq
		if err := decodeJSON(r, &req); err != nil {
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
		// This endpoint bypasses org-level authorization because the
		// device-agent credential model does not exist yet. Once it
		// does, the credential will carry (org_id, device_id) and this
		// handler will verify the sample matches.
		err := s.expStore.AddSampleRaw(metric.Sample{
			ExperimentID: req.ExperimentID,
			DeviceID:     req.DeviceID,
			MetricName:   req.MetricName,
			Value:        req.Value,
			Group:        g,
		})
		if err != nil {
			writeExperimentError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeExperimentError(w http.ResponseWriter, err error) {
	if errors.Is(err, experiment.ErrNotFound) {
		http.Error(w, "experiment not found", http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
