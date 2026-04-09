import { cva } from "class-variance-authority";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

export type MetricSeverity = "critical" | "warning" | "stable" | "neutral";

export interface MetricRowProps {
  name: string;
  before: number;
  after: number;
  format?: "percent" | "bytes" | "ms" | "number";
}

const THRESHOLDS = {
  percent: { critical: 5, warning: 2 },
  number: { critical: 20, warning: 10 },
} as const;

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

function getSeverity(
  delta: number,
  format: MetricRowProps["format"],
): MetricSeverity {
  const abs = Math.abs(delta);
  const t = format === "percent" ? THRESHOLDS.percent : THRESHOLDS.number;
  if (abs >= t.critical) return "critical";
  if (abs >= t.warning) return "warning";
  if (abs > 0) return "stable";
  return "neutral";
}

function getThresholdText(format: MetricRowProps["format"]): string {
  const t = format === "percent" ? THRESHOLDS.percent : THRESHOLDS.number;
  const unit = format === "percent" ? "%" : format === "bytes" ? "MB" : format === "ms" ? "ms" : "";
  return `Critical: ±${t.critical}${unit} · Warning: ±${t.warning}${unit}`;
}

const rowVariants = cva("border-b border-border/40 last:border-b-0 transition-colors relative", {
  variants: {
    severity: {
      critical: "bg-critical-bg/70 hover:bg-critical-bg",
      warning: "bg-warning-bg/50 hover:bg-warning-bg",
      stable: "hover:bg-muted/20",
      neutral: "hover:bg-muted/20",
    },
  },
});

const deltaVariants = cva(
  "px-5 py-4 text-right text-base font-bold tabular-nums",
  {
    variants: {
      severity: {
        critical: "text-critical",
        warning: "text-warning",
        stable: "text-stable",
        neutral: "text-muted-foreground",
      },
    },
  },
);

const afterValueVariants = cva("px-5 py-4 text-sm tabular-nums", {
  variants: {
    severity: {
      critical: "font-semibold text-critical",
      warning: "font-semibold text-warning",
      stable: "text-foreground",
      neutral: "text-foreground",
    },
  },
});

export function MetricRow({
  name,
  before,
  after,
  format = "number",
}: MetricRowProps) {
  const delta = after - before;
  const severity = getSeverity(delta, format);
  const sign = delta > 0 ? "+" : "";
  const showAccent = severity === "critical" || severity === "warning";

  const tooltipLabel =
    severity === "critical"
      ? `Critical regression · ${getThresholdText(format)}`
      : severity === "warning"
        ? `Above warning threshold · ${getThresholdText(format)}`
        : severity === "stable"
          ? "Within normal range"
          : "No change";

  return (
    <TooltipProvider delayDuration={200}>
      <Tooltip>
        <TooltipTrigger asChild>
          <tr className={cn(rowVariants({ severity }))}>
            <td className="relative px-5 py-4 text-sm font-medium">
              {showAccent && (
                <span
                  className={cn(
                    "absolute inset-y-0 left-0 w-[3px]",
                    severity === "critical" && "bg-critical",
                    severity === "warning" && "bg-warning",
                  )}
                />
              )}
              {name}
            </td>
            <td className="px-5 py-4 text-sm text-muted-foreground tabular-nums">
              {formatValue(before, format)}
            </td>
            <td className={cn(afterValueVariants({ severity }))}>
              {formatValue(after, format)}
            </td>
            <td className={cn(deltaVariants({ severity }))}>
              {sign}
              {formatValue(delta, format)}
            </td>
          </tr>
        </TooltipTrigger>
        <TooltipContent side="top" align="end">
          <div className="text-xs font-medium">{tooltipLabel}</div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
