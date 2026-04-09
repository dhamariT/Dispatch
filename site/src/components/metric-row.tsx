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
  if (format === "percent") {
    if (abs >= 5) return "critical";
    if (abs >= 2) return "warning";
    if (abs > 0) return "stable";
    return "neutral";
  }
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
        "border-b border-border/40 last:border-b-0 transition-colors",
        severity === "critical" && "bg-critical-bg/60 hover:bg-critical-bg",
        severity === "warning" && "bg-warning-bg/50 hover:bg-warning-bg",
        severity === "stable" && "hover:bg-muted/30",
        severity === "neutral" && "hover:bg-muted/30",
      )}
    >
      <td className="px-5 py-4 text-sm font-medium">{name}</td>
      <td className="px-5 py-4 text-sm text-muted-foreground tabular-nums">
        {formatValue(before, format)}
      </td>
      <td
        className={cn(
          "px-5 py-4 text-sm tabular-nums",
          severity === "critical" && "font-semibold text-critical",
          severity === "warning" && "font-semibold text-warning",
          severity !== "critical" && severity !== "warning" && "text-foreground",
        )}
      >
        {formatValue(after, format)}
      </td>
      <td
        className={cn(
          "px-5 py-4 text-right text-base font-bold tabular-nums",
          severity === "critical" && "text-critical",
          severity === "warning" && "text-warning",
          severity === "stable" && "text-stable",
          severity === "neutral" && "text-muted-foreground",
        )}
      >
        {sign}
        {formatValue(delta, format)}
      </td>
    </tr>
  );
}
