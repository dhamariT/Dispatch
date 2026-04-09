import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

export interface FleetSummaryProps {
  total: number;
  critical: number;
  warning: number;
  stable: number;
  waiting: number;
  offline: number;
}

const statVariants = cva(
  "flex items-center gap-2 rounded-md px-3 py-1.5 transition-opacity",
  {
    variants: {
      variant: {
        critical: "bg-critical-bg text-critical",
        warning: "bg-warning-bg text-warning",
        stable: "bg-stable-bg text-stable",
        waiting: "bg-muted text-muted-foreground",
        offline: "bg-offline-bg text-offline",
      },
      active: {
        true: "",
        false: "opacity-40",
      },
    },
  },
);

const dotVariants = cva("h-2 w-2 rounded-full", {
  variants: {
    variant: {
      critical: "bg-critical",
      warning: "bg-warning",
      stable: "bg-stable",
      waiting: "bg-muted-foreground/60",
      offline: "bg-offline",
    },
  },
});

interface StatProps extends VariantProps<typeof statVariants> {
  count: number;
  label: string;
  variant: NonNullable<VariantProps<typeof statVariants>["variant"]>;
}

function Stat({ count, label, variant }: StatProps) {
  const active = count > 0;
  return (
    <div className={cn(statVariants({ variant, active }))}>
      <span className={cn(dotVariants({ variant }))} />
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
        <Stat count={critical} label="Critical" variant="critical" />
        <Stat count={warning} label="Warning" variant="warning" />
        <Stat count={stable} label="Stable" variant="stable" />
        <Stat count={waiting} label="Waiting" variant="waiting" />
        <Stat count={offline} label="Offline" variant="offline" />
      </div>
    </div>
  );
}
