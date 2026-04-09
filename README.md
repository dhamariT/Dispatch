# Dispatch

Per-device deployment diffing for IoT fleets. Snapshot metrics before a deploy, snapshot after, diff them per device. Built on [Balena](https://www.balena.io/).

Every fleet tool today gates on "did the update install?" Dispatch gates on "did the update make this device worse?"

<div align="center">

[How It Works](#how-it-works) · [Architecture](#architecture) · [The PiRacer Story](#the-piracer-story) · [Tech Stack](#tech-stack) · [Research](#research) · [Roadmap](#roadmap)

</div>

---

## How It Works

```
1. You trigger a deploy                   →  Dispatch snapshots every device's metrics
2. Canary device gets the update via Balena →  Installs, reboots, starts reporting
3. Dispatch snapshots the canary again     →  Compares before vs. after
4. Metrics look good for your soak time    →  Promote to the next wave
5. Something looks wrong                   →  You see exactly what changed. You decide.
6. Full fleet healthy                      →  Done. Every device updated.
```

No ML. No learned baselines. Just before/after diffing, per device. You define the metrics, the soak time, and you make the call.

## Architecture

Dispatch is a distributed system split across two boundaries: the server and the fleet.

```
                    ┌──────────────────────────────────────────────┐
                    │              dispatchd (Go)                  │
                    │                                              │
                    │  ┌──────────┐ ┌──────────┐ ┌──────────────┐ │
                    │  │ Rollout  │ │ Snapshot  │ │    Fleet     │ │
                    │  │ Engine   │ │ Engine    │ │   Monitor    │ │
                    │  └────┬─────┘ └────┬─────┘ └──────┬───────┘ │
                    │       └────────────┼──────────────┘         │
                    │              ┌─────┴──────┐                 │
                    │              │    API     │◄── Next.js      │
                    │              └─────┬──────┘    Dashboard    │
                    │                    │                         │
                    │              ┌─────┴──────┐                 │
                    │              │ PostgreSQL │                 │
                    │              │ TimescaleDB│                 │
                    │              └────────────┘                 │
                    └────────────────────┬────────────────────────┘
                                         │
                              ┌──────────┼──────────┐
                              │    Balena Cloud API  │
                              │  (vitals, releases)  │
                              └──────────┬──────────┘
                                         │
                    ┌────────────────────┬┴───────────────────┐
                    │                    │                    │
              ┌─────┴──────┐      ┌─────┴──────┐      ┌─────┴──────┐
              │   Car 1    │      │   Car 2    │      │   Car 3    │
              │  balenaOS  │      │  balenaOS  │      │  balenaOS  │
              │            │      │            │      │            │
              │ ┌────────┐ │      │ ┌────────┐ │      │ ┌────────┐ │
              │ │your app│ │      │ │your app│ │      │ │your app│ │
              │ └───┬────┘ │      │ └───┬────┘ │      │ └───┬────┘ │
              │     │      │      │     │      │      │     │      │
              │ ┌───┴────┐ │      │ ┌───┴────┐ │      │ ┌───┴────┐ │
              │ │dispatch│ │      │ │dispatch│ │      │ │dispatch│ │
              │ │agent   │ │      │ │agent   │ │      │ │agent   │ │
              │ └───┬────┘ │      │ └───┬────┘ │      │ └───┬────┘ │
              └─────┼──────┘      └─────┼──────┘      └─────┼──────┘
                    │                    │                    │
                    └────────────────────┼────────────────────┘
                                         │
                               pushes metrics via HTTP
                                         │
                                         ▼
                                    dispatchd API
```

**`dispatchd`** — single Go binary. REST API, rollout engine (canary selection, soak timers, wave promotion via Balena API), snapshot engine (before/after metric capture and diffing), fleet monitor (polls Balena for device vitals, merges with agent metrics).

**`dispatch-agent`** — lightweight sidecar container on each device. Reads metrics from your app (shared volume or local file), pushes to dispatchd. Knows nothing about rollouts, diffs, or other devices.

**Dashboard** — Next.js frontend. Per-device diffs, rollout controls, deploy triggers.

### Adding Dispatch to your fleet

**1. Deploy dispatchd** (cloud VM, local machine, wherever):
```bash
# TODO: install instructions once the server is built
```

**2. Add the agent sidecar** to your existing `docker-compose.yml`:
```yaml
services:
  my-app:
    build: ./app
    privileged: true

  dispatch-agent:
    image: dispatch/agent
    environment:
      - DISPATCH_URL=https://your-dispatch-instance.com  # where dispatchd lives
      - DISPATCH_API_KEY=your-key                        # from the dashboard
    labels:
      io.balena.features.supervisor-api: 1  # lets agent read device info from Balena's on-device API
```

**3. Your app just writes metrics to a shared file** — no SDK, no imports, no Dispatch code in your app:
```python
# this is your app code — not Dispatch-specific
import json

metrics = {
    "lidar_accuracy": compute_lidar_accuracy(),
    "fusion_latency_ms": measure_fusion_latency()
}

with open("/data/dispatch/metrics.json", "w") as f:
    json.dump(metrics, f)
```

The agent picks it up and pushes it to dispatchd. Your app never talks to Dispatch.

## The PiRacer Story

I'm testing Dispatch on three [Waveshare PiRacer AI Kit](https://www.waveshare.com/piracer-ai-kit.htm) robots. Each one runs:

- **Raspberry Pi 4** — compute
- **[Slamtec RPLIDAR C1](https://www.waveshare.com/rplidar-c1.htm)** — 360° LiDAR
- **5MP camera** — vision

They run a sensor fusion pipeline that combines LiDAR point clouds with camera data to perceive the track. A bad deploy breaks a car's ability to see — that's why Dispatch exists. When Car 2's LiDAR accuracy drops after a deploy but Cars 1 and 3 are fine, I need to see that immediately. Fleet averages would hide it.

| PiRacer Fleet | Production Equivalent |
|--------------|----------------------|
| 3 PiRacer robots | Waymo robotaxis, Tesla FSD fleet |
| Balena device OTA | Vehicle OTA (SWUpdate, RAUC) |
| Dispatch rollout engine | Tesla's staged rollout pipeline |
| Per-device before/after diffing | Waymo's fleet health verification |
| LiDAR + camera sensor fusion | Perception stack (the payload being deployed) |

## Tech Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| Backend | **Go** | API server, rollout engine, snapshot diffing |
| Frontend | **TypeScript + Next.js** | Dashboard for diffs, rollouts, and deploy controls |
| Device Agent | **Docker container** | Sidecar on each device, collects and pushes custom metrics |
| Device OTA | **[Balena](https://docs.balena.io/reference/api/overview/)** | Device vitals, release pinning, deploy triggers |
| Database | **PostgreSQL + TimescaleDB** | Fleet state, append-only snapshot storage, time-series metrics |
| CI/CD | **GitHub Actions** | Automated builds and tests |

## Research

Before building, I studied how fleet OTA works in production — from the industry leaders down to the open-source tools. Notes in [`docs/ota-industry-research/`](docs/ota-industry-research/):

- **Tesla** — staged rollouts (1% → 12% → 41% → full fleet), VIN-keyed deployments, telemetry-gated promotion
- **Waymo** — simulation-first validation, fleet health monitoring
- **Rivian** — wave-based rollouts over ~5 days, zonal architecture
- **Memfault** — closest commercial IoT platform (cohort-level version comparison)
- **Kayenta** (Netflix/Google) — conceptual ancestor of Dispatch's diffing (parallel population comparison for cloud services)
- **Uptane** — Linux Foundation automotive OTA security standard
- **Mender, hawkBit, FoundriesFactory** — open-source fleet OTA (staged rollouts, but no metric awareness)

## Roadmap

- [ ] Get all 3 PiRacers provisioned on balenaCloud
- [ ] Sensor fusion pipeline on a single car (LiDAR + camera)
- [ ] Snapshot collection service — grab per-device metrics on demand
- [ ] Snapshot diff engine — before/after comparison per device
- [ ] Rollout engine + Balena API integration
- [ ] Staged rollouts (canary → soak → promote → full fleet)
- [ ] Canary selection logic
- [ ] Version pinning
- [ ] A/B fleet testing
- [ ] Web dashboard with real-time diffs and rollout controls
- [ ] Fleet-wide rollback
- [ ] End-to-end demo: push new perception code, canary deploy, diff the snapshots, promote to fleet

## About

Built by [Dhamari Trice-Hanson](https://github.com/dhamariT) — software engineer at [Hack Club](https://hackclub.com), incoming CS student at Kettering University.
