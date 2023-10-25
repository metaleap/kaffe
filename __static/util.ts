export function svgTextIconDataHref(emoji: string): string {
    return 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text y=".9em" font-size="83">' + emoji + '</text></svg>'
}

export function svgRoundedImgDataHref(imgHref: string): string {
    //
    return 'data:image/svg+xml,' + encodeURIComponent(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="100" height="100" viewBox="0 0 100 100">
        <clipPath id="c"><circle cx="50" cy="50" r="50" /></clipPath>
        <image x="0" y="0" width="100" height="100" clip-path="url(#c)" href="${imgHref}"/>
    </svg>`)
}

export function timeAgoStr(when: number, now: number, noSecs: boolean, suffix = " ago") {
    const second = 1000
    const minute = 60 * second
    const hour = 60 * minute
    const day = 24 * hour
    const diff = now - when
    if ((diff <= 1000) || ((diff <= minute) && noSecs))
        return "now"
    if (diff <= minute)
        return `${Math.floor(diff / second)}s${suffix}`
    if (diff < hour)
        return `${Math.floor(diff / minute)}m${suffix}`
    if (diff <= day)
        return `${Math.floor(diff / hour)}h${suffix}`
    return `${Math.floor(diff / day)}d${suffix}`
}
