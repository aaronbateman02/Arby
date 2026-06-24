import type { Config } from "tailwindcss"
const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  theme: {
    extend: {
      colors: {
        surface: { DEFAULT: "#0f1118", alt: "#1a1d2e", hover: "#252841" },
        accent: { DEFAULT: "#6366f1", hover: "#818cf8", dim: "#4f46e5" },
        green: { DEFAULT: "#22c55e", dim: "#166534" },
        red: { DEFAULT: "#ef4444", dim: "#991b1b" },
        amber: { DEFAULT: "#f59e0b", dim: "#92400e" },
        muted: "#6b7280",
        border: "#2d3148",
      },
    },
  },
  plugins: [],
}
export default config
