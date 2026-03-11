/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        hm: {
          bg: "#0a0a0a",
          surface: "#111111",
          border: "#1f1f1f",
          muted: "#666666",
          "muted-light": "#555555",
          text: "#fafafa",
          "text-passive": "#e8e8e8",
          accent: "#e5e5e5",
          amber: "#F5A623",
          "code-bg": "#0f0f0f",
          "terminal-bg": "#0d0d0d",
          "terminal-header": "#161616",
        },
      },
      fontFamily: {
        sans: ['"DM Sans"', "Geist", "system-ui", "-apple-system", "sans-serif"],
        mono: ['"JetBrains Mono"', "Geist Mono", "ui-monospace", "monospace"],
      },
      maxWidth: {
        prose: "65ch",
        content: "80ch",
      },
      spacing: {
        "hm-xs": "0.25rem",
        "hm-sm": "0.5rem",
        "hm-md": "0.75rem",
        "hm-lg": "1rem",
        "hm-xl": "1.25rem",
      },
      transitionDuration: { DEFAULT: "100ms" },
    },
  },
  plugins: [],
};
