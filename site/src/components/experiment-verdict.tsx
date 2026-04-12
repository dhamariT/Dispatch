import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";
import { StatusDot } from "./status-dot";

export type ExperimentDecision = "promote" | "hold" | "auto_hold";
export type ExperimentStatus = "collecting" | "analyzing" | "decided";

export interface ExperimentVerdictProps {
  status: ExperimentStatus;
  decision?: ExperimentDecision | null;
  holdReason?: string | null;
  regressionCount?: number;
  metricCount?: number;
  onPromote?: () => void;
  onHold?: () => void;
}

const containerVariants = cva(
  "flex flex-col gap-3 rounded-xl border p-5",
  {
    variants: {
      state: {
        collecting: "border-active/40 bg-active-bg/30",
        analyzing: "border-active/40 bg-active-bg/30",
        promote: "border-stable/40 bg-stable-bg/30",
        auto_hold: "border-critical/40 bg-critical-bg/30",
        hold: "border-warning/40 bg-warning-bg/30",
      },
    },
  },
);

const iconShield = (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.5"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
  >
    <path d="M10 2L3 5.5V10c0 4.5 3 7.5 7 8.5 4-1 7-4 7-8.5V5.5L10 2z" />
  </svg>
);

const iconCheck = (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
  >
    <path d="M4 10l4 4 8-8" />
  </svg>
);

const iconPause = (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    aria-hidden="true"
  >
    <path d="M7 4v12M13 4v12" />
  </svg>
);

export function ExperimentVerdict({
  status,
  decision,
  holdReason,
  regressionCount = 0,
  metricCount = 0,
}: ExperimentVerdictProps) {
  if (status === "collecting") {
    return (
      <div className={cn(containerVariants({ state: "collecting" }))}>
        <div className="flex items-center gap-3">
          <StatusDot variant="active" size="md" pulse />
          <div>
            <h3 className="text-sm font-semibold">Collecting metrics</h3>
            <p className="text-xs text-muted-foreground">
              Gathering samples from canary and control groups.
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (status === "analyzing") {
    return (
      <div className={cn(containerVariants({ state: "analyzing" }))}>
        <div className="flex items-center gap-3">
          <StatusDot variant="active" size="md" pulse />
          <div>
            <h3 className="text-sm font-semibold">Analyzing</h3>
            <p className="text-xs text-muted-foreground">
              Running statistical comparison across {metricCount} metrics.
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (decision === "promote") {
    return (
      <div className={cn(containerVariants({ state: "promote" }))}>
        <div className="flex items-center gap-3">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-stable-bg text-stable">
            {iconCheck}
          </span>
          <div>
            <h3 className="text-sm font-semibold text-stable">Safe to promote</h3>
            <p className="text-xs text-muted-foreground">
              No statistically significant regressions detected across{" "}
              {metricCount} metrics.
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (decision === "auto_hold") {
    return (
      <div className={cn(containerVariants({ state: "auto_hold" }))}>
        <div className="flex items-center gap-3">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-critical-bg text-critical">
            {iconShield}
          </span>
          <div>
            <h3 className="text-sm font-semibold text-critical">
              Automatically held
            </h3>
            <p className="text-xs text-muted-foreground">
              {regressionCount} of {metricCount} metrics regressed with
              statistical significance.
            </p>
          </div>
        </div>
        {holdReason && (
          <div className="rounded-lg bg-critical-bg/50 px-4 py-3 font-mono text-xs text-critical/90">
            {holdReason}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className={cn(containerVariants({ state: "hold" }))}>
      <div className="flex items-center gap-3">
        <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-warning-bg text-warning">
          {iconPause}
        </span>
        <div>
          <h3 className="text-sm font-semibold text-warning">
            Manually held
          </h3>
          {holdReason && (
            <p className="text-xs text-muted-foreground">{holdReason}</p>
          )}
        </div>
      </div>
    </div>
  );
}
