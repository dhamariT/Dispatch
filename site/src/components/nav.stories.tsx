import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Nav } from "./nav";

const meta: Meta<typeof Nav> = {
  title: "Components/Nav",
  component: Nav,
  parameters: { layout: "fullscreen" },
};

export default meta;
type Story = StoryObj<typeof Nav>;

export const Default: Story = {
  args: {
    fleetName: "PiRacer Fleet (Demo)",
  },
};

export const LongFleetName: Story = {
  args: {
    fleetName: "Production Delivery Robots — West Coast",
  },
};
