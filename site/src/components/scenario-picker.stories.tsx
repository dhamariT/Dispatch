import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ScenarioPicker, type ScenarioOption } from "./scenario-picker";

const scenarios: ScenarioOption[] = [
  {
    name: "lidar_regression",
    description: "Canary's LiDAR accuracy drops significantly. CPU spikes. Auto-hold triggers on 3 of 4 metrics.",
    outcome: "auto_hold",
    icon: "shield",
  },
  {
    name: "healthy_deploy",
    description: "All metrics stable across canary and control groups. No significant differences detected.",
    outcome: "promote",
    icon: "check",
  },
  {
    name: "noisy_but_safe",
    description: "High variance in both groups drowns out small differences. Tests the effect size gate.",
    outcome: "promote",
    icon: "activity",
  },
  {
    name: "slow_regression",
    description: "Small CPU regression with moderate effect size. Borderline case that tests the decision threshold.",
    outcome: "auto_hold",
    icon: "gauge",
  },
];

const meta: Meta<typeof ScenarioPicker> = {
  title: "Experiment/ScenarioPicker",
  component: ScenarioPicker,
  args: { scenarios },
  parameters: { layout: "padded" },
};

export default meta;
type Story = StoryObj<typeof ScenarioPicker>;

export const Default: Story = {
  args: { selected: null },
};

export const Selected: Story = {
  args: { selected: "lidar_regression" },
};

export const SelectedHealthy: Story = {
  args: { selected: "healthy_deploy" },
};

export const Running: Story = {
  args: { selected: "lidar_regression", running: true },
};
