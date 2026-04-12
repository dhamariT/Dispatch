import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ExperimentVerdict } from "./experiment-verdict";

const meta: Meta<typeof ExperimentVerdict> = {
  title: "Experiment/ExperimentVerdict",
  component: ExperimentVerdict,
};

export default meta;
type Story = StoryObj<typeof ExperimentVerdict>;

export const Collecting: Story = {
  args: {
    status: "collecting",
  },
};

export const Analyzing: Story = {
  args: {
    status: "analyzing",
    metricCount: 4,
  },
};

export const SafeToPromote: Story = {
  args: {
    status: "decided",
    decision: "promote",
    metricCount: 4,
  },
};

export const AutoHold: Story = {
  args: {
    status: "decided",
    decision: "auto_hold",
    regressionCount: 3,
    metricCount: 4,
    holdReason:
      "regression detected: [lidar_accuracy (p=0.0000, d=-8.92), cpu_usage (p=0.0000, d=6.83), fusion_latency_ms (p=0.0000, d=2.11)]",
  },
};

export const ManualHold: Story = {
  args: {
    status: "decided",
    decision: "hold",
    holdReason: "Holding for manual review — LiDAR numbers look suspicious even though stats say no change.",
  },
};

export const AutoHoldSingleMetric: Story = {
  args: {
    status: "decided",
    decision: "auto_hold",
    regressionCount: 1,
    metricCount: 4,
    holdReason: "regression detected: [cpu_usage (p=0.0000, d=0.89)]",
  },
};
