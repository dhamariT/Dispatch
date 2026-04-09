# Dispatch Architecture

This document provides an overview of Dispatch's architecture and core systems.

## What is Dispatch?

Dispatch is a deployment diffing system for IoT device fleets. It orchestrates staged rollouts via Balena, snapshots per-device metrics before and after each deploy, and shows operators exactly what changed on which device. It does not use ML, learned baselines, or anomaly detection — just before/after diffing, per device.

## Core Architecture

Dispatch is a distributed system split across two boundaries: the server and the fleet.

The server is a single Go binary (`dispatchd`) containing all backend subsystems: the REST API, rollout engine, snapshot engine, and fleet monitor. The frontend is a Next.js dashboard that talks to the API. Storage is PostgreSQL with TimescaleDB for time-series metric data. Snapshots are append-only and diffs are derived at query time, never stored.

The fleet side runs a lightweight sidecar container (`dispatch-agent`) on each device alongside the user's app. The agent reads metrics from the app (shared volume or local file) and pushes them to the Dispatch API over HTTP. The agent has no knowledge of rollouts, diffs, or other devices.

Dispatch does not own OTA delivery — Balena does. Dispatch reads device vitals and writes release pins through Balena's cloud API.

## Server Components (`dispatchd`)

### API (`api/`)

REST API serving two consumers:

1. **The Next.js dashboard** — deploy triggers, rollout controls, diff views, fleet status.
2. **Device agents** — authenticated metric pushes. Each agent identifies itself by device ID and authenticates via API key.

Routes are defined in `api/router.go`. Uses Chi for HTTP routing.

### Rollout Engine (`rollout/`)

Controls the deployment strategy:

- **Canary selection** — pick which device goes first, or let Dispatch choose.
- **Soak timers** — how long a canary needs to look healthy before promotion.
- **Wave promotion** — canary → wave 2 → wave 3 → full fleet.
- **Release pinning** — calls Balena's API to pin specific devices to specific releases.

The rollout engine does not auto-rollback. It flags problems and lets the operator decide.

### Snapshot Engine (`snapshot/`)

The core of Dispatch:

1. Before a deploy, snapshot every targeted device's metrics (device vitals from Balena + custom metrics from agents).
2. After the deploy and soak period, snapshot again.
3. Diff them per device.

Snapshots are immutable records tied to a device ID, deploy ID, and timestamp. Stored in TimescaleDB. Diffs are computed on read by comparing the before and after snapshots — never precomputed or cached.

### Fleet Monitor (`fleet/`)

Two data sources, one view:

- **Device vitals** (CPU, memory, temp, storage) — polled from Balena's cloud API on a configurable interval. No agent needed for these.
- **Custom app metrics** (sensor accuracy, LiDAR latency, fusion health, anything) — pushed by the dispatch-agent on each device via the API.

The fleet monitor merges both into a unified per-device metric view.

## Device Components (`dispatch-agent`)

The agent is a Docker container that runs as a sidecar alongside the user's app on each fleet device. It is intentionally simple:

1. Read metrics from a known path (e.g., `/data/dispatch/metrics.json`).
2. Push them to the Dispatch API via authenticated HTTP POST.
3. Repeat on a configurable interval.

The agent does not:
- Know about other devices in the fleet.
- Know about rollouts, waves, or deploy state.
- Make decisions about anything.

This keeps the device footprint tiny. New agent versions are only needed when the metric reporting format changes, not when rollout logic changes.

### Agent Authentication

Each agent authenticates to the Dispatch API using a `DISPATCH_API_KEY` set as an environment variable. The agent can only push metrics for its own device and read its own device config. It cannot access other devices' data or trigger deploys.

## Dashboard (`site/`)

Next.js frontend. Communicates with `dispatchd` via REST API. Provides:

- Per-device metric diffs for each deploy.
- Rollout progress and controls (promote, hold, rollback).
- Deploy triggers.
- Fleet-wide device status and version info.

Uses shadcn/ui components. Admin-specific components live in `site/components/admin/`.

## Database

PostgreSQL with TimescaleDB extension. Key design decisions:

- **Snapshots are append-only.** Never mutate historical snapshot data.
- **Diffs are derived.** Compare before/after snapshots at query time.
- **Device state is current.** Fleet monitor updates the latest known state per device.
- **Deploy history is a log.** Every deploy, its stages, and outcomes are recorded.

## External Dependencies

### Balena Cloud API

Dispatch reads and writes through Balena's REST/OData API:

- **Reads:** device vitals (CPU, memory, temp, storage), device status (online/offline), current release version.
- **Writes:** release pins (assigning a specific release to a specific device), which is how Dispatch controls which device gets which version.

Dispatch does not manage device provisioning, OS updates, or container builds — Balena handles all of that.

## Data Flow

```
  Operator clicks "Deploy"
         │
         ▼
  dispatchd snapshots all targeted devices (before)
         │
         ▼
  Rollout engine pins canary device to new release via Balena API
         │
         ▼
  Balena delivers the update to the canary device
         │
         ▼
  Canary reboots, app starts, dispatch-agent begins pushing metrics
         │
         ▼
  Soak timer runs (operator-configured duration)
         │
         ▼
  dispatchd snapshots canary again (after)
         │
         ▼
  Snapshot engine diffs before vs. after
         │
         ▼
  Dashboard shows the diff — operator decides: promote or hold
         │
         ├── Promote → next wave gets pinned → repeat
         │
         └── Hold → operator investigates, may rollback canary
```

## Project Structure

```
dispatch/
├── api/              # REST API handlers, router, middleware
├── rollout/          # Rollout engine, canary logic, soak timers
├── snapshot/         # Snapshot capture, storage, diffing
├── fleet/            # Fleet monitor, Balena API client, metric merging
├── agent/            # dispatch-agent (device-side sidecar)
├── cmd/
│   ├── dispatchd/    # Server entrypoint
│   └── agent/        # Agent entrypoint
├── db/
│   └── migrations/   # SQL migration files
├── site/             # Next.js dashboard
└── docker-compose.yml
```
