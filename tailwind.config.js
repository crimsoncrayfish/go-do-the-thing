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
        "tag-4": "#60a5fa",
        "tag-5": "#a78bfa",
        "tag-6": "#fcd34d",
        "tag-7": "#4ade80",
        "tag-8": "#e879f9",
        "tag-9": "#22d3ee",
        "tag-10": "#f87171",
      },
      screens: {
        'xs': '475px',
      },
      containers: {
        'xs': '475px',
        'sm': '640px',
        'md': '768px',
        'lg': '1024px',
        'xl': '1280px',
        '2xl': '1536px',
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
    require("flowbite/plugin"),
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
  flowbite: {
    theme: {
      light: {
        datepicker: {
          cell: {
            hover: "bg-accent-light",
            selected: "bg-secondary-light",
          },
          controls: {
            prevBtn: "hover:bg-secondary-light",
            nextBtn: "hover:bg-secondary-light",
            viewSwitch: "hover:bg-secondary-light",
          },
        },
      },
      dark: {
        datepicker: {
          cell: {
            hover: "dark:bg-accent-dark",
            selected: "dark:bg-secondary-dark",
          },
          controls: {
            prevBtn: "dark:hover:bg-secondary-dark",
            nextBtn: "dark:hover:bg-secondary-dark",
            viewSwitch: "dark:hover:bg-secondary-dark",
          },
        },
      },
    },
  },
};
