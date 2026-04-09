import { cn } from "@/lib/utils";
import { MetricRow, type MetricRowProps } from "./metric-row";

export type DeviceStatus = "deployed" | "waiting" | "offline";

export interface DeviceGroupProps {
  deviceId: string;
  status: DeviceStatus;
  metrics?: MetricRowProps[];
}

const statusConfig: Record<
  DeviceStatus,
  { label: string; dotClass: string; textClass: string }
> = {
  deployed: {
    label: "Deployed",
    dotClass: "bg-active shadow-[0_0_0_3px_var(--active-bg)]",
    textClass: "text-active",
  },
  waiting: {
    label: "Waiting",
    dotClass: "bg-transparent ring-2 ring-muted-foreground/50",
    textClass: "text-muted-foreground",
  },
  offline: {
    label: "Offline",
    dotClass: "bg-offline",
    textClass: "text-offline",
  },
};

// Figure out the worst metric severity in the group so we can
// color the left edge and make regressions visually unmissable.
function worstSeverity(metrics: MetricRowProps[]): "critical" | "warning" | "stable" | "neutral" {
  let worst: "critical" | "warning" | "stable" | "neutral" = "neutral";
  for (const m of metrics) {
    const delta = Math.abs(m.after - m.before);
    const isPct = m.format === "percent";
    const critical = isPct ? delta >= 5 : delta >= 20;
    const warning = isPct ? delta >= 2 : delta >= 10;
    if (critical) return "critical";
    if (warning && worst !== "critical") worst = "warning";
    if (delta > 0 && worst === "neutral") worst = "stable";
  }
  return worst;
}

export function DeviceGroup({ deviceId, status, metrics }: DeviceGroupProps) {
  const { label, dotClass, textClass } = statusConfig[status];
  const hasMetrics = status === "deployed" && metrics && metrics.length > 0;
  const severity = hasMetrics ? worstSeverity(metrics) : "neutral";

  const edgeClass = {
    critical: "bg-critical",
    warning: "bg-warning",
    stable: "bg-stable",
    neutral: "bg-border",
  }[severity];

  return (
    <section
      className={cn(
        "relative flex flex-col gap-5 overflow-hidden rounded-xl border bg-card p-6",
        severity === "critical" && "border-critical/40 shadow-lg shadow-critical/10",
        severity === "warning" && "border-warning/40",
        severity !== "critical" && severity !== "warning" && "border-border",
      )}
    >
      {/* Colored edge — instant visual severity indicator. */}
      <div className={cn("absolute inset-y-0 left-0 w-1", edgeClass)} />

      <header className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className={cn("h-2.5 w-2.5 rounded-full", dotClass)} />
          <h3 className="text-xl font-semibold tracking-tight">{deviceId}</h3>
          <span
            className={cn(
              "rounded-md px-2 py-0.5 text-xs font-medium uppercase tracking-wider",
              status === "deployed" && "bg-active-bg text-active",
              status === "waiting" && "bg-muted text-muted-foreground",
              status === "offline" && "bg-offline-bg text-offline",
            )}
          >
            {label}
          </span>
        </div>

        {hasMetrics && severity === "critical" && (
          <div className="flex items-center gap-2 rounded-md bg-critical-bg px-3 py-1.5 text-sm font-semibold text-critical">
            <span className="h-2 w-2 animate-pulse rounded-full bg-critical" />
            Regression detected
          </div>
        )}
        {hasMetrics && severity === "warning" && (
          <div className="flex items-center gap-2 rounded-md bg-warning-bg px-3 py-1.5 text-sm font-semibold text-warning">
            Needs attention
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
                <th className="px-5 py-3 font-medium text-right">Delta</th>
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
