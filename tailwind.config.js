const plugin = require("tailwindcss/plugin");

module.exports = {
  darkMode: "class",
  content: ["./**/*.gohtml", "./**/*.templ", "./node_modules/flowbite/**/*.js"],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#f8fafc",
          100: "#f1f5f9",
          200: "#e2e8f0",
          300: "#cbd5e1",
          400: "#94a3b8",
          500: "#64748b",
          600: "#475569",
          700: "#334155",
          800: "#1e293b",
          900: "#0f172a",
          950: "#020617",
        },
        transparent: "transparent",
        current: "currentColor",
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
    require("flowbite/plugin"),
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
  flowbite: {
    theme: {
      light: {
        // re-map Flowbite’s “primary” to your `primary-500`
        primary: "#64748b",
        // you can also override datepicker-specific tokens if you like:
        datepicker: {
          cell: {
            hover: "primary-100",
            selected: "primary-500",
          },
        },
      },
      dark: {
        primary: "#334155",
        datepicker: {
          cell: {
            hover: "primary-800",
            selected: "primary-600",
          },
        },
      },
    },
  },
};
