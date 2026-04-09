import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { StatusDot } from "./status-dot";

const meta: Meta<typeof StatusDot> = {
  title: "Components/StatusDot",
  component: StatusDot,
};

export default meta;
type Story = StoryObj<typeof StatusDot>;

export const AllVariants: Story = {
  render: () => (
    <div className="flex items-center gap-6">
      <div className="flex flex-col items-center gap-2">
        <StatusDot variant="critical" size="md" />
        <span className="text-xs text-muted-foreground">critical</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <StatusDot variant="warning" size="md" />
        <span className="text-xs text-muted-foreground">warning</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <StatusDot variant="stable" size="md" />
        <span className="text-xs text-muted-foreground">stable</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <StatusDot variant="active" size="md" />
        <span className="text-xs text-muted-foreground">active</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <StatusDot variant="waiting" size="md" />
        <span className="text-xs text-muted-foreground">waiting</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <StatusDot variant="offline" size="md" />
        <span className="text-xs text-muted-foreground">offline</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <StatusDot variant="neutral" size="md" />
        <span className="text-xs text-muted-foreground">neutral</span>
      </div>
    </div>
  ),
};

export const AllSizes: Story = {
  render: () => (
    <div className="flex items-center gap-6">
      <StatusDot variant="critical" size="xs" />
      <StatusDot variant="critical" size="sm" />
      <StatusDot variant="critical" size="md" />
      <StatusDot variant="critical" size="lg" />
    </div>
  ),
};

export const Pulsing: Story = {
  render: () => (
    <div className="flex items-center gap-6">
      <StatusDot variant="critical" size="md" pulse />
      <StatusDot variant="warning" size="md" pulse />
      <StatusDot variant="active" size="md" pulse />
    </div>
  ),
};
