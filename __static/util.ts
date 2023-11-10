import van from '../__yostatic/vanjs/van-1.2.6.js'
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

export function domLive<T extends { Id?: any }>(outer: () => Element, initial: T[], perItem: (_: T) => Element) {
    const me = {
        lastNodes: {} as { [_: string]: Element },
        lastItems: {} as { [_: string]: T },
        itemCount: van.state(initial.length),
        outer: outer(),
        update: (items: T[]) => {
            me.itemCount.val = items.length
            // any old items fully gone, hence dom nodes to remove?
            const del_nodes: Element[] = []
            for (const id in me.lastItems) {
                if (!initial.some(_ => (_.Id!) == id)) { // no `===` here due to string-vs-number ambiguity
                    const node_old = me.lastNodes[id]
                    del_nodes.push(node_old)
                    delete me.lastNodes[id]
                    delete me.lastItems[id]
                }
            }
            for (const del_node of del_nodes)
                del_node.replaceWith()
            // ignoring changes in sort order for now here, actual node-(re)create ops per item
            for (const i in initial.filter(_ => true)) {
                const item = initial[i]
                const node_old = me.lastNodes[item.Id!], item_old = me.lastItems[item.Id!]
                if (!youtil.deepEq(item, item_old, false, false)) {
                    let node = node_old
                    if (node)  // change dom node
                        node.replaceWith(perItem(item))
                    else  // new dom append
                        node = perItem(item) // not yet in dom, happens further down during sort-order-ensuring
                    me.lastNodes[item.Id!] = node
                }
                me.lastItems[item.Id!] = item
            }
            // ensure up-to-date sort order
            for (let i = 0, l = (items.length - 1); i < l; i++) {
                const node_this = me.lastNodes[items[i].Id!], node_next = me.lastNodes[items[i + 1].Id!]
                if (node_this.nextElementSibling !== node_next)
                    me.outer.insertBefore(node_this, node_next)
            }
        }
    }
    for (const item of initial)
        me.lastItems[item.Id!.toString()] = item
    me.update(initial)
    return me
}
