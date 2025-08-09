const plugin = require("tailwindcss/plugin");

module.exports = {
  darkMode: "class",
  content: [
    "./**/*.gohtml",
    "./**/*.templ",
    "./static/js/datepicker.js",
    "./static/js/colorpicker.js",
  ],
  theme: {
    extend: {
      colors: {
        "primary-dark": "#2D1B69",
        "secondary-dark": "#392B7D",
        "accent-dark": "#B197FC",
        "success-dark": "#5580CC",
        "danger-dark": "#ED4E4E",
        "warning-dark": "#FFB86C",
        "border-dark": "#4B3B8E",

        "primary-light": "#F5F3FF",
        "secondary-light": "#EDE9FE",
        "accent-light": "#8B5CF6",
        "success-light": "#7C3AED",
        "danger-light": "#E11D48",
        "warning-light": "#D97706",
        "border-light": "#C4B5FD",

        "text-on-light": "#5B21B6",
        "text-on-dark": "#D5CCF7",
        transparent: "transparent",
        current: "currentColor",
        link: "#3b82f6",
        red: {
          light: "#FECACA",
          dark: "#B91C1C",
        },
        orange: {
          light: "#FED7AA",
          dark: "#C2410C",
        },
        amber: {
          light: "#FDE68A",
          dark: "#B45309",
        },
        green: {
          light: "#A7F3D0",
          dark: "#047857",
        },
        cyan: {
          light: "#A5F3FC",
          dark: "#0E7490",
        },
        blue: {
          light: "#BFDBFE",
          dark: "#1D4ED8",
        },
        indigo: {
          light: "#C7D2FE",
          dark: "#4338CA",
        },
        purple: {
          light: "#DDD6FE",
          dark: "#7E22CE",
        },
        fuchsia: {
          light: "#F5D0FE",
          dark: "#C026D3",
        },
        rose: {
          light: "#FECDD3",
          dark: "#BE123C",
        },
      },
      screens: {
        xs: "475px",
      },
      containers: {
        xs: "475px",
        sm: "640px",
        md: "768px",
        lg: "1024px",
        xl: "1280px",
        "2xl": "1536px",
      },
      transitionProperty: {
        "max-height": "max-height",
      },
      lineClamp: {
        1: "1",
        2: "2",
        3: "3",
        4: "4",
        5: "5",
        6: "6",
      },
    },
    fontFamily: {
      sans: ["Fira Code", "fira-code"],
      mono: ["Fira Code", "fira-code"],
      serif: ["Fira Code", "fira-code"],
    },
  },
  plugins: [
    require("@tailwindcss/line-clamp"),
    require("@tailwindcss/container-queries"),
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
