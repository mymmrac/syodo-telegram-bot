/** @type {import("tailwindcss").Config} */
module.exports = {
    darkMode: "class",
    content: [
        "./index.html",
        "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                "tg-bg": "var(--tg-theme-bg-color, #ffffff)",
                "tg-text": "var(--tg-theme-text-color, #222222)",
                "tg-hint": "var(--tg-theme-hint-color, #a8a8a8)",
                "tg-link": "var(--tg-theme-link-color, #2678b6)",
                "tg-button": "#bb4347",
                "tg-button-text": "#ffffff",
                "tg-secondary-bg": "var(--tg-theme-secondary-bg-color, #f0f0f0)",
            },
        },
    },
    plugins: [
        require("@tailwindcss/forms"),
    ],
}
