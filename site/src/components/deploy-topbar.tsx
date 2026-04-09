import { Button } from "@/components/ui/button";
import { StatusBadge, type DeployStatus } from "./status-badge";
import { StatusDot } from "./status-dot";
import { Topbar, TopbarData, TopbarDivider } from "./topbar";

export interface DeployTopbarProps {
  from: string;
  to: string;
  status: DeployStatus;
  soakElapsed?: number;
  soakTotal?: number;
  critical: number;
  warning: number;
  waiting: number;
  offline: number;
  total: number;
  onPromote?: () => void;
  onRollback?: () => void;
  showActions?: boolean;
}

export function DeployTopbar({
  from,
  to,
  status,
  soakElapsed,
  soakTotal,
  critical,
  warning,
  waiting,
  offline,
  total,
  onPromote,
  onRollback,
  showActions = true,
}: DeployTopbarProps) {
  const soakPct =
    soakElapsed !== undefined && soakTotal !== undefined && soakTotal > 0
      ? Math.min(100, (soakElapsed / soakTotal) * 100)
      : null;

  return (
    <Topbar>
      <TopbarData
        label="Deploy"
        value={
          <span className="tabular-nums">
            {from} <span className="text-muted-foreground">→</span> {to}
          </span>
        }
      />

      <TopbarDivider />

      <StatusBadge status={status} />

      {soakPct !== null && (
        <>
          <TopbarDivider />
          <div className="flex items-center gap-2">
            <span className="text-xs uppercase tracking-wider text-muted-foreground">
              Soak
            </span>
            <div className="h-1 w-20 overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-active transition-all"
                style={{ width: `${soakPct}%` }}
              />
            </div>
            <span className="text-sm font-medium tabular-nums">
              {soakElapsed}m / {soakTotal}m
            </span>
          </div>
        </>
      )}

      <div className="ml-auto flex items-center gap-2">
        <div className="hidden items-center gap-2 md:flex">
          <FleetStat count={critical} label="Critical" variant="critical" />
          {warning > 0 && (
            <FleetStat count={warning} label="Warning" variant="warning" />
          )}
          {waiting > 0 && (
            <FleetStat count={waiting} label="Waiting" variant="waiting" />
          )}
          {offline > 0 && (
            <FleetStat count={offline} label="Offline" variant="offline" />
          )}
          <FleetStat count={total} label="Total" variant="total" />
        </div>

        {showActions && status === "soaking" && (
          <>
            <TopbarDivider />
            <Button variant="outline" size="sm" onClick={onRollback}>
              Rollback
            </Button>
            <Button variant="default" size="sm" onClick={onPromote}>
              Promote
            </Button>
          </>
        )}
      </div>
    </Topbar>
  );
}

type FleetStatVariant =
  | "critical"
  | "warning"
  | "waiting"
  | "offline"
  | "total";

interface FleetStatProps {
  count: number;
  label: string;
  variant: FleetStatVariant;
}

function FleetStat({ count, label, variant }: FleetStatProps) {
  const dimmed = variant !== "total" && count === 0;

  return (
    <div
      className={`flex items-center gap-1.5 px-2 ${dimmed ? "opacity-40" : ""}`}
    >
      {variant !== "total" && (
        <StatusDot
          variant={
            variant === "critical"
              ? "critical"
              : variant === "warning"
                ? "warning"
                : variant === "waiting"
                  ? "waiting"
                  : "offline"
          }
          size="xs"
          pulse={variant === "critical" && count > 0}
        />
      )}
      <span className="text-sm font-bold tabular-nums">{count}</span>
      <span className="text-xs uppercase tracking-wider text-muted-foreground">
        {label}
      </span>
    </div>
  );
}
