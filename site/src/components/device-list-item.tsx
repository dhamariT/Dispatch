import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";
import {
  findWorstMetric,
  formatMetricValue,
  getWorstSeverity,
  type Severity,
} from "@/lib/severity";
import { StatusDot } from "./status-dot";
import type { MetricRowProps } from "./metric-row";

export type DeviceStatus = "deployed" | "waiting" | "offline";

export interface DeviceListItemProps {
  deviceId: string;
  status: DeviceStatus;
  metrics?: MetricRowProps[];
  onClick?: () => void;
  expanded?: boolean;
}

const itemVariants = cva(
  "group relative flex w-full items-center gap-4 overflow-hidden rounded-lg border bg-card px-5 py-4 text-left transition-all duration-200 hover:-translate-y-px hover:shadow-md",
  {
    variants: {
      severity: {
        critical:
          "border-critical/40 hover:border-critical/60 hover:shadow-critical/20",
        warning: "border-warning/30 hover:border-warning/50",
        stable: "border-border/60 hover:border-border",
        neutral: "border-border/60 hover:border-border",
      },
      expanded: {
        true: "ring-2 ring-active/40",
        false: "",
      },
    },
    defaultVariants: {
      severity: "neutral",
      expanded: false,
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

const deltaTextVariants = cva("text-sm font-bold tabular-nums", {
  variants: {
    severity: {
      critical: "text-critical",
      warning: "text-warning",
      stable: "text-stable",
      neutral: "text-muted-foreground",
    },
  },
});

export function DeviceListItem({
  deviceId,
  status,
  metrics,
  onClick,
  expanded,
}: DeviceListItemProps) {
  const severity: Severity = getWorstSeverity(metrics);
  const worst = findWorstMetric(metrics);
  const delta = worst ? worst.after - worst.before : 0;
  const sign = delta > 0 ? "+" : "";

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

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(itemVariants({ severity, expanded }))}
    >
      <div className={cn(edgeVariants({ severity }))} />

      <StatusDot
        variant={dotVariant}
        size="md"
        pulse={severity === "critical"}
        className="ml-1 shrink-0"
      />

      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-semibold">{deviceId}</div>
        <div
          className={cn(
            "text-xs",
            status === "deployed" && "text-muted-foreground",
            status === "waiting" && "text-muted-foreground",
            status === "offline" && "text-offline",
          )}
        >
          {status === "deployed" && "Deployed"}
          {status === "waiting" && "Waiting"}
          {status === "offline" && "Offline"}
        </div>
      </div>

      {worst && status === "deployed" && (
        <div className="hidden items-center gap-3 sm:flex">
          <div className="flex flex-col items-end">
            <span className="text-[10px] uppercase tracking-widest text-muted-foreground">
              {worst.name}
            </span>
            <span className={cn(deltaTextVariants({ severity }))}>
              {sign}
              {formatMetricValue(delta, worst.format)}
            </span>
          </div>
          {(severity === "critical" || severity === "warning") && (
            <span
              className={cn(
                "rounded px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-widest",
                severity === "critical" && "bg-critical text-background",
                severity === "warning" && "bg-warning text-background",
              )}
            >
              {severity}
            </span>
          )}
        </div>
      )}

      {metrics && metrics.length > 0 && status === "deployed" && (
        <div className="hidden text-xs tabular-nums text-muted-foreground md:block">
          {metrics.length} {metrics.length === 1 ? "metric" : "metrics"}
        </div>
      )}

      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        aria-hidden="true"
        className={cn(
          "shrink-0 text-muted-foreground transition-transform duration-200",
          expanded && "rotate-90",
        )}
      >
        <path d="M6 4l4 4-4 4" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </button>
  );
}
