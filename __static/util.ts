export function emoIconDataHref(emoji: string): string {
    return "data:image/svg+xml,&lt;svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22&gt;&lt;text y=%22.9em%22 font-size=%2283%22&gt;" + emoji + "&lt;/text&gt;&lt;/svg&gt;"
}

export function timeAgoStr(when: number, now: number, noSecs: boolean) {
    const second = 1000
    const minute = 60 * second
    const hour = 60 * minute
    const day = 24 * hour
    const diff = now - when
    if ((diff <= 1000) || ((diff <= minute) && noSecs))
        return "just now"
    if (diff <= minute)
        return `${Math.floor(diff / second)}s ago`
    if (diff <= hour)
        return `${Math.floor(diff / minute)}m ago`
    if (diff <= day)
        return `${Math.floor(diff / hour)}h ago`
    return `${Math.floor(diff / day)}d ago`
}
