import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { DeviceGroup } from "./device-group";

const meta: Meta<typeof DeviceGroup> = {
  title: "Components/DeviceGroup",
  component: DeviceGroup,
  parameters: { layout: "padded" },
};

export default meta;
type Story = StoryObj<typeof DeviceGroup>;

export const DeployedWithRegression: Story = {
  args: {
    deviceId: "car-2",
    status: "deployed",
    metrics: [
      { name: "LiDAR accuracy", before: 98.1, after: 91.3, format: "percent" },
      { name: "CPU usage", before: 34, after: 48, format: "percent" },
      { name: "Memory", before: 412, after: 418, format: "bytes" },
      { name: "Fusion latency", before: 12, after: 14, format: "ms" },
    ],
  },
};

export const DeployedHealthy: Story = {
  args: {
    deviceId: "car-1",
    status: "deployed",
    metrics: [
      { name: "LiDAR accuracy", before: 97.8, after: 98.0, format: "percent" },
      { name: "CPU usage", before: 31, after: 33, format: "percent" },
      { name: "Memory", before: 408, after: 410, format: "bytes" },
    ],
  },
};

export const Waiting: Story = {
  args: {
    deviceId: "car-3",
    status: "waiting",
  },
};

export const Offline: Story = {
  args: {
    deviceId: "car-4",
    status: "offline",
  },
};

export const FullFleet: Story = {
  render: () => (
    <div className="flex flex-col gap-6">
      <DeviceGroup
        deviceId="car-2"
        status="deployed"
        metrics={[
          { name: "LiDAR accuracy", before: 98.1, after: 91.3, format: "percent" },
          { name: "CPU usage", before: 34, after: 48, format: "percent" },
          { name: "Memory", before: 412, after: 418, format: "bytes" },
          { name: "Fusion latency", before: 12, after: 14, format: "ms" },
        ]}
      />
      <DeviceGroup deviceId="car-1" status="waiting" />
      <DeviceGroup deviceId="car-3" status="waiting" />
    </div>
  ),
};
