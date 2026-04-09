import { cn } from "@/lib/utils";

export interface DemoBannerProps {
  onConnect?: () => void;
  className?: string;
}

export function DemoBanner({ onConnect, className }: DemoBannerProps) {
  return (
    <div
      className={cn(
        "flex items-center justify-between gap-4 border-b border-warning/30 bg-warning-bg px-6 py-3",
        className,
      )}
    >
      <div className="flex items-center gap-3">
        <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-warning/20 text-warning">
          <svg
            width="14"
            height="14"
            viewBox="0 0 16 16"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <path d="M8 1v8M8 12v1" strokeLinecap="round" />
            <circle cx="8" cy="8" r="7" />
          </svg>
        </span>
        <div className="flex flex-col">
          <span className="text-sm font-semibold text-warning">
            This is a demo fleet
          </span>
          <span className="text-xs text-warning/80">
            Connect your Balena fleet to start tracking your real devices
          </span>
        </div>
      </div>
      <button
        type="button"
        onClick={onConnect}
        className="shrink-0 rounded-md border border-warning/40 bg-warning/10 px-3 py-1.5 text-xs font-semibold text-warning transition-colors hover:bg-warning/20"
      >
        Connect Balena
      </button>
    </div>
  );
}
