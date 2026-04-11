import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";
import {
  countBySeverity,
  getWorstSeverity,
  type Severity,
} from "@/lib/severity";
import { InfoTooltip } from "./info-tooltip";
import { MetricRow, type MetricRowProps } from "./metric-row";
import { StatusDot } from "./status-dot";

export type DeviceStatus = "deployed" | "waiting" | "offline";

export interface DeviceGroupProps {
  deviceId: string;
  status: DeviceStatus;
  metrics?: MetricRowProps[];
  /** When true, only render the metrics table — no card chrome, no
   * device header. For use inside an already-expanded list item that
   * already shows the device ID and status. */
  tableOnly?: boolean;
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
      } satisfies Record<Severity, string>,
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
    } satisfies Record<Severity, string>,
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
      } satisfies Record<DeviceStatus, string>,
    },
  },
);

interface DeviceMetricsTableProps {
  metrics: MetricRowProps[];
}

export function DeviceMetricsTable({ metrics }: DeviceMetricsTableProps) {
  return (
    <div className="overflow-hidden rounded-lg border border-border/60">
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-border/60 bg-muted/20 text-left text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
            <th className="px-5 py-3 font-medium">Metric</th>
            <th className="px-5 py-3 font-medium">
              <div className="flex items-center gap-1.5">
                Before
                <InfoTooltip
                  title="Before"
                  message="The metric value captured just before the deploy was triggered."
                />
              </div>
            </th>
            <th className="px-5 py-3 font-medium">
              <div className="flex items-center gap-1.5">
                After
                <InfoTooltip
                  title="After"
                  message="The metric value captured after the soak period completed."
                />
              </div>
            </th>
            <th className="px-5 py-3 text-right font-medium">
              <div className="flex items-center justify-end gap-1.5">
                Delta
                <InfoTooltip
                  title="Delta"
                  message="The difference between after and before. Red means the metric regressed past the critical threshold, amber means warning."
                />
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          {metrics.map((m) => (
            <MetricRow key={m.name} {...m} />
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function DeviceGroup({
  deviceId,
  status,
  metrics,
  tableOnly = false,
}: DeviceGroupProps) {
  const hasMetrics = status === "deployed" && metrics && metrics.length > 0;
  const severity: Severity = hasMetrics ? getWorstSeverity(metrics) : "neutral";

  if (tableOnly) {
    return hasMetrics ? <DeviceMetricsTable metrics={metrics} /> : null;
  }

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

  const counts = hasMetrics ? countBySeverity(metrics) : null;

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

        {severity === "critical" && counts && (
          <div className="flex items-center gap-2 rounded-md bg-critical-bg px-3 py-1.5 text-sm font-semibold text-critical">
            <StatusDot variant="critical" size="xs" pulse />
            {counts.critical} critical{" "}
            {counts.critical === 1 ? "regression" : "regressions"}
          </div>
        )}
        {severity === "warning" && counts && (
          <div className="flex items-center gap-2 rounded-md bg-warning-bg px-3 py-1.5 text-sm font-semibold text-warning">
            <StatusDot variant="warning" size="xs" />
            {counts.warning} above threshold
          </div>
        )}
      </header>

      {hasMetrics && <DeviceMetricsTable metrics={metrics} />}
    </section>
  );
}
