import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { FleetSummary } from "./fleet-summary";

const meta: Meta<typeof FleetSummary> = {
  title: "Components/FleetSummary",
  component: FleetSummary,
  parameters: { layout: "padded" },
};

export default meta;
type Story = StoryObj<typeof FleetSummary>;

export const SmallFleet: Story = {
  args: {
    total: 3,
    critical: 1,
    warning: 0,
    stable: 0,
    waiting: 2,
    offline: 0,
  },
};

export const LargeFleetHealthy: Story = {
  args: {
    total: 1000,
    critical: 0,
    warning: 3,
    stable: 997,
    waiting: 0,
    offline: 0,
  },
};

export const LargeFleetProblems: Story = {
  args: {
    total: 1000,
    critical: 3,
    warning: 12,
    stable: 978,
    waiting: 0,
    offline: 7,
  },
};

export const AllHealthy: Story = {
  args: {
    total: 50,
    critical: 0,
    warning: 0,
    stable: 50,
    waiting: 0,
    offline: 0,
  },
};

export const MidDeploy: Story = {
  args: {
    total: 100,
    critical: 0,
    warning: 0,
    stable: 15,
    waiting: 85,
    offline: 0,
  },
};
