const plugin = require("tailwindcss/plugin");

module.exports = {
  darkMode: "class",
  content: ["./**/*.gohtml", "./**/*.templ"],
  theme: {
    colors: {
      transparent: "transparent",
      current: "currentColor",
      "primary-50": "#f8fafc",
      "primary-100": "#f1f5f9",
      "primary-200": "#e2e8f0",
      "primary-300": "#cbd5e1",
      "primary-400": "#94a3b8",
      "primary-500": "#64748b",
      "primary-600": "#475569",
      "primary-700": "#334155",
      "primary-800": "#1e293b",
      "primary-900": "#0f172a",
      "primary-950": "#020617",
      "action-800": "#166534",
      "action-700": "#15803d",
      "action-600": "#16a34a",
      "action-500": "#22c55e",
      "action-400": "#4ade80",
      "action-300": "#86efac",
      error: "#ef4444",
      warn: "#eab308",
      link: "#3b82f6",
      "tag-1": "#c084fc",
      "tag-2": "#f472b6",
      "tag-3": "#fb923c",
    },
    extend: {
      transitionProperty: {
        "max-height": "max-height",
      },
    },
    fontFamily: {
      sans: ["Fira Code", "fira-code"],
      mono: ["Fira Code", "fira-code"],
      serif: ["Fira Code", "fira-code"],
    },
  },
  plugins: [
    plugin(function ({ matchUtilities, theme }) {
      matchUtilities(
        {
          "translate-z": (value) => ({
            "--tw-translate-z": value,
            transform: ` translate3d(var(--tw-translate-x), var(--tw-translate-y), var(--tw-translate-z)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))`,
          }),
        },
        { values: theme("translate"), supportsNegativeValues: true },
      );
    }),
  ],
};
