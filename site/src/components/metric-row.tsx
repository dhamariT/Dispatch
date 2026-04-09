import { cn } from "@/lib/utils";

export type MetricSeverity = "critical" | "warning" | "stable" | "neutral";

export interface MetricRowProps {
  name: string;
  before: number;
  after: number;
  format?: "percent" | "bytes" | "ms" | "number";
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

function getSeverity(delta: number, format: MetricRowProps["format"]): MetricSeverity {
  const abs = Math.abs(delta);
  // For percent-based metrics, small changes are noise.
  if (format === "percent") {
    if (abs >= 5) return "critical";
    if (abs >= 2) return "warning";
    if (abs > 0) return "stable";
    return "neutral";
  }
  // For everything else, use relative thresholds later.
  // Hardcoded for now.
  if (abs >= 20) return "critical";
  if (abs >= 10) return "warning";
  if (abs > 0) return "stable";
  return "neutral";
}

export function MetricRow({ name, before, after, format = "number" }: MetricRowProps) {
  const delta = after - before;
  const severity = getSeverity(delta, format);
  const sign = delta > 0 ? "+" : "";

  return (
    <tr
      className={cn(
        "border-b border-border/50",
        severity === "critical" && "bg-critical-bg",
        severity === "warning" && "bg-warning-bg",
      )}
    >
      <td className="px-4 py-2.5 text-sm font-medium">{name}</td>
      <td className="px-4 py-2.5 text-sm text-muted-foreground tabular-nums">
        {formatValue(before, format)}
      </td>
      <td className="px-4 py-2.5 text-sm text-muted-foreground tabular-nums">
        {formatValue(after, format)}
      </td>
      <td
        className={cn(
          "px-4 py-2.5 text-sm font-medium tabular-nums",
          severity === "critical" && "text-critical",
          severity === "warning" && "text-warning",
          severity === "stable" && "text-stable",
          severity === "neutral" && "text-muted-foreground",
        )}
      >
        {sign}{formatValue(delta, format)}
      </td>
    </tr>
  );
}
