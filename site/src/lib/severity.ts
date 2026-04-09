export type Severity = "critical" | "warning" | "stable" | "neutral";

export type MetricFormat = "percent" | "bytes" | "ms" | "number";

export interface Metric {
  name: string;
  before: number;
  after: number;
  format?: MetricFormat;
}

export const SEVERITY_THRESHOLDS = {
  percent: { critical: 5, warning: 2 },
  number: { critical: 20, warning: 10 },
} as const;

export function getMetricSeverity(
  before: number,
  after: number,
  format: MetricFormat = "number",
): Severity {
  const abs = Math.abs(after - before);
  const t =
    format === "percent" ? SEVERITY_THRESHOLDS.percent : SEVERITY_THRESHOLDS.number;
  if (abs >= t.critical) return "critical";
  if (abs >= t.warning) return "warning";
  if (abs > 0) return "stable";
  return "neutral";
}

export function getWorstSeverity(metrics: Metric[] | undefined): Severity {
  if (!metrics || metrics.length === 0) return "neutral";
  let worst: Severity = "neutral";
  for (const m of metrics) {
    const s = getMetricSeverity(m.before, m.after, m.format);
    if (s === "critical") return "critical";
    if (s === "warning") worst = "warning";
    else if (s === "stable" && worst === "neutral") worst = "stable";
  }
  return worst;
}

export function findWorstMetric(metrics: Metric[] | undefined): Metric | null {
  if (!metrics || metrics.length === 0) return null;
  let worst = metrics[0];
  let worstAbs = Math.abs(metrics[0].after - metrics[0].before);
  for (const m of metrics) {
    const abs = Math.abs(m.after - m.before);
    if (abs > worstAbs) {
      worst = m;
      worstAbs = abs;
    }
  }
  return worst;
}

export function countBySeverity(metrics: Metric[] | undefined) {
  let critical = 0;
  let warning = 0;
  let stable = 0;
  if (!metrics) return { critical, warning, stable };
  for (const m of metrics) {
    const s = getMetricSeverity(m.before, m.after, m.format);
    if (s === "critical") critical++;
    else if (s === "warning") warning++;
    else if (s === "stable") stable++;
  }
  return { critical, warning, stable };
}

export function getThresholdText(format: MetricFormat = "number"): string {
  const t =
    format === "percent" ? SEVERITY_THRESHOLDS.percent : SEVERITY_THRESHOLDS.number;
  const unit =
    format === "percent"
      ? "%"
      : format === "bytes"
        ? "MB"
        : format === "ms"
          ? "ms"
          : "";
  return `Critical: ±${t.critical}${unit} · Warning: ±${t.warning}${unit}`;
}

export function formatMetricValue(
  value: number,
  format: MetricFormat = "number",
): string {
  switch (format) {
    case "percent":
      return `${value}%`;
    case "bytes":
      return `${value}MB`;
    case "ms":
      return `${value}ms`;
    default:
      return String(value);
  }
}
