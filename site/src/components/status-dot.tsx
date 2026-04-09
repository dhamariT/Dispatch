import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const dotVariants = cva("inline-block rounded-full border-solid", {
  variants: {
    variant: {
      critical: "bg-critical border-critical/20",
      warning: "bg-warning border-warning/20",
      stable: "bg-stable border-stable/20",
      active: "bg-active border-active/20",
      neutral: "bg-muted-foreground border-muted-foreground/20",
      offline: "bg-offline border-offline/20",
      waiting: "bg-transparent border-muted-foreground/60",
    },
    size: {
      xs: "size-2 border-2",
      sm: "size-2.5 border-[3px]",
      md: "size-3 border-4",
      lg: "size-4 border-4",
    },
    pulse: {
      true: "animate-pulse",
      false: "",
    },
  },
  defaultVariants: {
    variant: "neutral",
    size: "sm",
    pulse: false,
  },
});

export interface StatusDotProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof dotVariants> {}

export function StatusDot({
  variant,
  size,
  pulse,
  className,
  ...props
}: StatusDotProps) {
  return (
    <span
      className={cn(dotVariants({ variant, size, pulse }), className)}
      {...props}
    />
  );
}
