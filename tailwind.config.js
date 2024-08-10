/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: ["./tmpl/**/*.gohtml"],
  theme: {
    extend : {
      transitionProperty: {
        'max-height': 'max-height'
      }
    },
    fontFamily: {
      sans: ["Fira Code", "fira-code"],
      mono: ["Fira Code", "fira-code"],
      serif: ["Fira Code", "fira-code"],
    },
  },
  plugins: [],
}

