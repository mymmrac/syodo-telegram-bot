import { TelegramWebApps } from "telegram-bots-webapps-types"

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

export function insert<T>(arr: T[], index: number, newItem: T): T[] {
    return [ ...arr.slice(0, index), newItem, ...arr.slice(index) ]
}

export function scrollToTop(behavior: ScrollBehavior = "auto") {
    window.scrollTo({ top: 0, behavior: behavior })
}

export function scrollToID(id: string) {
    document.getElementById(id)?.scrollIntoView()
}

export function tgVersionSupported(version: string): boolean {
    const [ actualMajor, actualMinor ] = tg.version.split(".").map(Number)
    const [ expectedMajor, expectedMinor ] = version.split(".").map(Number)

    return actualMajor > expectedMajor || (actualMajor == expectedMajor && actualMinor >= expectedMinor)
}

export function showError(type: string, message: string, err?: any) {
    console.error(`Type:${ type }, message: ${ message }, error: ${ err }`)
    tg.HapticFeedback.notificationOccurred("error")

    const alertMsg = err ? message + "\n\n" + err : message

    if (tgVersionSupported("6.2")) {
        tg.showAlert(alertMsg)
    } else {
        alert(alertMsg)
    }
}

export function href(path: string): string {
    return new URL(import.meta.env.BASE_URL + path, import.meta.url).href
}
