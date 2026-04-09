// Semantic color roles for Dispatch.
// These map to deploy states and metric diffs — the core
// visual language of the dashboard.

export type RoleColors = {
  background: string;
  border: string;
  text: string;
  fill: string;
};

export type Role =
  | "critical"
  | "warning"
  | "stable"
  | "neutral"
  | "active"
  | "offline";

// Light mode role colors.
export const lightRoles: Record<Role, RoleColors> = {
  // Metric regression — sensor accuracy dropped, something broke.
  critical: {
    background: "oklch(0.95 0.04 25)",
    border: "oklch(0.7 0.18 25)",
    text: "oklch(0.45 0.2 25)",
    fill: "oklch(0.577 0.245 27.325)",
  },
  // Elevated but not broken — CPU spike, memory pressure.
  warning: {
    background: "oklch(0.96 0.04 85)",
    border: "oklch(0.75 0.15 85)",
    text: "oklch(0.45 0.15 70)",
    fill: "oklch(0.7 0.18 75)",
  },
  // Metric within normal range — no regression.
  stable: {
    background: "oklch(0.96 0.03 155)",
    border: "oklch(0.75 0.12 155)",
    text: "oklch(0.4 0.15 155)",
    fill: "oklch(0.6 0.17 155)",
  },
  // No change, baseline, informational.
  neutral: {
    background: "oklch(0.97 0 0)",
    border: "oklch(0.88 0 0)",
    text: "oklch(0.45 0 0)",
    fill: "oklch(0.7 0 0)",
  },
  // Deploy in progress, soak timer running.
  active: {
    background: "oklch(0.95 0.03 250)",
    border: "oklch(0.7 0.12 250)",
    text: "oklch(0.4 0.15 250)",
    fill: "oklch(0.55 0.2 255)",
  },
  // Device not reporting.
  offline: {
    background: "oklch(0.95 0 0)",
    border: "oklch(0.8 0 0)",
    text: "oklch(0.55 0 0)",
    fill: "oklch(0.65 0 0)",
  },
};

// Dark mode role colors.
export const darkRoles: Record<Role, RoleColors> = {
  critical: {
    background: "oklch(0.25 0.06 25)",
    border: "oklch(0.5 0.18 25)",
    text: "oklch(0.8 0.15 25)",
    fill: "oklch(0.65 0.22 27)",
  },
  warning: {
    background: "oklch(0.25 0.04 85)",
    border: "oklch(0.5 0.12 85)",
    text: "oklch(0.82 0.12 75)",
    fill: "oklch(0.65 0.16 80)",
  },
  stable: {
    background: "oklch(0.22 0.03 155)",
    border: "oklch(0.45 0.1 155)",
    text: "oklch(0.8 0.12 155)",
    fill: "oklch(0.6 0.15 155)",
  },
  neutral: {
    background: "oklch(0.22 0 0)",
    border: "oklch(0.35 0 0)",
    text: "oklch(0.7 0 0)",
    fill: "oklch(0.5 0 0)",
  },
  active: {
    background: "oklch(0.22 0.04 250)",
    border: "oklch(0.45 0.12 250)",
    text: "oklch(0.8 0.12 255)",
    fill: "oklch(0.55 0.18 255)",
  },
  offline: {
    background: "oklch(0.2 0 0)",
    border: "oklch(0.32 0 0)",
    text: "oklch(0.55 0 0)",
    fill: "oklch(0.4 0 0)",
  },
};
