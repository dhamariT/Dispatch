import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";
import { ConfidenceBadge } from "./confidence-badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export type Verdict = "regression" | "improvement" | "no_change" | "insufficient_data";
export type MetricFormat = "percent" | "bytes" | "ms" | "number";

export interface ExperimentMetricRowProps {
  name: string;
  format?: MetricFormat;
  canaryMean: number;
  canarySD: number;
  canaryN: number;
  controlMean: number;
  controlSD: number;
  controlN: number;
  pValue: number;
  effectSize: number;
  verdict: Verdict;
}

const rowVariants = cva(
  "border-b border-border/40 last:border-b-0 transition-colors relative",
  {
    variants: {
      verdict: {
        regression: "bg-critical-bg/70 hover:bg-critical-bg",
        improvement: "bg-stable-bg/50 hover:bg-stable-bg",
        no_change: "hover:bg-muted/20",
        insufficient_data: "hover:bg-muted/20",
      } satisfies Record<Verdict, string>,
    },
  },
);

const verdictLabels: Record<Verdict, string> = {
  regression: "Regression",
  improvement: "Improvement",
  no_change: "No change",
  insufficient_data: "Insufficient data",
};

const verdictVariants = cva(
  "inline-flex items-center rounded-md px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wider",
  {
    variants: {
      verdict: {
        regression: "bg-critical-bg text-critical",
        improvement: "bg-stable-bg text-stable",
        no_change: "text-muted-foreground",
        insufficient_data: "text-muted-foreground/60 italic",
      } satisfies Record<Verdict, string>,
    },
  },
);

function formatValue(value: number, format: MetricFormat): string {
  switch (format) {
    case "percent":
      return `${value.toFixed(1)}%`;
    case "bytes":
      return `${value.toFixed(0)}MB`;
    case "ms":
      return `${value.toFixed(1)}ms`;
    default:
      return value.toFixed(1);
  }
}

function formatSD(value: number, format: MetricFormat): string {
  switch (format) {
    case "percent":
      return `${value.toFixed(1)}%`;
    case "bytes":
      return `${value.toFixed(0)}MB`;
    case "ms":
      return `${value.toFixed(1)}ms`;
    default:
      return value.toFixed(1);
  }
}

function formatEffectSize(d: number): string {
  const abs = Math.abs(d);
  if (abs < 0.2) return "negligible";
  if (abs < 0.5) return "small";
  if (abs < 0.8) return "medium";
  return "large";
}

export function ExperimentMetricRow({
  name,
  format = "number",
  canaryMean,
  canarySD,
  canaryN,
  controlMean,
  controlSD,
  controlN,
  pValue,
  effectSize,
  verdict,
}: ExperimentMetricRowProps) {
  const diff = canaryMean - controlMean;
  const sign = diff > 0 ? "+" : "";
  const showAccent = verdict === "regression";

  return (
    <tr className={cn(rowVariants({ verdict }))}>
      <td className="relative px-5 py-4 text-sm font-medium">
        {showAccent && (
          <span className="absolute inset-y-0 left-0 w-[3px] bg-critical" />
        )}
        {name}
      </td>

      <td className="px-5 py-4 text-sm tabular-nums">
        <Tooltip>
          <TooltipTrigger
            render={(props) => (
              <span {...props} className="cursor-default">
                {formatValue(controlMean, format)}{" "}
                <span className="text-muted-foreground">
                  ± {formatSD(controlSD, format)}
                </span>
              </span>
            )}
          />
          <TooltipContent side="top">
            <div className="text-xs">
              n={controlN} samples from control group
            </div>
          </TooltipContent>
        </Tooltip>
      </td>

      <td className={cn(
        "px-5 py-4 text-sm tabular-nums",
        verdict === "regression" && "font-semibold text-critical",
      )}>
        <Tooltip>
          <TooltipTrigger
            render={(props) => (
              <span {...props} className="cursor-default">
                {formatValue(canaryMean, format)}{" "}
                <span className={cn(
                  verdict === "regression" ? "text-critical/70" : "text-muted-foreground"
                )}>
                  ± {formatSD(canarySD, format)}
                </span>
              </span>
            )}
          />
          <TooltipContent side="top">
            <div className="text-xs">
              n={canaryN} samples from canary group
            </div>
          </TooltipContent>
        </Tooltip>
      </td>

      <td className={cn(
        "px-5 py-4 text-right text-sm font-bold tabular-nums",
        verdict === "regression" && "text-critical",
        verdict === "improvement" && "text-stable",
      )}>
        {sign}{formatValue(diff, format)}
      </td>

      <td className="px-5 py-4">
        {verdict !== "insufficient_data" && (
          <Tooltip>
            <TooltipTrigger
              render={(props) => (
                <span {...props} className="cursor-default">
                  <ConfidenceBadge pValue={pValue} />
                </span>
              )}
            />
            <TooltipContent side="top">
              <div className="flex flex-col gap-1 text-xs">
                <span>Effect size: d={effectSize.toFixed(2)} ({formatEffectSize(effectSize)})</span>
                <span>Welch&apos;s t-test, unequal variance</span>
              </div>
            </TooltipContent>
          </Tooltip>
        )}
      </td>

      <td className="px-5 py-4">
        <span className={cn(verdictVariants({ verdict }))}>
          {verdictLabels[verdict]}
        </span>
      </td>
    </tr>
  );
}
