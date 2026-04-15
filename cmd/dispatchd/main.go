package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/dhamariT/dispatch/internal/database"
	"github.com/dhamariT/dispatch/internal/database/dbaudit"
	"github.com/dhamariT/dispatch/internal/database/dbauthz"
	"github.com/dhamariT/dispatch/internal/database/dbmetrics"
	"github.com/dhamariT/dispatch/internal/database/memstore"
	"github.com/dhamariT/dispatch/internal/simulation"
	"github.com/go-chi/chi/v5"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Build the layered store: memstore is the backing data, dbmetrics
	// records per-method latency, dbaudit writes audit entries on
	// mutating ops, and dbauthz enforces RBAC at the boundary. Every
	// handler holds the dbauthz-wrapped store and cannot reach past it.
	backing := memstore.New()
	metrics := dbmetrics.New(backing)
	audited := dbaudit.New(metrics, log)
	store := dbauthz.New(audited)

	boot, err := bootstrap(log, store)
	if err != nil {
		log.Error("bootstrap failed", "err", err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(cors)

	// Public: no auth required, prints the operator key on first run.
	r.Get("/api/health", health)

	// Authenticated routes. The auth middleware sets a Subject on the
	// request context; downstream Store calls are gated by dbauthz.
	auth := r.Group(nil)
	auth.Use(authMiddleware(store, log))

	auth.Post("/api/experiments", createExperiment(store))
	auth.Get("/api/experiments", listExperiments(store))
	auth.Get("/api/experiments/{id}", getExperiment(store))
	auth.Post("/api/experiments/{id}/analyze", analyzeExperiment(store))
	auth.Post("/api/experiments/{id}/promote", promoteExperiment(store))
	auth.Post("/api/experiments/{id}/hold", holdExperiment(store))
	auth.Post("/api/metrics", pushSample(store))

	auth.Get("/api/devices", listDevices(store))
	auth.Get("/api/audit", listAudit(store))
	auth.Get("/api/admin/metrics", adminMetrics(metrics))

	auth.Get("/api/simulation/scenarios", listScenarios())
	auth.Post("/api/simulation/run", runScenario(store, boot.agentSubjects))

	addr := ":" + cmp.Or(os.Getenv("PORT"), "8080")
	log.Info("dispatchd starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// --- bootstrap ----------------------------------------------------

type bootstrapResult struct {
	// agentSubjects maps device ID to the agent's RBAC subject so
	// the simulation handler can construct an agent context for
	// each sample insert.
	agentSubjects map[string]database.Subject
}

// bootstrap mints an operator key and one agent key per simulated
// device on first run, then prints the plaintext to stdout exactly
// once. In a real deployment these would be generated through an
// admin CLI and stored out-of-band; for the simulation we print
// them so the dashboard and the simulation handler have something
// to authenticate with.
func bootstrap(log *slog.Logger, store database.Store) (bootstrapResult, error) {
	// Step 1: create the operator key with the system subject. The
	// system subject is the only principal allowed to bypass authz,
	// and exists specifically for this bootstrap window before any
	// real key has been minted.
	sysCtx := database.WithSubject(context.Background(), database.SystemSubject)

	opKey, opPlaintext, err := store.CreateAPIKey(sysCtx, database.CreateAPIKeyParams{
		Subject: database.SubjectOperator,
	})
	if err != nil {
		return bootstrapResult{}, err
	}
	log.Info("bootstrap operator key minted", "id", opKey.ID)
	// The plaintext is shown exactly once and never again. A fresh
	// key has to be minted to recover from a lost operator token.
	os.Stdout.WriteString("OPERATOR API KEY: " + opPlaintext + "\n")

	// Step 2: switch to the operator subject so the rest of the
	// bootstrap appears in the audit log as a real action by a
	// real principal, not as the system bypass. The audit trail
	// is only useful if it tells you who did what.
	opCtx := database.WithSubject(context.Background(), database.Subject{
		APIKeyID: opKey.ID,
		Type:     database.SubjectOperator,
	})

	// Step 3: enroll the simulated devices the canary/control split
	// targets and mint one agent key per device. The agent keys are
	// what the simulation handler uses to authenticate sample
	// inserts as the device, not as the operator.
	deviceIDs := []string{"car-1", "car-2", "car-3"}
	agentSubjects := make(map[string]database.Subject, len(deviceIDs))
	for _, id := range deviceIDs {
		if _, err := store.CreateDevice(opCtx, database.CreateDeviceParams{
			ID:    id,
			Name:  id,
			Fleet: "piracer",
		}); err != nil {
			return bootstrapResult{}, err
		}
		agentKey, agentPlaintext, err := store.CreateAPIKey(opCtx, database.CreateAPIKeyParams{
			Subject:  database.SubjectAgent,
			DeviceID: id,
		})
		if err != nil {
			return bootstrapResult{}, err
		}
		os.Stdout.WriteString("AGENT API KEY (" + id + "): " + agentPlaintext + "\n")
		agentSubjects[id] = database.Subject{
			APIKeyID: agentKey.ID,
			Type:     database.SubjectAgent,
			DeviceID: id,
		}
	}

	return bootstrapResult{agentSubjects: agentSubjects}, nil
}

// --- middleware ---------------------------------------------------

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware reads the X-API-Key header, looks up the key with
// the system subject (the only principal allowed to call
// GetAPIKeyByPlaintext), constructs the real RBAC subject, and
// attaches it to the request context. Every handler downstream
// reads this subject implicitly via the dbauthz wrapper.
func authMiddleware(store database.Store, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			plaintext := r.Header.Get("X-API-Key")
			if plaintext == "" {
				http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
				return
			}
			// System context: the bootstrap principal that exists
			// only so this single call can resolve a plaintext key
			// into a real subject. It must not leak past this line.
			sysCtx := database.WithSubject(r.Context(), database.SystemSubject)
			key, err := store.GetAPIKeyByPlaintext(sysCtx, plaintext)
			if err != nil {
				log.Info("auth: invalid api key", "err", err)
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}
			subj := database.Subject{
				APIKeyID: key.ID,
				Type:     key.Subject,
				DeviceID: key.DeviceID,
			}
			ctx := database.WithSubject(r.Context(), subj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// --- handlers -----------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeStoreErr maps the package-level sentinel errors from the
// database package to HTTP status codes. Unknown errors fall through
// as 500 so they're loud enough to investigate.
func writeStoreErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, database.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, database.ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, database.ErrInvalidArgument):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, database.ErrConflict):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createReq struct {
	DeployID       string   `json:"deploy_id"`
	CanaryDevices  []string `json:"canary_devices"`
	ControlDevices []string `json:"control_devices"`
	WindowMinutes  int      `json:"window_minutes"`
}

func createExperiment(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exp, err := store.CreateExperiment(r.Context(), database.CreateExperimentParams{
			DeployID:       req.DeployID,
			CanaryDevices:  req.CanaryDevices,
			ControlDevices: req.ControlDevices,
			WindowMinutes:  req.WindowMinutes,
		})
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, exp)
	}
}

func listExperiments(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exps, err := store.ListExperiments(r.Context())
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, exps)
	}
}

func getExperiment(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exp, err := store.GetExperiment(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, exp)
	}
}

func analyzeExperiment(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exp, err := store.AnalyzeExperiment(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, exp)
	}
}

func promoteExperiment(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exp, err := store.PromoteExperiment(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, exp)
	}
}

type holdReq struct {
	Reason string `json:"reason"`
}

func holdExperiment(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req holdReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exp, err := store.HoldExperiment(r.Context(), database.HoldExperimentParams{
			ID:     chi.URLParam(r, "id"),
			Reason: req.Reason,
		})
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, exp)
	}
}

type sampleReq struct {
	ExperimentID string  `json:"experiment_id"`
	DeviceID     string  `json:"device_id"`
	MetricName   string  `json:"metric_name"`
	Value        float64 `json:"value"`
	Group        string  `json:"group"`
}

func pushSample(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req sampleReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := store.InsertSample(r.Context(), database.InsertSampleParams{
			ExperimentID: req.ExperimentID,
			DeviceID:     req.DeviceID,
			MetricName:   req.MetricName,
			Value:        req.Value,
			Group:        database.Group(req.Group),
		})
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func listDevices(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devices, err := store.ListDevices(r.Context())
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, devices)
	}
}

func listAudit(store database.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entries, err := store.ListAuditEntries(r.Context(), database.AuditFilter{Limit: 100})
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, entries)
	}
}

func adminMetrics(metrics *dbmetrics.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, metrics.Snapshot())
	}
}

func listScenarios() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, simulation.Scenarios)
	}
}

type runReq struct {
	Scenario string `json:"scenario"`
}

// runScenario is the one place in the API that switches between
// subjects mid-request. The operator triggers the scenario, but the
// simulated samples are inserted using each device's agent context
// so the audit log and dbauthz checks behave exactly as they would
// for real device traffic. This is also a deliberate test that the
// authz boundary is enforced even from inside trusted server code.
func runScenario(store database.Store, agentSubjects map[string]database.Subject) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req runReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exp, err := simulation.Run(r.Context(), store, req.Scenario, agentSubjects)
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, exp)
	}
}

