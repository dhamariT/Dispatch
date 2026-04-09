import { cn } from "@/lib/utils";

export interface SoakTimerProps {
  elapsed: number;
  total: number;
}

export function SoakTimer({ elapsed, total }: SoakTimerProps) {
  const pct = Math.min(100, Math.max(0, (elapsed / total) * 100));
  const complete = elapsed >= total;

  return (
    <div className="flex flex-col gap-1.5 w-48">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">Soak</span>
        <span className="tabular-nums text-foreground">
          {elapsed}m / {total}m
        </span>
      </div>
      <div className="h-1 w-full overflow-hidden rounded-full bg-muted">
        <div
          className={cn(
            "h-full rounded-full transition-all",
            complete ? "bg-stable" : "bg-active",
          )}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
