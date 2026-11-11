const defaultTheme = require('tailwindcss/defaultTheme');

module.exports = {
  content: ['./src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#2563eb', // blue-600
          light: '#eff6ff',
          dark: '#1d4ed8',
        },
        secondary: {
          DEFAULT: '#64748b', // slate-500
          light: '#f1f5f9',
          dark: '#475569',
        },
        success: { DEFAULT: '#10b981', light: '#d1fae5', dark: '#059669' },
        danger: { DEFAULT: '#ef4444', light: '#fee2e2', dark: '#dc2626' },
        warning: { DEFAULT: '#f59e0b', light: '#fef3c7', dark: '#d97706' },
        brand: {
          navy: '#0f172a', // slate-900
          accent: '#38bdf8', // sky-400
        },
      },
      fontFamily: {
        sans: ['Inter', ...defaultTheme.fontFamily.sans],
      },
    },
  },
  plugins: [],
};
