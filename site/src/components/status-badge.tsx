import type { ReactNode } from "react";
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

interface StatusConfig {
  label: string;
  className: string;
  icon: ReactNode;
}

// Icons kept inline so each status can have a distinct glyph without
// pulling a dependency. Size is controlled by the badge wrapper.
const iconPlay = (
  <svg viewBox="0 0 12 12" fill="currentColor" aria-hidden="true">
    <path d="M3 2l7 4-7 4z" />
  </svg>
);
const iconDot = (
  <svg viewBox="0 0 12 12" fill="currentColor" aria-hidden="true">
    <circle cx="6" cy="6" r="3" />
  </svg>
);
const iconCheck = (
  <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M2 6l3 3 5-6" />
  </svg>
);
const iconX = (
  <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
    <path d="M3 3l6 6M9 3l-6 6" />
  </svg>
);
const iconWave = (n: number) => (
  <span className="flex h-2.5 w-2.5 items-center justify-center text-[9px] font-bold">
    {n}
  </span>
);

const statusConfig: Record<DeployStatus, StatusConfig> = {
  canary: {
    label: "Canary",
    className: "bg-warning-bg text-warning border-warning/40",
    icon: iconPlay,
  },
  soaking: {
    label: "Soaking",
    className: "bg-active-bg text-active border-active/40",
    icon: (
      <span className="relative flex h-2 w-2">
        <span className="absolute inset-0 animate-ping rounded-full bg-active/60" />
        <span className="relative h-2 w-2 rounded-full bg-active" />
      </span>
    ),
  },
  wave2: {
    label: "Wave 2",
    className: "bg-active-bg text-active border-active/40",
    icon: iconWave(2),
  },
  wave3: {
    label: "Wave 3",
    className: "bg-active-bg text-active border-active/40",
    icon: iconWave(3),
  },
  complete: {
    label: "Complete",
    className: "bg-stable-bg text-stable border-stable/40",
    icon: iconCheck,
  },
  rolledback: {
    label: "Rolled back",
    className: "bg-critical-bg text-critical border-critical/40",
    icon: iconX,
  },
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const { label, className, icon } = statusConfig[status];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium [&>svg]:h-3 [&>svg]:w-3",
        className,
      )}
    >
      {icon}
      {label}
    </span>
  );
}
