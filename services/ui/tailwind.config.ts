import type { Config } from "tailwindcss"
const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  theme: {
    extend: {
      colors: {
        surface: { DEFAULT: "#ffffff", alt: "#f8f9fa", hover: "#f1f3f5" },
        accent: { DEFAULT: "#6366f1", hover: "#818cf8", dim: "#4f46e5" },
        green: { DEFAULT: "#16a34a", dim: "#bbf7d0" },
        red: { DEFAULT: "#dc2626", dim: "#fecaca" },
        amber: { DEFAULT: "#d97706", dim: "#fde68a" },
        muted: "#6b7280",
        border: "#e5e7eb",
        gray: {
          50: "#f9fafb",
          100: "#111827",
          200: "#1f2937",
          300: "#374151",
          400: "#4b5563",
          500: "#6b7280",
          600: "#9ca3af",
          700: "#d1d5db",
          800: "#e5e7eb",
          900: "#f3f4f6",
        },
      },
    },
  },
  plugins: [],
}
export default config
