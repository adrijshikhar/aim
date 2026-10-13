/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        geist: {
          canvas: "#000000",
          surface: "#0a0a0a",
          elevated: "#111111",
          recessed: "#000000",
          hover: "#1a1a1a",
          hairline: "#262626",
          hairlineSoft: "#1f1f1f",
          border: "#262626",
          borderHover: "#404040",
          heading: "#ffffff",
          ink: "#ededed",
          secondary: "#a1a1a1",
          body: "#a1a1a1",
          mute: "#888888",
          faint: "#666666",
          blue: "#0070f3",
          mint: "#30a46c",
          cyan: "#50e3c2",
          warning: "#f5a623",
          error: "#e5484d",
        },
        cursor: {
          orange: "#f54e00",
          orangeHover: "#d04200",
          orangeMuted: "rgba(245, 78, 0, 0.15)",
          peach: "#dfa88f",      // Thinking / Reasoning
          mint: "#9fc9a2",       // Exec / Bash / Commands
          blue: "#9fbbe0",       // Read / Inspection
          lavender: "#c0a8dd",   // Edit / Mutation / Files
          gold: "#c08532",       // Completion / Done / Fork
          green: "#1f8a65",      // Success / Authenticated
          crimson: "#cf2d56",    // Error / Destructive
        },
        timeline: {
          thinking: "#dfa88f",
          exec: "#9fc9a2",
          read: "#9fbbe0",
          edit: "#c0a8dd",
          done: "#c08532",
        },
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        ds: {
          red: {
            bg: "var(--ds-red-bg)",
            border: "var(--ds-red-border)",
            solid: "var(--ds-red-solid)",
            text: "var(--ds-red-text)",
          },
          amber: {
            bg: "var(--ds-amber-bg)",
            border: "var(--ds-amber-border)",
            solid: "var(--ds-amber-solid)",
            text: "var(--ds-amber-text)",
          },
          green: {
            bg: "var(--ds-green-bg)",
            border: "var(--ds-green-border)",
            solid: "var(--ds-green-solid)",
            text: "var(--ds-green-text)",
          },
          blue: {
            bg: "var(--ds-blue-bg)",
            border: "var(--ds-blue-border)",
            solid: "var(--ds-blue-solid)",
            text: "var(--ds-blue-text)",
          },
        },
        status: {
          danger: {
            bg: "var(--ds-red-bg)",
            border: "var(--ds-red-border)",
            solid: "var(--ds-red-solid)",
            text: "var(--ds-red-text)",
          },
          warning: {
            bg: "var(--ds-amber-bg)",
            border: "var(--ds-amber-border)",
            solid: "var(--ds-amber-solid)",
            text: "var(--ds-amber-text)",
          },
          success: {
            bg: "var(--ds-green-bg)",
            border: "var(--ds-green-border)",
            solid: "var(--ds-green-solid)",
            text: "var(--ds-green-text)",
          },
          info: {
            bg: "var(--ds-blue-bg)",
            border: "var(--ds-blue-border)",
            solid: "var(--ds-blue-solid)",
            text: "var(--ds-blue-text)",
          },
          meter: "var(--ds-meter-healthy)",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
        geist: "6px",
      },
      fontFamily: {
        sans: [
          "Geist",
          "-apple-system",
          "BlinkMacSystemFont",
          '"Segoe UI"',
          "Roboto",
          "sans-serif",
        ],
        mono: [
          '"Geist Mono"',
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "monospace",
        ],
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
};
