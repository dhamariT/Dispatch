import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export interface EmptyStateProps {
  message: ReactNode;
  description?: ReactNode;
  cta?: ReactNode;
  image?: ReactNode;
  className?: string;
  compact?: boolean;
}

export function EmptyState({
  message,
  description,
  cta,
  image,
  className,
  compact = false,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-4 text-center",
        compact ? "py-8" : "py-16",
        className,
      )}
    >
      {image && <div className="mb-2">{image}</div>}
      <h3 className="max-w-md text-lg font-semibold text-foreground">
        {message}
      </h3>
      {description && (
        <p className="max-w-md text-sm leading-relaxed text-muted-foreground">
          {description}
        </p>
      )}
      {cta && <div className="mt-2 flex items-center gap-2">{cta}</div>}
    </div>
  );
}
