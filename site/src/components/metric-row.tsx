import { cva } from "class-variance-authority";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import {
  formatMetricValue,
  getMetricSeverity,
  getThresholdText,
  type MetricFormat,
  type Severity,
} from "@/lib/severity";

export type MetricRowProps = {
  name: string;
  before: number;
  after: number;
  format?: MetricFormat;
};

const rowVariants = cva(
  "border-b border-border/40 last:border-b-0 transition-colors relative",
  {
    variants: {
      severity: {
        critical: "bg-critical-bg/70 hover:bg-critical-bg",
        warning: "bg-warning-bg/50 hover:bg-warning-bg",
        stable: "hover:bg-muted/20",
        neutral: "hover:bg-muted/20",
      } satisfies Record<Severity, string>,
    },
  },
);

const deltaVariants = cva(
  "px-5 py-4 text-right text-base font-bold tabular-nums",
  {
    variants: {
      severity: {
        critical: "text-critical",
        warning: "text-warning",
        stable: "text-stable",
        neutral: "text-muted-foreground",
      } satisfies Record<Severity, string>,
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
    } satisfies Record<Severity, string>,
  },
});

export function MetricRow({
  name,
  before,
  after,
  format = "number",
}: MetricRowProps) {
  const delta = after - before;
  const severity = getMetricSeverity(before, after, format);
  const sign = delta > 0 ? "+" : "";
  const showAccent = severity === "critical" || severity === "warning";

  const tooltipLabel =
    severity === "critical"
      ? `Critical regression · ${getThresholdText(format)}`
      : `Above warning threshold · ${getThresholdText(format)}`;

  return (
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
        <span className="inline-flex items-center gap-2">
          {name}
          {showAccent && (
            <Tooltip>
              <TooltipTrigger
                render={(props) => (
                  <button
                    type="button"
                    {...props}
                    className={cn(
                      "inline-flex h-4 w-4 items-center justify-center rounded-full transition-opacity hover:opacity-100",
                      severity === "critical" &&
                        "text-critical/70 hover:text-critical",
                      severity === "warning" &&
                        "text-warning/70 hover:text-warning",
                    )}
                  >
                    <svg
                      width="14"
                      height="14"
                      viewBox="0 0 16 16"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      aria-hidden="true"
                    >
                      <circle cx="8" cy="8" r="6.5" />
                      <path d="M8 7v4M8 4.5v.5" strokeLinecap="round" />
                    </svg>
                    <span className="sr-only">Threshold info</span>
                  </button>
                )}
              />
              <TooltipContent side="top" align="start">
                <div className="text-xs font-medium">{tooltipLabel}</div>
              </TooltipContent>
            </Tooltip>
          )}
        </span>
      </td>
      <td className="px-5 py-4 text-sm text-muted-foreground tabular-nums">
        {formatMetricValue(before, format)}
      </td>
      <td className={cn(afterValueVariants({ severity }))}>
        {formatMetricValue(after, format)}
      </td>
      <td className={cn(deltaVariants({ severity }))}>
        {sign}
        {formatMetricValue(delta, format)}
      </td>
    </tr>
  );
}
