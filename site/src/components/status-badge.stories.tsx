import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { StatusBadge } from "./status-badge";

const meta: Meta<typeof StatusBadge> = {
  title: "Components/StatusBadge",
  component: StatusBadge,
};

export default meta;
type Story = StoryObj<typeof StatusBadge>;

export const Canary: Story = { args: { status: "canary" } };
export const Soaking: Story = { args: { status: "soaking" } };
export const Wave2: Story = { args: { status: "wave2" } };
export const Wave3: Story = { args: { status: "wave3" } };
export const Complete: Story = { args: { status: "complete" } };
export const RolledBack: Story = { args: { status: "rolledback" } };

export const AllStates: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <StatusBadge status="canary" />
      <StatusBadge status="soaking" />
      <StatusBadge status="wave2" />
      <StatusBadge status="wave3" />
      <StatusBadge status="complete" />
      <StatusBadge status="rolledback" />
    </div>
  ),
};
