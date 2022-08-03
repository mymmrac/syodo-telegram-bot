import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"

import { fileURLToPath, URL } from "url"

export default defineConfig({
    base: process.env.NODE_ENV === "production" ? "/syodo/" : "/",
    plugins: [ vue() ],
    resolve: {
        alias: {
            "@": fileURLToPath(new URL("./src", import.meta.url)),
        },
    },
})
