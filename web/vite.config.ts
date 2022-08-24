import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"

import { fileURLToPath, URL } from "url"

export default defineConfig({
    plugins: [ vue() ],
    resolve: {
        alias: {
            "@": fileURLToPath(new URL("./src", import.meta.url)),
        },
    },
    base: process.env.NODE_ENV === "production" ? "/syodo/" : "/",
    define: {
        __SYODO_API__: JSON.stringify("https://e0uf7jciif.execute-api.eu-central-1.amazonaws.com/production"),
    },
})
