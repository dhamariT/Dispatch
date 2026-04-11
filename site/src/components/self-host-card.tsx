import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export interface SelfHostCardProps {
  className?: string;
}

const features = [
  "Self-hosted Go binary — runs anywhere",
  "Works with your existing Balena account",
  "Append-only snapshots in Postgres + TimescaleDB",
  "Free and open source",
];

export function SelfHostCard({ className }: SelfHostCardProps) {
  return (
    <section
      className={cn(
        "relative overflow-hidden rounded-xl border border-active/30 bg-gradient-to-br from-active/5 via-background to-background p-8",
        className,
      )}
    >
      <div className="grid gap-6 md:grid-cols-[1.2fr_1fr] md:gap-10">
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-2">
            <span className="text-[11px] font-medium uppercase tracking-widest text-active">
              Ready for the real thing?
            </span>
          </div>
          <h2 className="text-2xl font-bold leading-tight tracking-tight text-foreground">
            Run Dispatch on your own fleet
          </h2>
          <p className="max-w-md text-sm leading-relaxed text-muted-foreground">
            This demo uses three simulated cars and a canned deploy diff. Point
            Dispatch at your Balena fleet to diff real rollouts on real
            hardware.
          </p>
          <div className="mt-2 flex items-center gap-2">
            <Button variant="default" size="lg">
              Install Dispatch
            </Button>
            <Button variant="outline" size="lg">
              Read the architecture doc
            </Button>
          </div>
        </div>

        <div className="flex flex-col justify-center gap-2 border-t border-border/60 pt-6 md:border-l md:border-t-0 md:pl-8 md:pt-0">
          <h3 className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
            What you get
          </h3>
          <ul className="flex flex-col gap-2">
            {features.map((feature) => (
              <li
                key={feature}
                className="flex items-start gap-2.5 text-sm text-foreground"
              >
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 16 16"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  aria-hidden="true"
                  className="mt-0.5 shrink-0 text-active"
                >
                  <path
                    d="M3 8l3 3 7-7"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
                {feature}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}
