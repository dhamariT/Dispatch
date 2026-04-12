import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";

export type ConfidenceLevel = "high" | "moderate" | "low" | "insufficient";

export interface ConfidenceBadgeProps {
  pValue: number;
  className?: string;
}

function getLevel(p: number): ConfidenceLevel {
  if (p < 0.001) return "high";
  if (p < 0.01) return "moderate";
  if (p < 0.05) return "low";
  return "insufficient";
}

const badgeVariants = cva(
  "inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium tabular-nums",
  {
    variants: {
      level: {
        high: "bg-critical-bg text-critical",
        moderate: "bg-warning-bg text-warning",
        low: "bg-muted text-muted-foreground",
        insufficient: "bg-muted text-muted-foreground/60",
      } satisfies Record<ConfidenceLevel, string>,
    },
  },
);

function formatP(p: number): string {
  if (p < 0.0001) return "p<0.0001";
  if (p < 0.001) return `p=${p.toFixed(4)}`;
  return `p=${p.toFixed(3)}`;
}

export function ConfidenceBadge({ pValue, className }: ConfidenceBadgeProps) {
  const level = getLevel(pValue);
  return (
    <span className={cn(badgeVariants({ level }), className)}>
      {formatP(pValue)}
    </span>
  );
}
