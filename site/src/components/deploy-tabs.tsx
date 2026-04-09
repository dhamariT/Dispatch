import { cn } from "@/lib/utils";

export type DeploySeverity = "critical" | "warning" | "stable" | "neutral";

export interface DeployTab {
  id: string;
  from: string;
  to: string;
  severity: DeploySeverity;
}

export interface DeployTabsProps {
  deploys: DeployTab[];
  selected: string;
  onSelect: (id: string) => void;
  className?: string;
}

export function DeployTabs({
  deploys,
  selected,
  onSelect,
  className,
}: DeployTabsProps) {
  return (
    <div
      className={cn(
        "flex gap-1 overflow-x-auto border-b border-border",
        className,
      )}
      role="tablist"
    >
      {deploys.map((deploy) => {
        const isSelected = deploy.id === selected;
        return (
          <button
            key={deploy.id}
            type="button"
            role="tab"
            aria-selected={isSelected}
            onClick={() => onSelect(deploy.id)}
            className={cn(
              "group relative flex shrink-0 items-center gap-2 border-b-2 px-4 py-3 text-sm font-medium transition-colors",
              isSelected
                ? "border-foreground text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            <span
              className={cn(
                "h-1.5 w-1.5 rounded-full",
                deploy.severity === "critical" && "bg-critical",
                deploy.severity === "warning" && "bg-warning",
                deploy.severity === "stable" && "bg-stable",
                deploy.severity === "neutral" && "bg-muted-foreground/40",
              )}
            />
            <span className="tabular-nums">
              {deploy.from} <span className="text-muted-foreground">→</span>{" "}
              {deploy.to}
            </span>
          </button>
        );
      })}
    </div>
  );
}
