/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./App.{js,jsx,ts,tsx}", "./src/**/*.{js,jsx,ts,tsx}"],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        // Primary - Bold Purple
        primary: {
          50: '#f6f5ff',
          100: '#ede9fe',
          500: '#7c3aed',
          600: '#5b21b6',
          700: '#4c1d95',
        },
        // Secondary - Hot Pink
        secondary: {
          400: '#f43f9f',
          500: '#ec1b8d',
          600: '#be185d',
        },
        // Accent - Cyan
        accent: {
          400: '#22d3ee',
          500: '#06b6d4',
          600: '#0e7490',
        },
        // Semantic
        success: {
          500: '#059669',
          600: '#047857',
        },
        error: {
          500: '#dc2626',
          600: '#991b1b',
        },
        warning: {
          500: '#d97706',
          600: '#b45309',
        },
      },
      spacing: {
        'xs': '4px',
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '20px',
        '2xl': '24px',
      },
      borderRadius: {
        'lg': '12px',
        'xl': '16px',
        '2xl': '20px',
      },
      fontSize: {
        'xs': '12px',
        'sm': '14px',
        'base': '16px',
        'lg': '18px',
        'xl': '20px',
      },
      fontWeight: {
        'bold': '700',
        'extrabold': '800',
      },
      boxShadow: {
        'sm': '0 1px 3px 0 rgba(0, 0, 0, 0.08)',
        'md': '0 4px 8px 0 rgba(0, 0, 0, 0.12)',
        'lg': '0 8px 16px 0 rgba(0, 0, 0, 0.15)',
      },
    },
  },
  plugins: [],
  darkMode: "class",
}

