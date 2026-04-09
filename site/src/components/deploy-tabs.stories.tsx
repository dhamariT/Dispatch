import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { DeployTabs, type DeployTab } from "./deploy-tabs";

const meta: Meta<typeof DeployTabs> = {
  title: "Components/DeployTabs",
  component: DeployTabs,
  parameters: { layout: "padded" },
};

export default meta;
type Story = StoryObj<typeof DeployTabs>;

const sampleDeploys: DeployTab[] = [
  { id: "d1", from: "v1.4.2", to: "v1.4.3", severity: "critical" },
  { id: "d2", from: "v1.4.1", to: "v1.4.2", severity: "stable" },
  { id: "d3", from: "v1.4.0", to: "v1.4.1", severity: "warning" },
  { id: "d4", from: "v1.3.9", to: "v1.4.0", severity: "stable" },
];

export const Default: Story = {
  render: () => {
    const [selected, setSelected] = useState("d1");
    return (
      <DeployTabs
        deploys={sampleDeploys}
        selected={selected}
        onSelect={setSelected}
      />
    );
  },
};

export const SingleDeploy: Story = {
  render: () => {
    const [selected, setSelected] = useState("d1");
    return (
      <DeployTabs
        deploys={[sampleDeploys[0]]}
        selected={selected}
        onSelect={setSelected}
      />
    );
  },
};

export const ManyDeploys: Story = {
  render: () => {
    const [selected, setSelected] = useState("d1");
    const many: DeployTab[] = Array.from({ length: 12 }, (_, i) => ({
      id: `d${i}`,
      from: `v1.${4 - Math.floor(i / 3)}.${3 - (i % 3)}`,
      to: `v1.${4 - Math.floor(i / 3)}.${4 - (i % 3)}`,
      severity: (["critical", "warning", "stable", "stable"] as const)[i % 4],
    }));
    return (
      <DeployTabs deploys={many} selected={selected} onSelect={setSelected} />
    );
  },
};
