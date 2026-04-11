import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { InfoTooltip } from "./info-tooltip";

const meta: Meta<typeof InfoTooltip> = {
  title: "Components/InfoTooltip",
  component: InfoTooltip,
};

export default meta;
type Story = StoryObj<typeof InfoTooltip>;

export const Default: Story = {
  args: {
    title: "What is a canary?",
    message:
      "The canary is the first device to receive a deploy. If its metrics look healthy after the soak period, the deploy promotes to the next wave.",
  },
};

export const InlineWithLabel: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <span className="text-xs uppercase tracking-wider text-muted-foreground">
        Delta
      </span>
      <InfoTooltip
        title="Delta"
        message="The difference between the after and before values. Red means the metric regressed past the critical threshold, amber means it's above warning."
      />
    </div>
  ),
};

export const ColumnHeader: Story = {
  render: () => (
    <table className="w-full border-collapse">
      <thead>
        <tr className="border-b border-border text-left text-xs uppercase tracking-wider text-muted-foreground">
          <th className="px-4 py-2">Metric</th>
          <th className="px-4 py-2">
            <div className="flex items-center gap-1.5">
              Before
              <InfoTooltip
                title="Before"
                message="The metric value captured just before the deploy was triggered."
              />
            </div>
          </th>
          <th className="px-4 py-2">
            <div className="flex items-center gap-1.5">
              After
              <InfoTooltip
                title="After"
                message="The metric value captured after the soak period completed."
              />
            </div>
          </th>
          <th className="px-4 py-2">
            <div className="flex items-center gap-1.5">
              Delta
              <InfoTooltip
                title="Delta"
                message="The difference between after and before. Red means the metric regressed past the critical threshold."
              />
            </div>
          </th>
        </tr>
      </thead>
    </table>
  ),
};
