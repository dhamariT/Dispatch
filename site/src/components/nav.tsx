import { cn } from "@/lib/utils";

export interface NavProps {
  fleetName: string;
  className?: string;
}

export function Nav({ fleetName, className }: NavProps) {
  return (
    <nav
      className={cn(
        "flex h-14 items-center justify-between border-b border-border bg-background px-6",
        className,
      )}
    >
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-foreground text-background">
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              aria-hidden="true"
            >
              <path d="M2 8l3 3 9-9" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </div>
          <span className="text-base font-bold tracking-tight">Dispatch</span>
        </div>

        <div className="flex items-center gap-2 text-sm">
          <span className="text-muted-foreground">/</span>
          <span className="font-medium">{fleetName}</span>
        </div>
      </div>

      <div className="flex items-center gap-1">
        <button
          type="button"
          className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          aria-label="Settings"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.75"
            aria-hidden="true"
          >
            <circle cx="8" cy="8" r="2" />
            <path d="M12.5 9.5l1.3.75-1 1.73-1.3-.75a4 4 0 01-1.5.87V13h-2v-1.4a4 4 0 01-1.5-.87l-1.3.75-1-1.73 1.3-.75a4 4 0 010-1.74l-1.3-.75 1-1.73 1.3.75a4 4 0 011.5-.87V3h2v1.4a4 4 0 011.5.87l1.3-.75 1 1.73-1.3.75a4 4 0 010 1.74z" />
          </svg>
        </button>
      </div>
    </nav>
  );
}
