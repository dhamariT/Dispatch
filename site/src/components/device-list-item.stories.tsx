import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { DeviceListItem } from "./device-list-item";

const meta: Meta<typeof DeviceListItem> = {
  title: "Components/DeviceListItem",
  component: DeviceListItem,
  parameters: { layout: "padded" },
};

export default meta;
type Story = StoryObj<typeof DeviceListItem>;

export const Critical: Story = {
  args: {
    deviceId: "car-2",
    status: "deployed",
    metrics: [
      { name: "LiDAR accuracy", before: 98.1, after: 91.3, format: "percent" },
      { name: "CPU usage", before: 34, after: 48, format: "percent" },
    ],
  },
};

export const Warning: Story = {
  args: {
    deviceId: "car-7",
    status: "deployed",
    metrics: [
      { name: "CPU usage", before: 34, after: 37, format: "percent" },
      { name: "Memory", before: 412, after: 425, format: "bytes" },
    ],
  },
};

export const Stable: Story = {
  args: {
    deviceId: "car-1",
    status: "deployed",
    metrics: [
      { name: "LiDAR accuracy", before: 97.8, after: 98.0, format: "percent" },
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

export const Expanded: Story = {
  args: {
    deviceId: "car-2",
    status: "deployed",
    expanded: true,
    metrics: [
      { name: "LiDAR accuracy", before: 98.1, after: 91.3, format: "percent" },
    ],
  },
};

export const LargeFleet: Story = {
  render: () => (
    <div className="flex flex-col gap-2">
      <DeviceListItem
        deviceId="car-042"
        status="deployed"
        metrics={[{ name: "LiDAR accuracy", before: 98.1, after: 91.3, format: "percent" }]}
      />
      <DeviceListItem
        deviceId="car-108"
        status="deployed"
        metrics={[{ name: "Fusion latency", before: 12, after: 38, format: "ms" }]}
      />
      <DeviceListItem
        deviceId="car-215"
        status="deployed"
        metrics={[{ name: "CPU usage", before: 34, after: 37, format: "percent" }]}
      />
      <DeviceListItem
        deviceId="car-003"
        status="deployed"
        metrics={[{ name: "LiDAR accuracy", before: 97.8, after: 98.0, format: "percent" }]}
      />
      <DeviceListItem
        deviceId="car-019"
        status="deployed"
        metrics={[{ name: "CPU usage", before: 31, after: 33, format: "percent" }]}
      />
      <DeviceListItem deviceId="car-512" status="waiting" />
      <DeviceListItem deviceId="car-887" status="offline" />
    </div>
  ),
};
