import { cn } from "@/lib/utils";

export interface DemoBannerProps {
  className?: string;
}

export function DemoBanner({ className }: DemoBannerProps) {
  return (
    <div
      role="status"
      className={cn(
        "flex items-center justify-center gap-3 border-b border-active/20 bg-active/10 px-4 py-2 text-xs",
        className,
      )}
    >
      <span className="flex h-4 w-4 shrink-0 items-center justify-center text-active">
        <svg
          width="14"
          height="14"
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.75"
          aria-hidden="true"
        >
          <circle cx="8" cy="8" r="7" />
          <path d="M8 5v3.5M8 11v.5" strokeLinecap="round" />
        </svg>
      </span>
      <span className="text-foreground">
        <span className="font-semibold">You're exploring a demo</span>
        <span className="text-muted-foreground">
          {" "}
          with simulated devices. Actions like Promote and Rollback mutate
          local state only.
        </span>
      </span>
      <span className="text-muted-foreground">·</span>
      <a
        href="https://github.com/dhamariT/Dispatch"
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-1 font-medium text-active transition-colors hover:text-active/80"
      >
        Run Dispatch on your own fleet
        <svg
          width="12"
          height="12"
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          aria-hidden="true"
        >
          <path
            d="M6 3h7v7M13 3L6 10M3 6v7h7"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </a>
    </div>
  );
}
