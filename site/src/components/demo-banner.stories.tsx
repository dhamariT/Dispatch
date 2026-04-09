import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { DemoBanner } from "./demo-banner";

const meta: Meta<typeof DemoBanner> = {
  title: "Components/DemoBanner",
  component: DemoBanner,
  parameters: { layout: "fullscreen" },
};

export default meta;
type Story = StoryObj<typeof DemoBanner>;

export const Default: Story = {};
