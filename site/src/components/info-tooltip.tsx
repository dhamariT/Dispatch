import type { ReactNode } from "react";
import {
  HelpPopover,
  HelpPopoverContent,
  HelpPopoverIconTrigger,
  HelpPopoverText,
  HelpPopoverTitle,
} from "./help-popover";

export interface InfoTooltipProps {
  title: ReactNode;
  message: ReactNode;
  size?: "sm" | "md";
}

export function InfoTooltip({ title, message, size = "sm" }: InfoTooltipProps) {
  return (
    <HelpPopover>
      <HelpPopoverIconTrigger size={size} />
      <HelpPopoverContent>
        <HelpPopoverTitle>{title}</HelpPopoverTitle>
        <HelpPopoverText>{message}</HelpPopoverText>
      </HelpPopoverContent>
    </HelpPopover>
  );
}
