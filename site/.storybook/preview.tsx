import type { Preview, ReactRenderer } from "@storybook/nextjs-vite";
import type { DecoratorFunction } from "storybook/internal/types";
import React, { useEffect } from "react";
import "../src/app/globals.css";

// Applies the selected theme by toggling the dark class on <html>.
// Tailwind and our CSS variables respond to this class.
const withTheme: DecoratorFunction<ReactRenderer> = (Story, context) => {
  const theme = context.globals.theme || "dark";

  useEffect(() => {
    document.documentElement.classList.toggle("dark", theme === "dark");
  }, [theme]);

  return (
    <div className="bg-background text-foreground min-h-screen p-6">
      <Story />
    </div>
  );
};

const preview: Preview = {
  globalTypes: {
    theme: {
      description: "Theme",
      toolbar: {
        title: "Theme",
        icon: "paintbrush",
        items: [
          { value: "dark", title: "Dark", icon: "moon" },
          { value: "light", title: "Light", icon: "sun" },
        ],
        dynamicTitle: true,
      },
    },
  },
  initialGlobals: {
    theme: "dark",
  },
  parameters: {
    layout: "padded",
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    a11y: {
      test: "todo",
    },
    viewport: {
      viewports: {
        desktop: { name: "Desktop", styles: { width: "1440px", height: "900px" } },
        tablet: { name: "Tablet", styles: { width: "768px", height: "1024px" } },
        compact: { name: "Compact", styles: { width: "400px", height: "400px" } },
      },
    },
  },
  decorators: [withTheme],
};

export default preview;
