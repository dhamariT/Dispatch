/**
 * Simulation store — the backing state for Dispatch's demo fleet.
 *
 * Why this exists: until dispatchd is production-ready, the dashboard
 * runs against a client-side simulation instead of a real API. This
 * store holds that simulation state, persists it to localStorage so
 * refresh doesn't wipe progress, and exposes actions (promote, rollback,
 * expand device) that actually mutate the state like a real backend would.
 *
 * When dispatchd ships, swap this for React Query hooks that hit /api.
 */

import { create } from "zustand";
import { persist } from "zustand/middleware";
import { demoFleet, type DemoDeploy, type DemoDevice } from "./demo-data";
import type { MetricRowProps } from "@/components/metric-row";

type Severity = "critical" | "warning" | "stable" | "neutral";

interface SimulationState {
  deploys: DemoDeploy[];
  selectedDeployId: string;
  expandedDevice: string | null;
  lastTick: number;
}

interface SimulationActions {
  selectDeploy: (id: string) => void;
  toggleExpand: (deviceId: string) => void;
  tick: () => void;
  promote: () => void;
  rollback: () => void;
}

type SimulationStore = SimulationState & SimulationActions;

export const useSimulationStore = create<SimulationStore>()(
  persist(
    (set, get) => ({
      deploys: demoFleet.deploys,
      selectedDeployId: demoFleet.deploys[0].id,
      expandedDevice: "car-2",
      lastTick: Date.now(),

      selectDeploy: (id) => set({ selectedDeployId: id }),

      toggleExpand: (deviceId) =>
        set((state) => ({
          expandedDevice: state.expandedDevice === deviceId ? null : deviceId,
        })),

      // Advances the soak timer of the currently selected deploy by 1 minute
      // each real-world second. Speeds up the demo while keeping the UX legible.
      tick: () =>
        set((state) => {
          const now = Date.now();
          const deploys: DemoDeploy[] = state.deploys.map((d): DemoDeploy => {
            if (d.id !== state.selectedDeployId) return d;
            if (
              d.status !== "soaking" ||
              d.soakElapsed === undefined ||
              d.soakTotal === undefined
            ) {
              return d;
            }
            if (d.soakElapsed >= d.soakTotal) return d;
            return {
              ...d,
              soakElapsed: Math.min(d.soakTotal, d.soakElapsed + 1),
            };
          });
          return { deploys, lastTick: now };
        }),

      promote: () =>
        set((state) => {
          const deploys: DemoDeploy[] = state.deploys.map((d): DemoDeploy => {
            if (d.id !== state.selectedDeployId) return d;
            if (d.status !== "soaking") return d;

            // Find waiting devices and move them to deployed with synthetic metrics.
            const promotedDevices: DemoDevice[] = d.devices.map(
              (device): DemoDevice => {
                if (device.status !== "waiting") return device;
                return {
                  ...device,
                  status: "deployed",
                  metrics: synthesizeHealthyMetrics(device.deviceId),
                };
              },
            );

            const waitingRemaining = promotedDevices.filter(
              (device) => device.status === "waiting",
            ).length;

            const nextStatus: DemoDeploy["status"] =
              waitingRemaining === 0 ? "complete" : "wave2";

            return {
              ...d,
              status: nextStatus,
              devices: promotedDevices,
              soakElapsed: d.soakTotal,
            };
          });
          return { deploys };
        }),

      rollback: () =>
        set((state) => {
          const deploys: DemoDeploy[] = state.deploys.map((d): DemoDeploy => {
            if (d.id !== state.selectedDeployId) return d;
            if (d.status !== "soaking" && d.status !== "wave2") return d;

            // Revert deployed devices back to waiting (like unpinning the release).
            const revertedDevices: DemoDevice[] = d.devices.map(
              (device): DemoDevice => ({
                ...device,
                status:
                  device.status === "deployed" ? "waiting" : device.status,
                metrics: undefined,
              }),
            );

            return {
              ...d,
              status: "rolledback",
              devices: revertedDevices,
            };
          });
          return { deploys };
        }),
    }),
    {
      name: "dispatch-simulation-v1",
      // Only persist state, not actions.
      partialize: (state) => ({
        deploys: state.deploys,
        selectedDeployId: state.selectedDeployId,
        expandedDevice: state.expandedDevice,
        lastTick: state.lastTick,
      }),
    },
  ),
);

// Start the tick interval once the store is loaded in the browser.
// This is a module-level singleton — it runs once per page load.
if (typeof window !== "undefined") {
  setInterval(() => {
    useSimulationStore.getState().tick();
  }, 1000);
}

// Synthesizes plausible "healthy" metrics for newly promoted devices
// so the UI has something to show when they move from waiting to deployed.
function synthesizeHealthyMetrics(deviceId: string): MetricRowProps[] {
  // Seed slight per-device variation using a cheap hash of the device ID.
  const hash = deviceId
    .split("")
    .reduce((acc, c) => acc + c.charCodeAt(0), 0);
  const jitter = (hash % 5) / 10;
  return [
    {
      name: "LiDAR accuracy",
      before: 97.8 + jitter,
      after: 97.9 + jitter,
      format: "percent",
    },
    {
      name: "CPU usage",
      before: 31,
      after: 32 + (hash % 3),
      format: "percent",
    },
    {
      name: "Memory",
      before: 408,
      after: 410 + (hash % 4),
      format: "bytes",
    },
  ];
}

// Derived selector — fetches the currently selected deploy.
export function selectCurrentDeploy(state: SimulationStore): DemoDeploy {
  return (
    state.deploys.find((d) => d.id === state.selectedDeployId) ??
    state.deploys[0]
  );
}
