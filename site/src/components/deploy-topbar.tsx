import { Button } from "@/components/ui/button";
import { InfoTooltip } from "./info-tooltip";
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
      <div className="flex items-center gap-1.5 px-2 py-1 text-sm">
        <span className="text-xs uppercase tracking-wider text-muted-foreground">
          Deploy
        </span>
        <InfoTooltip
          title="What is a deploy?"
          message="A deploy is a new release rolled out to your fleet. Dispatch snapshots metrics before and after to show you what changed on each device."
        />
        <span className="font-medium tabular-nums text-foreground">
          {from} <span className="text-muted-foreground">→</span> {to}
        </span>
      </div>

      <TopbarDivider />

      <div className="flex items-center gap-1.5">
        <StatusBadge status={status} />
        <InfoTooltip
          title={statusHelpTitle(status)}
          message={statusHelpMessage(status)}
        />
      </div>

      {soakPct !== null && (
        <>
          <TopbarDivider />
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1">
              <span className="text-xs uppercase tracking-wider text-muted-foreground">
                Soak
              </span>
              <InfoTooltip
                title="Soak period"
                message="The wait period after a deploy where Dispatch collects post-deploy metrics. Once complete, you can promote to the next wave."
              />
            </div>
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

function statusHelpTitle(status: DeployStatus): string {
  switch (status) {
    case "canary":
      return "Canary stage";
    case "soaking":
      return "Soaking";
    case "wave2":
      return "Wave 2";
    case "wave3":
      return "Wave 3";
    case "complete":
      return "Complete";
    case "rolledback":
      return "Rolled back";
  }
}

function statusHelpMessage(status: DeployStatus): string {
  switch (status) {
    case "canary":
      return "The canary is the first device to receive this deploy. Dispatch is watching for metric regressions before promoting to the next wave.";
    case "soaking":
      return "The canary has been deployed. Dispatch is collecting post-deploy metrics during the soak period. Promote becomes available once soak completes.";
    case "wave2":
      return "Wave 2 is the second batch of devices to receive this deploy. It runs after the canary soak completes successfully.";
    case "wave3":
      return "Wave 3 is the third batch of devices. It runs after Wave 2 completes successfully.";
    case "complete":
      return "Every targeted device has received this deploy and all soak periods completed without regressions.";
    case "rolledback":
      return "This deploy was rolled back. Affected devices have been reverted to the previous release version via Balena.";
  }
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
