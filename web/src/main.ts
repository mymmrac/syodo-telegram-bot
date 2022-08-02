import { createApp } from "vue"
import "./style.scss"
import App from "./App.vue"
import { createPinia } from "pinia"

const pinia = createPinia()

createApp(App)
    .use(pinia)
    .mount("#app")

// Updates light/dark theme
function setThemeClass() {
    document.documentElement.className = window.Telegram.WebApp.colorScheme
}

window.Telegram.WebApp.onEvent("theme_changed", setThemeClass)
setThemeClass()
