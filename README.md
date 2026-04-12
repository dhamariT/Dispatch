# Dispatch

Causal deploy validation for IoT fleets. When you push a release, Dispatch runs canary devices against a holdback control group, compares their metrics with statistical tests, and tells you whether the deploy made things worse — with confidence scores, not gut feelings.

Every fleet tool today gates on "did the update install?" Dispatch gates on "is this deploy statistically safe to promote?"

<div align="center">

[How It Works](#how-it-works) · [Architecture](#architecture) · [The PiRacer Story](#the-piracer-story) · [Tech Stack](#tech-stack) · [Research](#research) · [Roadmap](#roadmap)

</div>

---

## How It Works

```
1. You trigger a deploy                   →  Dispatch creates an experiment
2. Fleet splits into canary + control     →  Canary gets the new release, control stays on old
3. Both groups stream metrics             →  Same time window, same conditions
4. Dispatch runs statistical comparison   →  Welch's t-test + Cohen's d per metric
5. Regression detected (p < 0.05, d > 0.5) →  Rollout auto-holds. No human needed.
6. Metrics look the same                  →  Safe to promote to the next wave.
```

This is not before/after diffing — it's a controlled experiment. Time-of-day effects, network conditions, and load changes wash out because both groups experience them simultaneously. The deploy is the only variable.

### What the engine actually computes

For each metric (LiDAR accuracy, CPU, memory, latency, anything):

- **Welch's t-test** — compares canary mean vs. control mean, not assuming equal variance. Produces a p-value.
- **Cohen's d** — standardized effect size. A p-value of 0.01 with d = 0.1 means "statistically different but practically identical." Dispatch requires both significance *and* meaningful effect size before flagging.
- **Metric direction** — the engine knows that higher CPU is bad but higher accuracy is good. A canary running hotter than control is a regression, not an improvement.
- **Automatic decision** — if any metric regresses with p < 0.05 and |d| > 0.5, the rollout holds. Operators investigate with evidence, not dashboards.

## Architecture

```
                    ┌──────────────────────────────────────────────┐
                    │              dispatchd (Go)                  │
                    │                                              │
                    │  ┌──────────┐ ┌──────────┐ ┌──────────────┐ │
                    │  │ Rollout  │ │Experiment │ │    Fleet     │ │
                    │  │ Engine   │ │  Engine   │ │   Monitor    │ │
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
              │  Car 1     │      │  Car 2     │      │  Car 3     │
              │ (canary)   │      │ (control)  │      │ (control)  │
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
                           pushes metrics via HTTP
                                         │
                                         ▼
                              Experiment Engine
                         (canary vs control, same window)
```

**`dispatchd`** — single Go binary. REST API, experiment engine (Welch's t-test, Cohen's d, auto-hold decisions), rollout engine (canary selection, wave promotion via Balena API), fleet monitor (polls Balena for device vitals, merges with agent metrics).

**`dispatch-agent`** — lightweight sidecar container on each device. Reads metrics from your app (shared volume or local file), pushes to dispatchd. Knows nothing about experiments, statistics, or other devices.

**Dashboard** — Next.js frontend. Scenario picker, experiment verdicts with statistical evidence, metric comparison tables showing canary vs. control with p-values and effect sizes.

### Adding Dispatch to your fleet

**1. Deploy dispatchd** (cloud VM, local machine, wherever):
```bash
# TODO: install instructions once the server is production-ready
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

They run a sensor fusion pipeline that combines LiDAR point clouds with camera data to perceive the track. A bad deploy breaks a car's ability to see — and Dispatch's experiment engine catches it.

When Car 1 (canary) gets a new release and its LiDAR accuracy drops to 91% while Cars 2 and 3 (control) stay at 98%, Dispatch doesn't just show the number — it runs a t-test (p < 0.0001, d = -8.9) and auto-holds the rollout before the bad code reaches the rest of the fleet.

| PiRacer Fleet | Production Equivalent |
|--|--|
| 3 PiRacers, 1 canary + 2 control | Waymo fleet with holdback cohort |
| Welch's t-test on concurrent metrics | Netflix Kayenta canary analysis |
| Auto-hold on p < 0.05 and \|d\| > 0.5 | Automated deployment gates at Google |
| Per-device metric streams | Fleet telemetry pipeline |
| LiDAR + camera sensor fusion | Perception stack (the payload being deployed) |

## Tech Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| Backend | **Go** | API server, experiment engine, statistical analysis |
| Frontend | **TypeScript + Next.js** | Dashboard for experiments, verdicts, and metric comparison |
| Statistical Engine | **Welch's t-test + Cohen's d** | Unequal variance comparison with effect size gating |
| Device Agent | **Docker container** | Sidecar on each device, collects and pushes custom metrics |
| Device OTA | **[Balena](https://docs.balena.io/reference/api/overview/)** | Device vitals, release pinning, deploy triggers |
| Database | **PostgreSQL + TimescaleDB** | Fleet state, append-only metric storage, time-series queries |
| CI/CD | **GitHub Actions** | Automated builds and tests |

## Research

Before building, I studied how fleet OTA and canary analysis work in production — from industry leaders down to open-source tools. Notes in [`docs/ota-industry-research/`](docs/ota-industry-research/):

- **Kayenta** (Netflix/Google) — the closest conceptual ancestor. Automated canary analysis for cloud services using statistical comparison of canary vs. baseline populations. Dispatch applies the same idea to IoT fleets, where per-device hardware variance makes statistical rigor harder and more necessary.
- **Tesla** — staged rollouts (1% → 12% → 41% → full fleet), VIN-keyed deployments, telemetry-gated promotion.
- **Waymo** — simulation-first validation, fleet health monitoring with holdback cohorts.
- **Rivian** — wave-based rollouts over ~5 days, zonal architecture.
- **Memfault** — closest commercial IoT platform (cohort-level version comparison).
- **Uptane** — Linux Foundation automotive OTA security standard.
- **Mender, hawkBit, FoundriesFactory** — open-source fleet OTA (staged rollouts, but no metric awareness or statistical validation).

## Roadmap

- [x] Experiment engine — Welch's t-test, Cohen's d, metric direction, auto-hold
- [x] Simulation scenarios — LiDAR regression, healthy deploy, noisy-but-safe, slow regression
- [x] REST API — experiment lifecycle, metric ingestion, simulation runner
- [x] Dashboard — scenario picker, verdict cards, canary-vs-control metric tables
- [ ] Get all 3 PiRacers provisioned on balenaCloud
- [ ] Sensor fusion pipeline on a single car (LiDAR + camera)
- [ ] Live metric streaming from dispatch-agent to experiment engine
- [ ] Bonferroni correction for multiple metric comparisons
- [ ] PostgreSQL + TimescaleDB persistence
- [ ] Balena API integration (release pinning, device vitals)
- [ ] Staged rollouts (canary → soak → promote → full fleet)
- [ ] End-to-end demo: push new perception code, canary experiment, auto-hold, investigate, promote

## About

Built by [Dhamari Trice-Hanson](https://github.com/dhamariT) — software engineer at [Hack Club](https://hackclub.com), incoming CS student at Kettering University.
