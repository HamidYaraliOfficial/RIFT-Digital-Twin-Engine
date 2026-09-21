import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["selector", '[data-mode="dark"]'],
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        surface: "var(--surface)",
        "surface-2": "var(--surface-2)",
        "surface-3": "var(--surface-3)",
        border: "var(--border)",
        accent: "var(--accent)",
        "accent-2": "var(--accent-2)",
        ink: "var(--ink)",
        "ink-muted": "var(--ink-muted)",
        success: "var(--success)",
        warning: "var(--warning)",
        danger: "var(--danger)",
      },
      fontFamily: {
        sans: ["var(--font-sans)"],
        fa: ["var(--font-fa)"],
      },
      borderRadius: {
        xl2: "14px",
      },
      boxShadow: {
        acrylic: "0 8px 30px rgba(0,0,0,0.12)",
      },
    },
  },
  plugins: [],
};
export default config;
