import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ConfidenceBadge } from "./confidence-badge";

const meta: Meta<typeof ConfidenceBadge> = {
  title: "Experiment/ConfidenceBadge",
  component: ConfidenceBadge,
};

export default meta;
type Story = StoryObj<typeof ConfidenceBadge>;

export const HighConfidence: Story = {
  args: { pValue: 0.00007 },
};

export const Moderate: Story = {
  args: { pValue: 0.003 },
};

export const Low: Story = {
  args: { pValue: 0.042 },
};

export const Insufficient: Story = {
  args: { pValue: 0.34 },
};

export const AllLevels: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <ConfidenceBadge pValue={0.00003} />
      <ConfidenceBadge pValue={0.0041} />
      <ConfidenceBadge pValue={0.038} />
      <ConfidenceBadge pValue={0.52} />
    </div>
  ),
};
