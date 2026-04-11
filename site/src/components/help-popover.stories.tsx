import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import {
  HelpPopover,
  HelpPopoverContent,
  HelpPopoverIconTrigger,
  HelpPopoverLink,
  HelpPopoverLinksGroup,
  HelpPopoverText,
  HelpPopoverTitle,
} from "./help-popover";

const meta: Meta = {
  title: "Components/HelpPopover",
};

export default meta;
type Story = StoryObj;

export const TitleAndText: Story = {
  render: () => (
    <HelpPopover defaultOpen>
      <HelpPopoverIconTrigger />
      <HelpPopoverContent>
        <HelpPopoverTitle>What is a deploy?</HelpPopoverTitle>
        <HelpPopoverText>
          A deploy is a new release rolled out to your fleet. Dispatch snapshots
          metrics before and after to show you what changed on each device.
        </HelpPopoverText>
      </HelpPopoverContent>
    </HelpPopover>
  ),
};

export const WithLinks: Story = {
  render: () => (
    <HelpPopover defaultOpen>
      <HelpPopoverIconTrigger />
      <HelpPopoverContent>
        <HelpPopoverTitle>What is a soak period?</HelpPopoverTitle>
        <HelpPopoverText>
          Soak is the wait period after a deploy where Dispatch collects
          post-deploy metrics. Once complete, you can promote to the next wave.
        </HelpPopoverText>
        <HelpPopoverLinksGroup>
          <HelpPopoverLink href="https://dispatch.dev/docs/rollouts">
            Rollout strategies
          </HelpPopoverLink>
          <HelpPopoverLink href="https://dispatch.dev/docs/soak-timers">
            Configuring soak timers
          </HelpPopoverLink>
        </HelpPopoverLinksGroup>
      </HelpPopoverContent>
    </HelpPopover>
  ),
};

export const InlineWithLabel: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <span className="text-sm font-medium">Soak</span>
      <HelpPopover>
        <HelpPopoverIconTrigger />
        <HelpPopoverContent>
          <HelpPopoverTitle>Soak period</HelpPopoverTitle>
          <HelpPopoverText>
            The wait period after a deploy where Dispatch collects post-deploy
            metrics.
          </HelpPopoverText>
        </HelpPopoverContent>
      </HelpPopover>
    </div>
  ),
};

export const MediumIcon: Story = {
  render: () => (
    <HelpPopover>
      <HelpPopoverIconTrigger size="md" />
      <HelpPopoverContent>
        <HelpPopoverTitle>Larger trigger</HelpPopoverTitle>
        <HelpPopoverText>
          Use the medium size when the icon needs more visual weight, like next
          to a page title.
        </HelpPopoverText>
      </HelpPopoverContent>
    </HelpPopover>
  ),
};
