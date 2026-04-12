import { cva } from "class-variance-authority";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export type ExpectedOutcome = "promote" | "auto_hold";

export interface ScenarioOption {
  name: string;
  description: string;
  outcome: ExpectedOutcome;
  icon: "shield" | "check" | "activity" | "gauge";
}

export interface ScenarioPickerProps {
  scenarios: ScenarioOption[];
  selected?: string | null;
  onSelect?: (name: string) => void;
  onRun?: () => void;
  running?: boolean;
}

const cardVariants = cva(
  "group relative flex items-start gap-4 rounded-xl border px-5 py-4 text-left transition-all",
  {
    variants: {
      state: {
        idle: "border-border hover:border-muted-foreground/30 hover:bg-muted/20",
        selected: "border-active bg-active-bg/40 shadow-sm shadow-active/10",
      },
    },
  },
);

const outcomeBadgeVariants = cva(
  "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider",
  {
    variants: {
      outcome: {
        promote: "bg-stable-bg text-stable",
        auto_hold: "bg-critical-bg text-critical",
      } satisfies Record<ExpectedOutcome, string>,
    },
  },
);

const iconContainerVariants = cva(
  "flex h-9 w-9 shrink-0 items-center justify-center rounded-lg transition-colors",
  {
    variants: {
      state: {
        idle: "bg-muted text-muted-foreground",
        selected: "bg-active-bg text-active",
      },
    },
  },
);

function ScenarioIcon({ icon }: { icon: ScenarioOption["icon"] }) {
  const props = {
    width: 18,
    height: 18,
    viewBox: "0 0 18 18",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.5,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true as const,
  };

  switch (icon) {
    case "shield":
      return (
        <svg {...props}>
          <path d="M9 1.5L2.5 4.5V8.5c0 4 2.7 6.7 6.5 7.5 3.8-.8 6.5-3.5 6.5-7.5V4.5L9 1.5z" />
          <path d="M6.5 9l2 2 3.5-4" />
        </svg>
      );
    case "check":
      return (
        <svg {...props}>
          <circle cx="9" cy="9" r="7" />
          <path d="M6 9l2 2 4-4" />
        </svg>
      );
    case "activity":
      return (
        <svg {...props}>
          <path d="M1.5 9h3l2-5 3 10 2-5h3" />
        </svg>
      );
    case "gauge":
      return (
        <svg {...props}>
          <path d="M9 16a7 7 0 1 1 0-14 7 7 0 0 1 0 14z" />
          <path d="M9 5v4l2.5 2.5" />
        </svg>
      );
  }
}

const checkIcon = (
  <svg
    width="16"
    height="16"
    viewBox="0 0 16 16"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
    className="text-active"
  >
    <path d="M3 8l3.5 3.5L13 5" />
  </svg>
);

export function ScenarioPicker({
  scenarios,
  selected,
  onSelect,
  onRun,
  running = false,
}: ScenarioPickerProps) {
  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-2 sm:grid-cols-2">
        {scenarios.map((s) => {
          const isSelected = selected === s.name;
          return (
            <button
              key={s.name}
              type="button"
              onClick={() => onSelect?.(s.name)}
              className={cn(cardVariants({ state: isSelected ? "selected" : "idle" }))}
            >
              <div className={cn(iconContainerVariants({ state: isSelected ? "selected" : "idle" }))}>
                <ScenarioIcon icon={s.icon} />
              </div>

              <div className="flex min-w-0 flex-1 flex-col gap-1.5">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold">
                    {s.name.replace(/_/g, " ")}
                  </span>
                  <span className={cn(outcomeBadgeVariants({ outcome: s.outcome }))}>
                    {s.outcome === "promote" ? "promotes" : "holds"}
                  </span>
                </div>
                <span className="text-xs leading-relaxed text-muted-foreground">
                  {s.description}
                </span>
              </div>

              <div className={cn(
                "flex h-5 w-5 shrink-0 items-center justify-center rounded-full border transition-all",
                isSelected
                  ? "border-active bg-active-bg"
                  : "border-border group-hover:border-muted-foreground/40",
              )}>
                {isSelected && checkIcon}
              </div>
            </button>
          );
        })}
      </div>

      <div className="flex items-center gap-3">
        <Button
          variant="default"
          size="sm"
          disabled={!selected || running}
          onClick={onRun}
        >
          {running ? (
            <span className="flex items-center gap-2">
              <span className="h-3 w-3 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent" />
              Running experiment...
            </span>
          ) : (
            "Run experiment"
          )}
        </Button>
        {selected && !running && (
          <span className="text-xs text-muted-foreground">
            Expected outcome:{" "}
            <span className={cn(
              "font-medium",
              scenarios.find((s) => s.name === selected)?.outcome === "auto_hold"
                ? "text-critical"
                : "text-stable",
            )}>
              {scenarios.find((s) => s.name === selected)?.outcome === "auto_hold"
                ? "auto-hold"
                : "promote"}
            </span>
          </span>
        )}
      </div>
    </div>
  );
}
