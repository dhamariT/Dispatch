import type { ComponentProps, ReactNode } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

export function HelpPopover(props: ComponentProps<typeof Popover>) {
  return <Popover {...props} />;
}

export function HelpPopoverTrigger(props: ComponentProps<typeof PopoverTrigger>) {
  return <PopoverTrigger {...props} />;
}

export function HelpPopoverIcon({ className }: { className?: string }) {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      aria-hidden="true"
      className={className}
    >
      <circle cx="8" cy="8" r="7" />
      <path d="M6 6a2 2 0 114 0c0 1-1 1.5-2 2v.5M8 11v.5" strokeLinecap="round" />
    </svg>
  );
}

export interface HelpPopoverIconTriggerProps
  extends ComponentProps<typeof PopoverTrigger> {
  size?: "sm" | "md";
}

export function HelpPopoverIconTrigger({
  size = "sm",
  className,
  ...props
}: HelpPopoverIconTriggerProps) {
  return (
    <PopoverTrigger
      render={(triggerProps) => (
        <button
          type="button"
          aria-label="More info"
          {...triggerProps}
          className={cn(
            "inline-flex shrink-0 items-center justify-center rounded-full text-muted-foreground/70 transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
            size === "sm" && "h-4 w-4",
            size === "md" && "h-5 w-5",
            className,
          )}
        >
          <HelpPopoverIcon className={size === "md" ? "h-4 w-4" : "h-3.5 w-3.5"} />
        </button>
      )}
      {...props}
    />
  );
}

export function HelpPopoverContent({
  className,
  ...props
}: ComponentProps<typeof PopoverContent>) {
  return (
    <PopoverContent
      side="bottom"
      align="start"
      className={cn("w-80 gap-3 p-4", className)}
      {...props}
    />
  );
}

export function HelpPopoverTitle({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <h4
      className={cn(
        "text-sm font-semibold leading-tight text-foreground",
        className,
      )}
    >
      {children}
    </h4>
  );
}

export function HelpPopoverText({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <p
      className={cn(
        "text-xs leading-relaxed text-muted-foreground",
        className,
      )}
    >
      {children}
    </p>
  );
}

export function HelpPopoverLink({
  href,
  children,
  className,
}: {
  href: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className={cn(
        "inline-flex items-center gap-1.5 text-xs font-medium text-active transition-colors hover:text-active/80",
        className,
      )}
    >
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
      {children}
    </a>
  );
}

export function HelpPopoverLinksGroup({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex flex-col gap-1.5 border-t border-border/60 pt-3",
        className,
      )}
    >
      {children}
    </div>
  );
}
