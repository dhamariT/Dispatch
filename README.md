## Welcome to OTA Fleet Deployer
**A fleet orchestration platform for coordinated autonomous RC cars with LiDAR + camera sensor fusion — currently in early development.**

I'm building a multi-vehicle autonomous fleet where three PiRacer AI robots coordinate their movements using sensor fusion (LiDAR + camera), and a fleet management layer keeps their software in sync as I iterate. Balena handles the device-level OTA. I'm building the layer on top: staged rollouts, fleet health monitoring, canary deployments, and the orchestration that decides which car gets which software, when, and what to do when something goes wrong.

The hard part isn't pushing bits to a device — Balena already does that. The hard part is managing a fleet of moving robots that need to stay coordinated while their software evolves underneath them. That's what this project is about.

This is a learning project. I'm transitioning from web/backend (Ruby on Rails, TypeScript, Docker) into fleet infrastructure engineering and autonomous systems. Everything here is being built as I learn.

<!-- TODO: Add fleet photo/gif here once hardware is assembled -->
<!-- ![Fleet Demo](docs/assets/fleet-demo.gif) -->

<div align="center">

**Navigation**

[Current Status](#current-status) | [The Big Picture](#the-big-picture) | [Architecture](#architecture) | [Planned Features](#planned-features) | [Hardware](#hardware) | [Tech Stack](#tech-stack) | [Industry Research](#industry-research) | [Milestones](#milestones)

</div>

---

## Current Status

> [!NOTE]
> **This project is in the research and architecture phase.** No application code has been written yet. I've completed industry research on how production AV/EV fleet systems work (Tesla, Rivian, Uptane, aktualizr, Balena) and I'm designing the fleet orchestration layer.

**What's done:**
- Industry research on Tesla, Rivian, Uptane, HERE OTA Connect, and Balena — see [`docs/ota-industry-research.md`](docs/ota-industry-research.md)
- Fleet architecture design (Balena for device OTA, custom orchestration on top)
- Hardware acquired (3x PiRacer AI Kits + Raspberry Pi 4s + D500 LiDAR)
- Career research into Fleet Infrastructure Engineering roles (Waymo, Latitude AI, Zoox)
- This README

**What's next:**
- Balena fleet provisioning (get all 3 PiRacers on balenaCloud)
- Fleet orchestration service scaffolding
- First coordinated movement test across the fleet
- Sensor fusion pipeline (LiDAR + camera) on a single car

## The Big Picture

This project has two layers that work together:

**Layer 1 — The Autonomy Stack (the cars):**
Three PiRacer AI robots, each with a Raspberry Pi 4, a 5MP camera, and a D500 LiDAR sensor. Each car runs a sensor fusion pipeline that combines LiDAR point clouds with camera vision to understand its environment. The cars coordinate their movements — they're not just three independent robots, they're a fleet that works together.

**Layer 2 — The Fleet Orchestration Platform (what I'm building):**
As I develop and improve the autonomy stack, I need a way to push updates to the fleet safely. Not just "deploy new code" — I need staged rollouts where one car gets the new LiDAR fusion algorithm first, I watch its perception accuracy for 10 minutes, and if it stays healthy, the update rolls out to the other two. If something breaks, automatic rollback. If I'm testing two different approaches to path planning, I need to be able to run version A on car 1 and version B on cars 2 and 3.

**Balena handles the plumbing** — OS updates, container deployment, device provisioning, secure connectivity. I'm not rebuilding that.

**I'm building the brain** — the fleet orchestration logic that decides *what* gets deployed *where* and *when*, monitors health across the fleet, and coordinates rollouts so the cars stay in sync while their software evolves.

### Why this matters:

> [!TIP]
> * **This is what Fleet Infrastructure Engineers actually do.** Waymo's Fleet Infrastructure team builds the systems that manage and optimize a large fleet of driverless vehicles. That's this project, miniaturized.
> * **The cars are real and they move.** Sensor fusion, coordinated movement, real LiDAR data. Not a simulation, not LEDs blinking on a desk.
> * **The orchestration problems are real.** Canary deployments, health-gated rollouts, version pinning, fleet-wide rollback — these are production fleet management problems at any scale.

### Why this matters for AV/EV:

A bad rollout on my PiRacer fleet means three RC cars collide or lose sensor fusion. A bad rollout at Waymo means robotaxis stop picking up passengers in San Francisco. The orchestration patterns are identical — staged deployment, health monitoring, automatic rollback, fleet coordination. The stakes just scale up.

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│              Fleet Orchestration Platform                  │
│                                                            │
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
└─────────────────────┤────────────────────────────────────┘
                      │
               Balena API / balenaCloud
                      │
     ┌────────────────┤────────────────┐
     │                │                │
┌────┴─────┐   ┌─────┴────┐   ┌──────┴───┐
│ PiRacer 1│   │ PiRacer 2│   │ PiRacer 3│
│ balenaOS │   │ balenaOS │   │ balenaOS │
│ LiDAR    │   │ LiDAR    │   │ LiDAR    │
│ Camera   │   │ Camera   │   │ Camera   │
│ Fusion   │   │ Fusion   │   │ Fusion   │
│ Movement │   │ Movement │   │ Movement │
└──────────┘   └──────────┘   └───────────┘
     🏎️              🏎️              🏎️
        ← coordinated movement →
```

**The platform has three main components:**

The **Rollout Engine** manages staged deployments across the fleet. It talks to Balena's API to push container updates, but it controls the *strategy*: which car gets the update first (canary), what health metrics to watch, when to promote to the next car, and when to abort and roll back. Think of it as the deployment brain sitting on top of Balena's deployment muscle.

The **Fleet Monitor** collects health metrics from every car in real time — sensor fusion accuracy, LiDAR point cloud quality, camera frame rate, movement coordination status, CPU/memory/temperature. These metrics feed into the Rollout Engine's promotion and rollback decisions.

The **Web Dashboard** (React + TypeScript) gives real-time visibility into the fleet — which cars are running which version, rollout progress, health metrics, and manual override controls. This is the part I already know how to build from my web background.

## Planned Features

> [!IMPORTANT]
> None of these are implemented yet — this is the target feature set.
>
> **Fleet Orchestration:**
> * **Staged rollouts** — Canary car first → health gate → percentage-based expansion → full fleet. Automatic pause if any health metric crosses a threshold. Same pattern Tesla uses to roll out FSD updates.
> * **Health-gated promotions** — A canary deployment doesn't promote until the car's sensor fusion accuracy, LiDAR quality, and movement coordination all pass configurable checks for a configurable duration.
> * **Automatic rollback** — If a car's health degrades after an update, roll it back to the previous version automatically. No human needed.
> * **Version pinning** — Pin a specific car to a specific version for testing while the rest of the fleet moves forward.
> * **A/B fleet testing** — Run version A on car 1 and version B on cars 2-3 to compare sensor fusion approaches.
>
> **Sensor Fusion & Coordination:**
> * **LiDAR + camera fusion** — Combine D500 LiDAR 360° point clouds with 5MP camera data for environment perception.
> * **Multi-car coordination** — Cars share their fused sensor data and coordinate movement to avoid collisions and cover areas together.
>
> **Observability:**
> * **Real-time fleet dashboard** — Per-car status, version info, health metrics, rollout controls.
> * **Fleet health metrics** — Sensor accuracy, fusion latency, movement precision, device vitals (CPU, memory, temp).

## Hardware

This isn't a simulation — the fleet is real hardware I already own.

### The Fleet: 3x Waveshare PiRacer AI Kits

| Spec | Detail |
|------|--------|
| Platform | [Waveshare PiRacer AI Kit](https://www.waveshare.com/piracer-ai-kit.htm) (SKU: 17674) |
| Compute | Raspberry Pi 4 Model B per car |
| Camera | 5MP onboard camera (supports DonkeyCar autonomous driving) |
| Motors | Dual high-power metal DC motors, 4-wheel drive |
| Display | 0.91" OLED (128x32) — shows IP, memory, power status |
| Power | 3x 18650 Li-ion batteries with onboard protection circuit |
| Dimensions | 240mm x 196mm x 130mm |
| Storage | 32GB microSD per car — managed by balenaOS |

### Sensors

| Sensor | Model | Purpose |
|--------|-------|---------|
| LiDAR | [Waveshare D500 LiDAR Kit](https://www.waveshare.com/d500-lidar-kit.htm) | 360° DTOF scanning, 12m range, 5000 measurements/sec — environment perception |
| Camera | 5MP (onboard PiRacer) | Visual perception — fused with LiDAR for multi-modal environment understanding |

### Control Plane

| Component | Purpose |
|-----------|---------|
| Apple M4 Pro (24GB) | Development — write code, iterate, test |
| Dell Latitude 5544 (Ubuntu Server) | Always-on host — fleet orchestration service + PostgreSQL + dashboard |
| Local Wi-Fi network | Fleet communication |

## Tech Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| Device OTA | **Balena (balenaOS + balenaCloud)** | Container-based fleet management, handles OS updates, provisioning, secure connectivity |
| Fleet Orchestration | **TypeScript** | Rollout engine, health monitoring, Balena API integration |
| Sensor Fusion | **C++ / Python** | LiDAR + camera processing, performance-critical perception pipeline |
| Coordination | **gRPC + Protobuf** | Inter-car communication for movement coordination |
| Database | **PostgreSQL** | Fleet state, rollout history, health metrics |
| Dashboard | **React + TypeScript** | Real-time fleet visibility — my existing strength |
| Testing | **pytest** | Unit and integration tests for orchestration and fusion |
| CI/CD | **GitHub Actions** | Automated builds, tests, container builds pushed to balenaCloud |

## Industry Research

Before writing any code, I studied both the OTA systems and the fleet management platforms used in production. Full research notes live in [`docs/ota-industry-research.md`](docs/ota-industry-research.md), covering:

- **Tesla** — staged rollouts, VIN-keyed deployments, fleet-wide health monitoring
- **Rivian** — zonal ECU architecture, consolidated update targets
- **Uptane** — the Linux Foundation's AV/EV OTA security standard
- **aktualizr (HERE OTA Connect)** — open-source C++ Uptane client, library+daemon pattern
- **Balena** — container-based IoT fleet management, the device-level OTA layer I'm building on top of
- **Waymo Fleet Infrastructure** — fleet optimization, health monitoring, field escalation

### How this maps to real fleet infrastructure:

| This Project | Production AV/EV Equivalent |
|-------------|----------------------------|
| PiRacer fleet (3 cars) | Waymo's robotaxi fleet / Latitude AI test vehicles |
| Balena (device OTA) | Vehicle OTA system (SWUpdate, RAUC, proprietary) |
| Rollout Engine (staged deploys) | Fleet deployment pipeline (what Waymo Fleet Infra builds) |
| Fleet Monitor (health metrics) | Vehicle health monitoring / field escalation systems |
| Sensor fusion (LiDAR + camera) | Perception stack updates (the payload being deployed) |
| Canary → health gate → promote | Phased AV fleet deployment (how Tesla ships FSD) |
| Multi-car coordination | Multi-agent coordination in autonomous fleets |
| Web dashboard | Fleet operations console (Waymo Fleet Monitoring & Platform) |

### Target career: Fleet Infrastructure Engineering

This project directly maps to roles like:

- **Software Engineer, Fleet Infrastructure** (Waymo) — builds systems that manage and optimize a large fleet of driverless vehicles
- **Software Engineer, Fleet Monitoring & Platform** (Waymo) — internal tooling and monitoring for fleet operations
- **Software Engineer, Deploy** (Latitude AI / Ford) — ships software to Ford's autonomous vehicle fleet
- **Robotics Platform Engineer** — fleet orchestration for warehouse robots, drones, autonomous vehicles
