import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { SoakTimer } from "./soak-timer";

const meta: Meta<typeof SoakTimer> = {
  title: "Components/SoakTimer",
  component: SoakTimer,
};

export default meta;
type Story = StoryObj<typeof SoakTimer>;

export const JustStarted: Story = { args: { elapsed: 3, total: 30 } };
export const HalfWay: Story = { args: { elapsed: 15, total: 30 } };
export const AlmostDone: Story = { args: { elapsed: 27, total: 30 } };
export const Complete: Story = { args: { elapsed: 30, total: 30 } };
