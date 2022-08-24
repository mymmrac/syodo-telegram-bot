import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"

import { fileURLToPath, URL } from "url"

const isProd = process.env.NODE_ENV === "production"

export default defineConfig({
    plugins: [ vue() ],
    resolve: {
        alias: {
            "@": fileURLToPath(new URL("./src", import.meta.url)),
        },
    },
    base: isProd ? "/syodo/" : "/",
    define: {
        __SYODO_API__: JSON.stringify("https://e0uf7jciif.execute-api.eu-central-1.amazonaws.com/production"),
        __BOT_API__: JSON.stringify(isProd ? "https://mymm.gq/syodo-bot" : "http://localhost:8080"),
    },
})
