const plugin = require("tailwindcss/plugin");

module.exports = {
  darkMode: "class",
  content: ["./**/*.gohtml", "./**/*.templ", "./node_modules/flowbite/**/*.js"],
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
        primary: "primary-light",
        // you can also override datepicker-specific tokens if you like:
        datepicker: {
          cell: {
            hover: "secondary-light",
            selected: "accent-light",
          },
        },
      },
      dark: {
        primary: "primary-dark",
        datepicker: {
          cell: {
            hover: "secondary-dark",
            selected: "accent-dark",
          },
        },
      },
    },
  },
};
