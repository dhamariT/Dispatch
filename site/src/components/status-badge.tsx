import { cn } from "@/lib/utils";

export type DeployStatus =
  | "canary"
  | "soaking"
  | "wave2"
  | "wave3"
  | "complete"
  | "rolledback";

export interface StatusBadgeProps {
  status: DeployStatus;
}

const statusConfig: Record<
  DeployStatus,
  { label: string; className: string }
> = {
  canary: {
    label: "Canary",
    className: "bg-active-bg text-active border-active/40",
  },
  soaking: {
    label: "Soaking",
    className: "bg-active-bg text-active border-active/40",
  },
  wave2: {
    label: "Wave 2",
    className: "bg-active-bg text-active border-active/40",
  },
  wave3: {
    label: "Wave 3",
    className: "bg-active-bg text-active border-active/40",
  },
  complete: {
    label: "Complete",
    className: "bg-stable-bg text-stable border-stable/40",
  },
  rolledback: {
    label: "Rolled back",
    className: "bg-critical-bg text-critical border-critical/40",
  },
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const { label, className } = statusConfig[status];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium",
        className,
      )}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {label}
    </span>
  );
}
