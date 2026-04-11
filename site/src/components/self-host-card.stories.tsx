import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { SelfHostCard } from "./self-host-card";

const meta: Meta<typeof SelfHostCard> = {
  title: "Components/SelfHostCard",
  component: SelfHostCard,
  parameters: { layout: "padded" },
};

export default meta;
type Story = StoryObj<typeof SelfHostCard>;

export const Default: Story = {};
