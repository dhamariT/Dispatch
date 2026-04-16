package main

import (
	"net/http"

	"github.com/dhamariT/dispatch/internal/auth"
	"github.com/dhamariT/dispatch/internal/simulation"
)

func (s *server) listScenarios() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, simulation.Scenarios)
	}
}

type runScenarioReq struct {
	Scenario string `json:"scenario"`
}

func (s *server) runScenario() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, _ := auth.OrgFromCtx(r.Context())
		var req runScenarioReq
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		e, err := simulation.Run(s.expStore, org.ID, req.Scenario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, e.Snapshot())
	}
}
