import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

export function Topbar({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      className={cn(
        "flex h-14 items-center gap-2 border-b border-border bg-background/80 px-4 backdrop-blur-sm",
        className,
      )}
      {...props}
    />
  );
}

export function TopbarData({
  icon,
  label,
  value,
  className,
}: {
  icon?: ReactNode;
  label?: ReactNode;
  value: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex items-center gap-1.5 px-2 py-1 text-sm",
        className,
      )}
    >
      {icon && (
        <span className="flex h-4 w-4 shrink-0 items-center justify-center text-muted-foreground [&>svg]:size-full">
          {icon}
        </span>
      )}
      {label && (
        <span className="text-xs uppercase tracking-wider text-muted-foreground">
          {label}
        </span>
      )}
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}

export function TopbarDivider({ className }: { className?: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn("h-5 w-px bg-border", className)}
    />
  );
}
