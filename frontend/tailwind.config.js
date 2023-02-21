/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
    },
    fontFamily: {
      library: ['LIBRARY 3 AM']
    }
  },
  daisyui: {
    themes: ["business"],
  },
  plugins: [require("daisyui")],
};
