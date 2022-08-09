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

export function sendError(type: string, data: any) {
    console.error(`Error type:${ type }, data: ${ data }`)
    tg.HapticFeedback.notificationOccurred("error")
    tg.sendData(`${ type }:${ data }`)
}

export function href(path: string): string {
    return new URL(import.meta.env.BASE_URL + path, import.meta.url).href
}
