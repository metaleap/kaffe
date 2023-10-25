import { State } from '../__yostatic/vanjs/van-1.2.3.debug.js'


export function svgTextIconDataHref(emoji: string): string {
    return 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text y=".9em" font-size="83">' + emoji + '</text></svg>'
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
    if (diff <= hour)
        return `${Math.floor(diff / minute)}m${suffix}`
    if (diff <= day)
        return `${Math.floor(diff / hour)}h${suffix}`
    return `${Math.floor(diff / day)}d${suffix}`
}


export function reactive<T extends { uxId: string }>(outer: HTMLElement, inner: (_: State<T>) => HTMLElement) {
    const states: { [_: string]: State<T> } = {}
    return function (fresh: T[]) {
        for (const ux_id in states)
            if (!fresh.find(_ => (_.uxId === ux_id)))
                delete states[ux_id]
        for (const item of fresh) {
            const state = states[item.uxId]
            const is_new = state ? true : false
            if (!is_new)
                state.val = item
        }
    }
}
