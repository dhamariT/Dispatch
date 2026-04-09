import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";
import { MetricRow, type MetricRowProps } from "./metric-row";
import { StatusDot } from "./status-dot";

export type DeviceStatus = "deployed" | "waiting" | "offline";

export interface DeviceGroupProps {
  deviceId: string;
  status: DeviceStatus;
  metrics?: MetricRowProps[];
}

type Severity = "critical" | "warning" | "stable" | "neutral";

function worstSeverity(metrics: MetricRowProps[]): Severity {
  let worst: Severity = "neutral";
  for (const m of metrics) {
    const delta = Math.abs(m.after - m.before);
    const isPct = m.format === "percent";
    if (isPct ? delta >= 5 : delta >= 20) return "critical";
    if (isPct ? delta >= 2 : delta >= 10) {
      if (worst !== "critical") worst = "warning";
    } else if (delta > 0 && worst === "neutral") {
      worst = "stable";
    }
  }
  return worst;
}

const sectionVariants = cva(
  "relative flex flex-col gap-5 overflow-hidden rounded-xl border bg-card p-6 transition-shadow",
  {
    variants: {
      severity: {
        critical: "border-critical/40 shadow-lg shadow-critical/10",
        warning: "border-warning/40",
        stable: "border-border",
        neutral: "border-border",
      },
    },
    defaultVariants: {
      severity: "neutral",
    },
  },
);

const edgeVariants = cva("absolute inset-y-0 left-0 w-1", {
  variants: {
    severity: {
      critical: "bg-critical",
      warning: "bg-warning",
      stable: "bg-stable",
      neutral: "bg-border",
    },
  },
});

const statusPillVariants = cva(
  "rounded-md px-2 py-0.5 text-xs font-medium uppercase tracking-wider",
  {
    variants: {
      status: {
        deployed: "bg-active-bg text-active",
        waiting: "bg-muted text-muted-foreground",
        offline: "bg-offline-bg text-offline",
      },
    },
  },
);

export function DeviceGroup({ deviceId, status, metrics }: DeviceGroupProps) {
  const hasMetrics = status === "deployed" && metrics && metrics.length > 0;
  const severity: Severity = hasMetrics ? worstSeverity(metrics) : "neutral";

  const dotVariant =
    status === "deployed"
      ? severity === "critical"
        ? "critical"
        : severity === "warning"
          ? "warning"
          : severity === "stable"
            ? "stable"
            : "active"
      : status === "waiting"
        ? "waiting"
        : "offline";

  const statusLabel =
    status === "deployed"
      ? "Deployed"
      : status === "waiting"
        ? "Waiting"
        : "Offline";

  const criticalCount = hasMetrics
    ? metrics.filter((m) => {
        const d = Math.abs(m.after - m.before);
        return m.format === "percent" ? d >= 5 : d >= 20;
      }).length
    : 0;

  const warningCount = hasMetrics
    ? metrics.filter((m) => {
        const d = Math.abs(m.after - m.before);
        const isPct = m.format === "percent";
        const crit = isPct ? d >= 5 : d >= 20;
        const warn = isPct ? d >= 2 : d >= 10;
        return !crit && warn;
      }).length
    : 0;

  return (
    <section className={cn(sectionVariants({ severity }))}>
      <div className={cn(edgeVariants({ severity }))} />

      <header className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <StatusDot
            variant={dotVariant}
            size="md"
            pulse={severity === "critical"}
          />
          <h3 className="text-xl font-semibold tracking-tight">{deviceId}</h3>
          <span className={cn(statusPillVariants({ status }))}>
            {statusLabel}
          </span>
        </div>

        {severity === "critical" && (
          <div className="flex items-center gap-2 rounded-md bg-critical-bg px-3 py-1.5 text-sm font-semibold text-critical">
            <StatusDot variant="critical" size="xs" pulse />
            {criticalCount} critical {criticalCount === 1 ? "regression" : "regressions"}
          </div>
        )}
        {severity === "warning" && (
          <div className="flex items-center gap-2 rounded-md bg-warning-bg px-3 py-1.5 text-sm font-semibold text-warning">
            <StatusDot variant="warning" size="xs" />
            {warningCount} above threshold
          </div>
        )}
      </header>

      {hasMetrics && (
        <div className="overflow-hidden rounded-lg border border-border/60">
          <table className="w-full border-collapse">
            <thead>
              <tr className="border-b border-border/60 bg-muted/20 text-left text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
                <th className="px-5 py-3 font-medium">Metric</th>
                <th className="px-5 py-3 font-medium">Before</th>
                <th className="px-5 py-3 font-medium">After</th>
                <th className="px-5 py-3 text-right font-medium">Delta</th>
              </tr>
            </thead>
            <tbody>
              {metrics.map((m) => (
                <MetricRow key={m.name} {...m} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
