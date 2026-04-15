# Dispatch Architecture

This document describes Dispatch's current architecture. Items the README calls "roadmap" are explicitly noted as not yet built.

## What is Dispatch?

Dispatch is a causal deploy validation system for IoT device fleets. When a release rolls out, Dispatch runs a concurrent canary-vs-control experiment: canary devices get the new code while control devices stay on the current release. It collects per-device metrics from both groups, runs a frequentist statistical comparison, and decides whether the deploy is safe to promote or should be held for investigation.

It does not use machine learning, learned baselines, or anomaly detection. It uses Welch's unequal-variances t-test and Cohen's d, both implemented from scratch.

## Current Core Architecture

Dispatch today is a single Go binary (`dispatchd`) and a Next.js dashboard. There is no fleet side yet — the simulation package generates synthetic metric streams so the system can be exercised end-to-end without real devices.

The binary contains the HTTP API, the experiment engine, the statistical analysis engine, and the simulation system. All state is held in an in-memory store guarded by `sync.RWMutex`. Nothing is persisted to disk.

## Server Components (`dispatchd`)

### Entrypoint (`cmd/dispatchd/main.go`)

Chi router. CORS middleware only. Routes:

- `POST /api/experiments` — create a canary-vs-control experiment for a deploy ID with canary and control device lists and a collection window.
- `GET /api/experiments` — list experiments.
- `GET /api/experiments/{id}` — snapshot a single experiment.
- `POST /api/experiments/{id}/analyze` — run the analysis engine over collected samples and set a verdict.
- `POST /api/experiments/{id}/promote` — operator override: force-promote a held or analyzing experiment.
- `POST /api/experiments/{id}/hold` — operator override: force-hold with a reason.
- `POST /api/metrics` — push a single metric sample (`experiment_id`, `device_id`, `metric_name`, `value`, `group`).
- `GET /api/simulation/scenarios` — list the built-in scenarios.
- `POST /api/simulation/run` — create an experiment and inject a scenario's synthetic samples.

### Experiment Engine (`internal/experiment/`)

The core of Dispatch.

**`experiment.go`** — the `Experiment` type and its lifecycle.

- States: `Collecting` → `Analyzing` → `Decided`.
- Samples arrive via `AddSample`, partitioned into canary and control pools per metric.
- `AnalyzeDefault()` runs the analysis engine and transitions to `Decided` with an overall verdict plus per-metric verdicts.
- `ManualPromote()` and `ManualHold(reason)` are the operator-override entry points. The human always has the final call.
- Each metric has a `Direction` (`HigherIsBetter` or `LowerIsBetter`) so the engine can tell a regression from an improvement.

**`analysis.go`** — the statistical engine. No external stats library.

- Welch's unequal-variances t-test.
- Cohen's d for effect size.
- Regularized incomplete beta function via Lentz's continued fraction (used to turn the Welch t-statistic into a p-value).
- Dual-gate decision rule: a metric is flagged as a regression only when **both** `p < 0.05` **and** `|d| ≥ 0.5`. Either gate alone is not enough. This prevents large-sample false positives where a tiny real difference looks statistically significant but isn't practically meaningful.

**`store.go`** — in-memory `Store` keyed by experiment ID, guarded by `sync.RWMutex`. All read/write operations go through the store.

### Metric Types (`internal/metric/`)

`sample.go` defines the `Sample` struct and the `Group` enum (`Canary` / `Control`) used by the ingest endpoint and the experiment engine.

### Simulation System (`internal/simulation/`)

Generates synthetic metric streams so the dashboard has realistic data without a live fleet. Each scenario targets a specific edge of the decision logic, not a specific fleet condition.

**`scenarios.go`** defines exactly four scenarios:

1. **`lidar_regression`** — a clear LiDAR-latency regression. The dual gate should flag it and the verdict should be auto-hold.
2. **`healthy_deploy`** — canary and control are drawn from the same distribution. Verdict should be auto-promote.
3. **`noisy_but_safe`** — statistically significant but practically small. Tests the effect-size gate: `p < 0.05` but `|d| < 0.5`, so the verdict should still be promote.
4. **`slow_regression`** — a small real degradation sitting near the decision boundary. Tests the border case.

**`generator.go`** draws samples from controlled distributions and feeds them through `store.AddSample`.

## Dashboard (`site/`)

Next.js 16 App Router. The main page is `site/src/app/page.tsx`. It talks to `dispatchd` over the REST API and provides:

- Scenario picker and "run experiment" button.
- Per-metric comparison table: canary mean ± SD, control mean ± SD, sample count per group, p-value, Cohen's d, per-metric verdict.
- Overall experiment verdict with the reasoning (which gate tripped, on which metric).
- Operator-override buttons (promote a held experiment, hold a promoted one).

Components live under `site/src/components/`. Theme and shared utilities under `site/src/theme/` and `site/src/lib/`.

## Storage

**In-memory only.** The store is a Go map guarded by `sync.RWMutex`. There is no database. Restarting `dispatchd` discards all experiments and samples.

Persistence (Postgres + TimescaleDB) is on the roadmap but not implemented. Don't write code that assumes a DB exists.

## Data Flow

```
  Operator picks a scenario in the dashboard
         │
         ▼
  POST /api/simulation/run → dispatchd creates an Experiment
         │
         ▼
  simulation.Run injects synthetic samples into both groups
  via store.AddSample (canary + control, per metric)
         │
         ▼
  POST /api/experiments/{id}/analyze → experiment.AnalyzeDefault
         │
         ▼
  For each metric: Welch's t-test → p-value (via Lentz CF beta)
                   Cohen's d     → effect size
                   Dual gate     → per-metric verdict
         │
         ▼
  Overall verdict rolls up: any regression → HOLD, else PROMOTE
         │
         ▼
  Dashboard displays means ± SD, p-values, d values, verdict
         │
         ├── Operator promotes → ManualPromote transitions state
         │
         └── Operator holds    → ManualHold with reason
```

## Roadmap (NOT implemented)

Do not assume any of the following exist. If a task requires one of them, ask before building.

- **`dispatch-agent` sidecar** — the device-side binary that would read metrics from a host app and push them to `/api/metrics` with authentication. The ingest endpoint exists, but there is no agent and no auth on the endpoint.
- **Balena cloud API integration** — release pinning, device vitals polling. None of this is wired up. Dispatch is explicitly fleet-agnostic: the validation engine has no Balena coupling.
- **Rollout orchestration** — canary selection, soak timers, wave promotion (canary → wave 2 → wave 3 → full fleet). The experiment engine handles one canary-vs-control comparison; it does not drive multi-wave rollouts.
- **Persistence** — Postgres + TimescaleDB for samples, experiments, and verdict history.
- **Auth** — no API keys, no operator auth, no agent auth. Dashboard and ingest are both open on the dev server.

## Project Structure

```
dispatch/
├── cmd/
│   └── dispatchd/          # Server entrypoint (main.go with all HTTP handlers)
├── internal/
│   ├── experiment/         # Experiment lifecycle, analysis engine, in-memory store
│   │   ├── experiment.go
│   │   ├── analysis.go     # Welch's t-test, Cohen's d, Lentz CF beta
│   │   └── store.go        # sync.RWMutex-guarded map
│   ├── metric/
│   │   └── sample.go       # Sample struct, Group enum
│   └── simulation/
│       ├── scenarios.go    # The four built-in scenarios
│       └── generator.go    # Synthetic stream generation
├── site/                   # Next.js 16 App Router dashboard
│   └── src/
│       ├── app/            # Pages (page.tsx is the experiment dashboard)
│       ├── components/
│       ├── lib/
│       └── theme/
├── go.mod                  # Only dependency: chi/v5
├── README.md
├── AGENTS.md
└── CLAUDE.md
```
