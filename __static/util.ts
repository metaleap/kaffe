import van, { State } from '../__yostatic/vanjs/van-1.2.6.js'
import * as youtil from '../__yostatic/util.js'

export function svgTextIconDataHref(emoji: string): string {
    return 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text x="0em" y="0.9em" font-size="80">' + emoji + '</text></svg>'
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

export type DomLive<T extends { Id?: any }> = {
    outer: Element
    itemCount: State<number>
    update: (_: T[]) => void
}

export function domLive<T extends { Id?: any }>(outer: () => Element, initial: T[], perItem: (_: T) => Element): DomLive<T> {
    const me = {
        _lastNodes: {} as { [_: string]: Element },
        _lastItems: {} as { [_: string]: T },

        itemCount: van.state(initial.length),
        outer: outer(),
        update: (items: T[]) => {
            me.itemCount.val = items.length
            // any old items fully gone, hence dom nodes to remove?
            const del_nodes: Element[] = []
            for (const id in me._lastItems) {
                if (!items.some(_ => (_.Id!) == id)) { // no `===` here due to string-vs-number ambiguity
                    const node_old = me._lastNodes[id]
                    del_nodes.push(node_old)
                    delete me._lastNodes[id]
                    delete me._lastItems[id]
                }
            }
            for (const del_node of del_nodes)
                del_node.replaceWith()
            // ignoring changes in sort order for now here, actual node-(re)create ops per item
            const new_nodes = [] as Element[]
            for (let i = 0; i < items.length; i++) {
                const item = items[i]
                const node_old = me._lastNodes[item.Id!], item_old = me._lastItems[item.Id!]
                if (!youtil.deepEq(item, item_old, false, false)) {
                    const node_now = perItem(item)
                    if (node_old)  // change dom node
                        node_old.replaceWith(node_now)
                    else  // new dom append
                        new_nodes.push(node_now)
                    me._lastNodes[item.Id!] = node_now
                }
                me._lastItems[item.Id!] = item
            }
            if (new_nodes.length > 0)
                me.outer.append(...new_nodes)
            // ensure up-to-date sort order
            for (let i = 0, l = (items.length - 1); i < l; i++) {
                const node_this = me._lastNodes[items[i].Id!], node_next = me._lastNodes[items[i + 1].Id!]
                if (node_this.nextElementSibling !== node_next)
                    me.outer.insertBefore(node_this, node_next)
            }
        }
    }
    me.update(initial)
    return me
}
