import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Button } from "@/components/ui/button";
import { EmptyState } from "./empty-state";

const meta: Meta<typeof EmptyState> = {
  title: "Components/EmptyState",
  component: EmptyState,
  parameters: { layout: "padded" },
};

export default meta;
type Story = StoryObj<typeof EmptyState>;

export const Terse: Story = {
  args: {
    message: "No results matched your search",
  },
};

export const WithDescription: Story = {
  args: {
    message: "No deploys yet",
    description:
      "Trigger your first deploy to see per-device metric diffs appear here.",
  },
};

export const WithCTA: Story = {
  args: {
    message: "No deploys yet",
    description:
      "Trigger your first deploy to see per-device metric diffs appear here.",
    cta: (
      <>
        <Button variant="default" size="sm">
          Trigger deploy
        </Button>
        <Button variant="outline" size="sm">
          Read the docs
        </Button>
      </>
    ),
  },
};

export const PermissionDenied: Story = {
  args: {
    message: "No deploys yet",
    description:
      "An operator will trigger the first deploy. You'll see it here once it runs.",
  },
};

export const WithCodeExample: Story = {
  args: {
    message: "No devices in the fleet",
    description: "Install dispatch-agent on a device to register it.",
    cta: (
      <pre className="rounded-md border border-border bg-muted/50 px-4 py-3 text-left font-mono text-xs text-foreground">
        <code>curl -sSL dispatch.dev/install.sh | sh</code>
      </pre>
    ),
  },
};

export const Compact: Story = {
  args: {
    message: "No metrics yet",
    description:
      "Device is online but hasn't reported any metrics.",
    compact: true,
  },
};
