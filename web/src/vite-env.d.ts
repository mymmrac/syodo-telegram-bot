/// <reference types="vite/client" />

declare module "*.vue" {
    import type { DefineComponent } from "vue"
    const component: DefineComponent<{}, {}, any>
    export default component
}

declare const __IS_DEV__: boolean
declare const __SYODO_API__: string
declare const __BOT_API__: string
