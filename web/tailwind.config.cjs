/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./index.html",
        "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                "tg-bg": "var(--tg-theme-bg-color)",
                "tg-text": "var(--tg-theme-text-color)",
                "tg-hint": "var(--tg-theme-hint-color)",
                "tg-link": "var(--tg-theme-link-color)",
                "tg-button": "var(--tg-theme-button-color)",
                "tg-button-text": "var(--tg-theme-button-text-color)",
                "tg-secondary-bg": "var(--tg-theme-secondary-bg-color)",
            },
        },
    },
    plugins: [
        require("@tailwindcss/forms"),
    ],
}
