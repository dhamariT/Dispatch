import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { DeployTopbar } from "./deploy-topbar";

const meta: Meta<typeof DeployTopbar> = {
  title: "Components/DeployTopbar",
  component: DeployTopbar,
  parameters: { layout: "fullscreen" },
};

export default meta;
type Story = StoryObj<typeof DeployTopbar>;

export const CanarySoaking: Story = {
  args: {
    from: "v1.4.2",
    to: "v1.4.3",
    status: "soaking",
    soakElapsed: 14,
    soakTotal: 30,
    critical: 1,
    warning: 0,
    waiting: 2,
    offline: 0,
    total: 3,
  },
};

export const HealthyDeploy: Story = {
  args: {
    from: "v1.4.1",
    to: "v1.4.2",
    status: "complete",
    critical: 0,
    warning: 0,
    waiting: 0,
    offline: 0,
    total: 3,
  },
};

export const LargeFleetIncident: Story = {
  args: {
    from: "v1.4.2",
    to: "v1.4.3",
    status: "soaking",
    soakElapsed: 8,
    soakTotal: 30,
    critical: 3,
    warning: 12,
    waiting: 985,
    offline: 0,
    total: 1000,
  },
};
