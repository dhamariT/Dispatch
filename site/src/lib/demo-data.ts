import type { MetricRowProps } from "@/components/metric-row";
import { getWorstSeverity } from "./severity";

export type DeviceStatus = "deployed" | "waiting" | "offline";
export type DeploySeverity = "critical" | "warning" | "stable" | "neutral";

export interface DemoDevice {
  deviceId: string;
  status: DeviceStatus;
  metrics?: MetricRowProps[];
}

export interface DemoDeploy {
  id: string;
  from: string;
  to: string;
  severity: DeploySeverity;
  status: "canary" | "soaking" | "wave2" | "wave3" | "complete" | "rolledback";
  soakElapsed?: number;
  soakTotal?: number;
  devices: DemoDevice[];
}

export interface DemoFleet {
  name: string;
  deploys: DemoDeploy[];
}

export const demoFleet: DemoFleet = {
  name: "PiRacer Fleet (Demo)",
  deploys: [
    {
      id: "d-143",
      from: "v1.4.2",
      to: "v1.4.3",
      severity: "critical",
      status: "soaking",
      soakElapsed: 14,
      soakTotal: 30,
      devices: [
        {
          deviceId: "car-2",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 98.1, after: 91.3, format: "percent" },
            { name: "CPU usage", before: 34, after: 48, format: "percent" },
            { name: "Memory", before: 412, after: 418, format: "bytes" },
            { name: "Fusion latency", before: 12, after: 14, format: "ms" },
          ],
        },
        { deviceId: "car-1", status: "waiting" },
        { deviceId: "car-3", status: "waiting" },
      ],
    },
    {
      id: "d-142",
      from: "v1.4.1",
      to: "v1.4.2",
      severity: "stable",
      status: "complete",
      devices: [
        {
          deviceId: "car-1",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 97.8, after: 98.0, format: "percent" },
            { name: "CPU usage", before: 31, after: 33, format: "percent" },
          ],
        },
        {
          deviceId: "car-2",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 97.9, after: 98.1, format: "percent" },
            { name: "CPU usage", before: 32, after: 34, format: "percent" },
          ],
        },
        {
          deviceId: "car-3",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 97.7, after: 97.9, format: "percent" },
            { name: "CPU usage", before: 30, after: 31, format: "percent" },
          ],
        },
      ],
    },
    {
      id: "d-141",
      from: "v1.4.0",
      to: "v1.4.1",
      severity: "warning",
      status: "complete",
      devices: [
        {
          deviceId: "car-1",
          status: "deployed",
          metrics: [
            { name: "CPU usage", before: 29, after: 31, format: "percent" },
          ],
        },
        {
          deviceId: "car-2",
          status: "deployed",
          metrics: [
            { name: "CPU usage", before: 30, after: 33, format: "percent" },
          ],
        },
        {
          deviceId: "car-3",
          status: "deployed",
          metrics: [
            { name: "CPU usage", before: 28, after: 30, format: "percent" },
          ],
        },
      ],
    },
  ],
};

export function countFleetBySeverity(devices: DemoDevice[]) {
  let critical = 0;
  let warning = 0;
  let stable = 0;
  let waiting = 0;
  let offline = 0;

  for (const device of devices) {
    if (device.status === "waiting") {
      waiting++;
      continue;
    }
    if (device.status === "offline") {
      offline++;
      continue;
    }
    const severity = getWorstSeverity(device.metrics);
    if (severity === "critical") critical++;
    else if (severity === "warning") warning++;
    else stable++;
  }

  return { critical, warning, stable, waiting, offline, total: devices.length };
}
