import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ExperimentMetricsTable } from "./experiment-metrics-table";
import type { ExperimentMetricRowProps } from "./experiment-metric-row";

const lidarRegression: ExperimentMetricRowProps[] = [
  {
    name: "LiDAR accuracy",
    format: "percent",
    canaryMean: 91.1,
    canarySD: 1.3,
    canaryN: 30,
    controlMean: 98.0,
    controlSD: 0.3,
    controlN: 60,
    pValue: 0.0000001,
    effectSize: -8.92,
    verdict: "regression",
  },
  {
    name: "CPU usage",
    format: "percent",
    canaryMean: 47.0,
    canarySD: 2.3,
    canaryN: 30,
    controlMean: 32.8,
    controlSD: 2.0,
    controlN: 60,
    pValue: 0.0000001,
    effectSize: 6.83,
    verdict: "regression",
  },
  {
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
  {
    name: "Fusion latency",
    format: "ms",
    canaryMean: 14.4,
    canarySD: 1.4,
    canaryN: 30,
    controlMean: 11.9,
    controlSD: 1.0,
    controlN: 60,
    pValue: 0.0000001,
    effectSize: 2.11,
    verdict: "regression",
  },
];

const healthyDeploy: ExperimentMetricRowProps[] = [
  {
    name: "LiDAR accuracy",
    format: "percent",
    canaryMean: 97.9,
    canarySD: 0.4,
    canaryN: 30,
    controlMean: 98.0,
    controlSD: 0.4,
    controlN: 60,
    pValue: 0.71,
    effectSize: -0.28,
    verdict: "no_change",
  },
  {
    name: "CPU usage",
    format: "percent",
    canaryMean: 33.2,
    canarySD: 2.5,
    canaryN: 30,
    controlMean: 33.0,
    controlSD: 2.5,
    controlN: 60,
    pValue: 0.71,
    effectSize: -0.23,
    verdict: "no_change",
  },
  {
    name: "Memory",
    format: "bytes",
    canaryMean: 412.5,
    canarySD: 8.0,
    canaryN: 30,
    controlMean: 412.0,
    controlSD: 8.0,
    controlN: 60,
    pValue: 0.86,
    effectSize: 0.08,
    verdict: "no_change",
  },
  {
    name: "Fusion latency",
    format: "ms",
    canaryMean: 12.2,
    canarySD: 1.2,
    canaryN: 30,
    controlMean: 12.0,
    controlSD: 1.2,
    controlN: 60,
    pValue: 0.73,
    effectSize: 0.32,
    verdict: "no_change",
  },
];

const meta: Meta<typeof ExperimentMetricsTable> = {
  title: "Experiment/ExperimentMetricsTable",
  component: ExperimentMetricsTable,
};

export default meta;
type Story = StoryObj<typeof ExperimentMetricsTable>;

export const LiDARRegression: Story = {
  args: { metrics: lidarRegression },
};

export const HealthyDeploy: Story = {
  args: { metrics: healthyDeploy },
};
