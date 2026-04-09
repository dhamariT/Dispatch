import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";
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

type Severity = "critical" | "warning" | "stable" | "neutral";

function getSeverity(metrics: MetricRowProps[] | undefined): Severity {
  if (!metrics || metrics.length === 0) return "neutral";
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

function findWorstMetric(
  metrics: MetricRowProps[] | undefined,
): MetricRowProps | null {
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

function formatValue(value: number, format: MetricRowProps["format"]): string {
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
  const severity = getSeverity(metrics);
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
        <div className="hidden items-baseline gap-2 sm:flex">
          <span className="text-xs text-muted-foreground">{worst.name}</span>
          <span className={cn(deltaTextVariants({ severity }))}>
            {sign}
            {formatValue(delta, worst.format)}
          </span>
        </div>
      )}

      {metrics && metrics.length > 0 && status === "deployed" && (
        <div className="text-xs tabular-nums text-muted-foreground">
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
