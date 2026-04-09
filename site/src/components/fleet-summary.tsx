import { cn } from "@/lib/utils";

export interface FleetSummaryProps {
  total: number;
  critical: number;
  warning: number;
  stable: number;
  waiting: number;
  offline: number;
}

interface StatProps {
  count: number;
  label: string;
  className: string;
  dotClass: string;
}

function Stat({ count, label, className, dotClass }: StatProps) {
  const active = count > 0;
  return (
    <div
      className={cn(
        "flex items-center gap-2 rounded-md px-3 py-1.5 transition-opacity",
        !active && "opacity-40",
        className,
      )}
    >
      <span className={cn("h-2 w-2 rounded-full", dotClass)} />
      <span className="text-sm font-bold tabular-nums">{count}</span>
      <span className="text-xs uppercase tracking-wider">{label}</span>
    </div>
  );
}

export function FleetSummary({
  total,
  critical,
  warning,
  stable,
  waiting,
  offline,
}: FleetSummaryProps) {
  return (
    <div className="flex flex-wrap items-center gap-3 rounded-lg border border-border bg-card p-4">
      <div className="mr-2 flex items-baseline gap-2">
        <span className="text-2xl font-bold tabular-nums">{total}</span>
        <span className="text-xs uppercase tracking-wider text-muted-foreground">
          devices
        </span>
      </div>

      <div className="ml-auto flex flex-wrap gap-2">
        <Stat
          count={critical}
          label="Critical"
          className="bg-critical-bg text-critical"
          dotClass="bg-critical"
        />
        <Stat
          count={warning}
          label="Warning"
          className="bg-warning-bg text-warning"
          dotClass="bg-warning"
        />
        <Stat
          count={stable}
          label="Stable"
          className="bg-stable-bg text-stable"
          dotClass="bg-stable"
        />
        <Stat
          count={waiting}
          label="Waiting"
          className="bg-muted text-muted-foreground"
          dotClass="bg-muted-foreground/60"
        />
        <Stat
          count={offline}
          label="Offline"
          className="bg-offline-bg text-offline"
          dotClass="bg-offline"
        />
      </div>
    </div>
  );
}
