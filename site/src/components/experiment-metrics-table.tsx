import { InfoTooltip } from "./info-tooltip";
import { ExperimentMetricRow, type ExperimentMetricRowProps } from "./experiment-metric-row";

export interface ExperimentMetricsTableProps {
  metrics: ExperimentMetricRowProps[];
}

export function ExperimentMetricsTable({ metrics }: ExperimentMetricsTableProps) {
  return (
    <div className="overflow-hidden rounded-lg border border-border/60">
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-border/60 bg-muted/20 text-left text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
            <th className="px-5 py-3 font-medium">Metric</th>
            <th className="px-5 py-3 font-medium">
              <div className="flex items-center gap-1.5">
                Control
                <InfoTooltip
                  title="Control group"
                  message="Devices running the previous release. Their metrics serve as the baseline for comparison."
                />
              </div>
            </th>
            <th className="px-5 py-3 font-medium">
              <div className="flex items-center gap-1.5">
                Canary
                <InfoTooltip
                  title="Canary group"
                  message="Devices running the new release. Compared against the control group over the same time window."
                />
              </div>
            </th>
            <th className="px-5 py-3 text-right font-medium">Diff</th>
            <th className="px-5 py-3 font-medium">
              <div className="flex items-center gap-1.5">
                Confidence
                <InfoTooltip
                  title="Statistical confidence"
                  message="p-value from Welch's t-test. Lower values mean higher confidence that the difference is real, not noise."
                />
              </div>
            </th>
            <th className="px-5 py-3 font-medium">
              <div className="flex items-center gap-1.5">
                Verdict
                <InfoTooltip
                  title="Verdict"
                  message="Requires both statistical significance (p < 0.05) and meaningful effect size (Cohen's d > 0.5) to flag a regression."
                />
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          {metrics.map((m) => (
            <ExperimentMetricRow key={m.name} {...m} />
          ))}
        </tbody>
      </table>
    </div>
  );
}
