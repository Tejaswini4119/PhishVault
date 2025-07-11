module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      animation: {
        'spin-slow': 'spin 18s linear infinite',
        'fade-in-down': 'fadeInDown 0.8s both',
        'fade-in-up': 'fadeInUp 0.8s both',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {
        fadeInDown: {
          '0%': { opacity: 0, transform: 'translateY(-40px)' },
          '100%': { opacity: 1, transform: 'translateY(0)' },
        },
        fadeInUp: {
          '0%': { opacity: 0, transform: 'translateY(40px)' },
          '100%': { opacity: 1, transform: 'translateY(0)' },
        },
      },
      colors: {
        'dark-bg': '#10141a',
        'dark-card': '#181c23',
        'accent': '#38bdf8',
        'accent2': '#2563eb',
      },
    },
  },
  plugins: [],
}