// Package database defines the Store interface that every part of
// dispatchd talks to for persistence. It deliberately mirrors the
// shape of a sqlc-generated querier: every method takes a context,
// every method takes a typed Params struct, and every method returns
// value types (never live pointers). That shape is what lets the
// dbauthz, dbaudit, and dbmetrics decorators wrap each call without
// caring about the backing implementation.
//
// The concrete in-memory backing lives in database/memstore. The
// wrappers live in database/dbauthz, database/dbaudit, and
// database/dbmetrics. The wiring happens in cmd/dispatchd.
package database

import (
	"context"
	"time"

	"github.com/dhamariT/dispatch/internal/analysis"
)

// Store is the single abstraction every handler depends on. Every
// mutation in Dispatch flows through one of these methods, which is
// what makes the decorator layers actually load-bearing: a future
// handler that wants to skip authorization has nowhere to call.
type Store interface {
	// Experiments.
	CreateExperiment(ctx context.Context, arg CreateExperimentParams) (Experiment, error)
	GetExperiment(ctx context.Context, id string) (Experiment, error)
	ListExperiments(ctx context.Context) ([]Experiment, error)
	AnalyzeExperiment(ctx context.Context, id string) (Experiment, error)
	HoldExperiment(ctx context.Context, arg HoldExperimentParams) (Experiment, error)
	PromoteExperiment(ctx context.Context, id string) (Experiment, error)
	SetMetricDirection(ctx context.Context, arg SetMetricDirectionParams) error

	// Samples.
	InsertSample(ctx context.Context, arg InsertSampleParams) error

	// Devices.
	CreateDevice(ctx context.Context, arg CreateDeviceParams) (Device, error)
	GetDevice(ctx context.Context, id string) (Device, error)
	ListDevices(ctx context.Context) ([]Device, error)

	// API keys. CreateAPIKey returns the plaintext token exactly
	// once; the caller is responsible for showing it to the user
	// and never persisting it.
	CreateAPIKey(ctx context.Context, arg CreateAPIKeyParams) (key APIKey, plaintext string, err error)
	GetAPIKeyByPlaintext(ctx context.Context, plaintext string) (APIKey, error)
	RevokeAPIKey(ctx context.Context, id string) error

	// Audit log.
	InsertAuditEntry(ctx context.Context, arg InsertAuditEntryParams) (AuditEntry, error)
	ListAuditEntries(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)
}

// Experiment is the value-type representation of a deploy validation
// experiment. The in-memory backing struct may carry mutexes and
// running aggregates; this is what crosses the Store boundary.
type Experiment struct {
	ID             string            `json:"id"`
	DeployID       string            `json:"deploy_id"`
	Status         Status            `json:"status"`
	Decision       Decision          `json:"decision,omitzero"`
	HoldReason     string            `json:"hold_reason,omitzero"`
	CanaryDevices  []string          `json:"canary_devices"`
	ControlDevices []string          `json:"control_devices"`
	WindowMinutes  int               `json:"window_minutes"`
	StartedAt      time.Time         `json:"started_at"`
	Results        []analysis.Result `json:"results,omitzero"`
}

type Status string

const (
	StatusCollecting Status = "collecting"
	StatusAnalyzing  Status = "analyzing"
	StatusDecided    Status = "decided"
)

type Decision string

const (
	DecisionPromote  Decision = "promote"
	DecisionHold     Decision = "hold"
	DecisionAutoHold Decision = "auto_hold"
)

type Group string

const (
	GroupCanary  Group = "canary"
	GroupControl Group = "control"
)

type Sample struct {
	ID           string    `json:"id"`
	ExperimentID string    `json:"experiment_id"`
	DeviceID     string    `json:"device_id"`
	MetricName   string    `json:"metric_name"`
	Group        Group     `json:"group"`
	Value        float64   `json:"value"`
	Timestamp    time.Time `json:"timestamp"`
}

type Device struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Fleet      string    `json:"fleet"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

type APIKey struct {
	ID         string      `json:"id"`
	HashedKey  string      `json:"-"`
	Subject    SubjectType `json:"subject"`
	DeviceID   string      `json:"device_id,omitzero"`
	CreatedAt  time.Time   `json:"created_at"`
	LastUsedAt time.Time   `json:"last_used_at,omitzero"`
	RevokedAt  time.Time   `json:"revoked_at,omitzero"`
}

type AuditEntry struct {
	ID         string         `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	SubjectID  string         `json:"subject_id"`
	Action     string         `json:"action"`
	TargetType string         `json:"target_type"`
	TargetID   string         `json:"target_id"`
	Metadata   map[string]any `json:"metadata,omitzero"`
}

// CreateExperimentParams etc. are passed by value into Store methods.
// They mirror sqlc's generated argument structs.
type CreateExperimentParams struct {
	DeployID       string
	CanaryDevices  []string
	ControlDevices []string
	WindowMinutes  int
}

type HoldExperimentParams struct {
	ID     string
	Reason string
}

type SetMetricDirectionParams struct {
	ExperimentID string
	MetricName   string
	Direction    analysis.Direction
}

type InsertSampleParams struct {
	ExperimentID string
	DeviceID     string
	MetricName   string
	Group        Group
	Value        float64
	Timestamp    time.Time
}

type CreateDeviceParams struct {
	ID    string
	Name  string
	Fleet string
}

type CreateAPIKeyParams struct {
	Subject  SubjectType
	DeviceID string
}

type InsertAuditEntryParams struct {
	SubjectID  string
	Action     string
	TargetType string
	TargetID   string
	Metadata   map[string]any
}

type AuditFilter struct {
	SubjectID string
	Action    string
	Limit     int
}
