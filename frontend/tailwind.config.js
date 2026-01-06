/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        background: "#131314",
        surface: "#1E1F20",
        border: "#3C4043",
        primary: "#A8C7FA",
        success: "#C4EED0",
        warning: "#FDE293",
        purple: "#D291FF",
        error: "#F28B82",
        main: "#FFFFFF",
        sub: "#E3E3E3",
        muted: "#9AA0A6",
      },
      fontFamily: {
        sans: ['Google Sans', 'Inter', 'system-ui', 'sans-serif'],
      }
    },
  },
  plugins: [],
}