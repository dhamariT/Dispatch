import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ExperimentMetricRow } from "./experiment-metric-row";

const meta: Meta<typeof ExperimentMetricRow> = {
  title: "Experiment/ExperimentMetricRow",
  component: ExperimentMetricRow,
  decorators: [
    (Story) => (
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-border text-left text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
            <th className="px-5 py-3 font-medium">Metric</th>
            <th className="px-5 py-3 font-medium">Control</th>
            <th className="px-5 py-3 font-medium">Canary</th>
            <th className="px-5 py-3 text-right font-medium">Diff</th>
            <th className="px-5 py-3 font-medium">Confidence</th>
            <th className="px-5 py-3 font-medium">Verdict</th>
          </tr>
        </thead>
        <tbody>
          <Story />
        </tbody>
      </table>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof ExperimentMetricRow>;

export const Regression: Story = {
  args: {
    name: "LiDAR accuracy",
    format: "percent",
    canaryMean: 91.1,
    canarySD: 1.3,
    canaryN: 30,
    controlMean: 98.0,
    controlSD: 0.3,
    controlN: 60,
    pValue: 0.00000000000000000000000771,
    effectSize: -8.92,
    verdict: "regression",
  },
};

export const Improvement: Story = {
  args: {
    name: "Fusion latency",
    format: "ms",
    canaryMean: 10.2,
    canarySD: 0.8,
    canaryN: 30,
    controlMean: 12.0,
    controlSD: 1.0,
    controlN: 60,
    pValue: 0.00001,
    effectSize: -2.1,
    verdict: "improvement",
  },
};

export const NoChange: Story = {
  args: {
    name: "Memory",
    format: "bytes",
    canaryMean: 415.5,
    canarySD: 7.8,
    canaryN: 30,
    controlMean: 411.9,
    controlSD: 7.4,
    controlN: 60,
    pValue: 0.003,
    effectSize: 0.48,
    verdict: "no_change",
  },
};

export const InsufficientData: Story = {
  args: {
    name: "Sensor temp",
    format: "number",
    canaryMean: 42.0,
    canarySD: 3.0,
    canaryN: 3,
    controlMean: 41.5,
    controlSD: 2.8,
    controlN: 4,
    pValue: 0.8,
    effectSize: 0.17,
    verdict: "insufficient_data",
  },
};

export const FullTable: Story = {
  decorators: [
    (Story) => (
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-border text-left text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
            <th className="px-5 py-3 font-medium">Metric</th>
            <th className="px-5 py-3 font-medium">Control</th>
            <th className="px-5 py-3 font-medium">Canary</th>
            <th className="px-5 py-3 text-right font-medium">Diff</th>
            <th className="px-5 py-3 font-medium">Confidence</th>
            <th className="px-5 py-3 font-medium">Verdict</th>
          </tr>
        </thead>
        <tbody>
          <Story />
        </tbody>
      </table>
    ),
  ],
  render: () => (
    <>
      <ExperimentMetricRow
        name="LiDAR accuracy"
        format="percent"
        canaryMean={91.1}
        canarySD={1.3}
        canaryN={30}
        controlMean={98.0}
        controlSD={0.3}
        controlN={60}
        pValue={0.0000001}
        effectSize={-8.92}
        verdict="regression"
      />
      <ExperimentMetricRow
        name="CPU usage"
        format="percent"
        canaryMean={47.0}
        canarySD={2.3}
        canaryN={30}
        controlMean={32.8}
        controlSD={2.0}
        controlN={60}
        pValue={0.0000001}
        effectSize={6.83}
        verdict="regression"
      />
      <ExperimentMetricRow
        name="Memory"
        format="bytes"
        canaryMean={415.5}
        canarySD={7.8}
        canaryN={30}
        controlMean={411.9}
        controlSD={7.4}
        controlN={60}
        pValue={0.003}
        effectSize={0.48}
        verdict="no_change"
      />
      <ExperimentMetricRow
        name="Fusion latency"
        format="ms"
        canaryMean={14.4}
        canarySD={1.4}
        canaryN={30}
        controlMean={11.9}
        controlSD={1.0}
        controlN={60}
        pValue={0.0000001}
        effectSize={2.11}
        verdict="regression"
      />
    </>
  ),
};
