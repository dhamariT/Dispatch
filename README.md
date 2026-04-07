# Dispatch

**Every node in your fleet gets a cryptographically signed update, verifies it, installs to a backup partition, reboots, checks its own health, and rolls back if anything breaks. No human in the loop.**

I wanted to push self-driving code to a fleet of racing robots without bricking them. So I built Dispatch: a fleet orchestration layer on top of Balena that handles staged rollouts, health-gated promotions, automatic rollback, and A/B testing across a fleet. Balena handles the plumbing. Dispatch is the brain.

<div align="center">

[How It Works](#how-it-works) | [Features](#features) | [Architecture](#architecture) | [Tech Stack](#tech-stack) | [The Story](#the-story) | [Roadmap](#roadmap)

</div>

---

## How It Works

You push code. Dispatch handles the rest.

```
1. You trigger a rollout                →  Dispatch picks a canary node
2. Canary gets the update via Balena    →  Installs to backup partition, reboots
3. Dispatch watches health metrics      →  Sensor accuracy, latency, device vitals
4. Health checks pass for 10 minutes    →  Dispatch promotes to the next wave
5. Something breaks on any node         →  Automatic rollback, fleet paused, you get notified
6. Everything healthy                   →  Full fleet updated. Zero downtime.
```

That's it. You define the health checks, the rollout strategy, and the failure thresholds. Dispatch executes it.

## Features

> [!NOTE]
> Dispatch is in active development. Here's what's built and what's coming.

**Fleet Orchestration:**
- **Staged rollouts** — Canary node → health gate → percentage-based expansion → full fleet. Same pattern Tesla uses to ship FSD to millions of vehicles.
- **Health-gated promotions** — A canary doesn't promote until every metric you define passes for a duration you configure.
- **Automatic rollback** — Health degrades after an update? The node rolls back. No human needed.
- **Version pinning** — Pin specific nodes to specific versions while the rest of the fleet moves forward. Great for debugging.
- **A/B fleet testing** — Run version A on one group, version B on another. Compare real-world performance across your fleet.

**Observability:**
- **Real-time fleet dashboard** — Per-node status, version info, health metrics, rollout progress, manual override controls.
- **Custom health metrics** — Define what "healthy" means for your fleet. Sensor accuracy, API latency, memory pressure, inference speed — whatever matters.
- **Fleet-wide rollback** — One button. Every node. Back to the last known-good version.

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                       Dispatch                            │
│                                                           │
│  ┌─────────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │  Rollout Engine  │  │    Fleet     │  │     Web      │ │
│  │  staged deploys  │  │   Monitor    │  │  Dashboard   │ │
│  │  canary/% gates  │  │   health     │  │  (React/TS)  │ │
│  │  auto-rollback   │  │   metrics    │  │              │ │
│  └────────┬─────────┘  └──────┬───────┘  └──────┬──────┘ │
│           │                   │                  │        │
│           └─────────┬─────────┘                  │        │
│                     │                            │        │
│              ┌──────┴──────┐                     │        │
│              │  PostgreSQL  │◄────────────────────┘        │
│              │  fleet state │                              │
│              └──────┬──────┘                               │
└─────────────────────┼──────────────────────────────────────┘
                      │
               Balena API / balenaCloud
                      │
     ┌────────────────┼────────────────┐
     │                │                │
┌────┴─────┐   ┌─────┴────┐   ┌──────┴───┐
│  Node 1  │   │  Node 2  │   │  Node N  │
│ balenaOS │   │ balenaOS │   │ balenaOS │
│ your app │   │ your app │   │ your app │
└──────────┘   └──────────┘   └──────────┘
```

**Rollout Engine** — controls the deployment strategy. Talks to Balena's API to push container updates, but owns the *logic*: which node gets the update first, what metrics to watch, when to promote, when to abort.

**Fleet Monitor** — collects health metrics from every node in real time. These feed directly into the Rollout Engine's promotion and rollback decisions.

**Web Dashboard** — real-time fleet visibility. Which nodes are running which version, rollout progress, health metrics, manual override controls.

## Tech Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| Device OTA | **Balena** | Container-based fleet management, OS updates, provisioning, secure connectivity |
| Fleet Orchestration | **Python / TypeScript** | Rollout engine, health monitoring, Balena API integration |
| Node Communication | **gRPC + Protobuf** | Typed contracts, streaming, bidirectional — for inter-node coordination |
| Database | **PostgreSQL** | Fleet state, rollout history, health metric time series |
| Dashboard | **React + TypeScript** | Real-time fleet visibility and rollout controls |
| CI/CD | **GitHub Actions** | Automated builds, tests, container builds pushed to balenaCloud |

---

## The Story

I'm testing Dispatch on a fleet of three [Waveshare PiRacer](https://www.waveshare.com/piracer-ai-kit.htm) autonomous racing robots — each running a Raspberry Pi 4 with a [Slamtec RPLIDAR C1](https://www.waveshare.com/rplidar-c1.htm) LiDAR sensor and a 5MP camera.

The cars run a sensor fusion pipeline that combines 360° LiDAR point clouds with camera vision to perceive their environment. They coordinate movement as a fleet. When I push a new version of the perception algorithm, Dispatch handles the rollout — canary car first, health-gated promotion, automatic rollback if sensor fusion accuracy drops.

This is what makes the project real: a bad deploy doesn't just break a dashboard. It breaks a moving vehicle's ability to see.

### Why this fleet:

| PiRacer Fleet | Production AV/EV Equivalent |
|--------------|----------------------------|
| PiRacer fleet | Waymo robotaxis / Latitude AI test vehicles |
| Balena (device OTA) | Vehicle OTA system (SWUpdate, RAUC) |
| Dispatch rollout engine | Fleet deployment pipeline |
| Fleet health monitoring | Vehicle health / field escalation systems |
| LiDAR + camera fusion | Perception stack (the payload being deployed) |
| Multi-car coordination | Multi-agent autonomous fleet coordination |

### Research

Before building, I studied the fleet systems that run in production. Full notes in [`docs/ota-industry-research.md`](docs/ota-industry-research.md):

- **Tesla** — staged rollouts, VIN-keyed deployments, fleet-wide health monitoring
- **Rivian** — zonal architecture, consolidated update targets
- **Uptane** — Linux Foundation's automotive OTA security standard
- **aktualizr (HERE OTA Connect)** — open-source C++ Uptane client
- **Waymo Fleet Infrastructure** — fleet optimization, health monitoring, field escalation

---

## Roadmap

- [ ] Balena fleet provisioning — get all 3 PiRacers on balenaCloud
- [ ] Sensor fusion on a single car (LiDAR + camera)
- [ ] Fleet orchestration service + Balena API integration
- [ ] Staged rollout engine (canary → health gate → promote)
- [ ] Fleet health monitoring (sensor accuracy, device vitals, fusion latency)
- [ ] Automatic rollback on health degradation
- [ ] Multi-car coordination (shared sensor data, coordinated movement)
- [ ] A/B fleet testing (different versions on different nodes)
- [ ] Web dashboard with real-time fleet status
- [ ] End-to-end demo: push a new perception algorithm, canary deploy, health-gated promotion, coordinated fleet driving

## About

Built by [Dhamari Trice-Hanson](https://github.com/dhamariT) — software engineer at Hack Club, incoming CS student at Kettering University. Currently building fleet infrastructure for autonomous systems.
