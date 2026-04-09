import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { MetricRow } from "./metric-row";

const meta: Meta<typeof MetricRow> = {
  title: "Components/MetricRow",
  component: MetricRow,
  decorators: [
    (Story) => (
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-border text-left text-xs text-muted-foreground uppercase tracking-wider">
            <th className="px-4 py-2">Metric</th>
            <th className="px-4 py-2">Before</th>
            <th className="px-4 py-2">After</th>
            <th className="px-4 py-2">Delta</th>
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
type Story = StoryObj<typeof MetricRow>;

export const Critical: Story = {
  args: {
    name: "LiDAR accuracy",
    before: 98.1,
    after: 91.3,
    format: "percent",
  },
};

export const Warning: Story = {
  args: {
    name: "CPU usage",
    before: 34,
    after: 48,
    format: "percent",
  },
};

export const Stable: Story = {
  args: {
    name: "Memory",
    before: 412,
    after: 418,
    format: "bytes",
  },
};

export const Neutral: Story = {
  args: {
    name: "Fusion latency",
    before: 12,
    after: 12,
    format: "ms",
  },
};
