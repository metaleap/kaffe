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

export type DomLive<T extends { [_: string]: any }> = {
    all: T[]
    outer: Element
    itemCount: State<number>
    timeLastModifiedDomWise: State<number>
    onUpdated: (_: T[]) => void
}

export function domLive<T extends { [_: string]: any }>(outer: Element, initial: T[], perItem: (_: T) => Element, identPropName = 'Id'): DomLive<T> {
    const id = identPropName
    type _DomLive = DomLive<T> & {
        _lastNodes: { [_: string]: Element }
        _lastItems: { [_: string]: T }
    }
    const me: _DomLive = {
        _lastNodes: {} as { [_: string]: Element },
        _lastItems: {} as { [_: string]: T },

        all: initial,
        itemCount: van.state(initial.length),
        timeLastModifiedDomWise: van.state(0),
        outer: outer,
        onUpdated: (items: T[]) => {
            me.itemCount.val = items.length
            let dom_modified = false
            // any old items fully gone, hence dom nodes to remove?
            const del_nodes: Element[] = []
            for (const id in me._lastItems) {
                if (!items.some(_ => (_[id]!) == id)) { // no `===` here due to string-vs-number ambiguity
                    const node_old = me._lastNodes[id]
                    del_nodes.push(node_old)
                    delete me._lastNodes[id]
                    delete me._lastItems[id]
                }
            }
            if (dom_modified = (del_nodes.length > 0))
                for (const del_node of del_nodes)
                    del_node.replaceWith()
            // ignoring changes in sort order for now here, actual node-(re)create ops per item
            const new_nodes = [] as Element[]
            for (let i = 0; i < items.length; i++) {
                const item = items[i]
                const node_old = me._lastNodes[item[id]!], item_old = me._lastItems[item[id]!]
                if (!youtil.deepEq(item, item_old, false, false)) {
                    const node_now = perItem(item)
                    if (node_old)  // change dom node
                        node_old.replaceWith(node_now)
                    else  // new dom append
                        new_nodes.push(node_now)
                    me._lastNodes[item[id]!] = node_now
                }
                me._lastItems[item[id]!] = item
            }
            if (new_nodes.length > 0) {
                dom_modified = true
                me.outer.append(...new_nodes)
            }
            // ensure up-to-date sort order
            for (let i = 0, l = (items.length - 1); i < l; i++) {
                const node_this = me._lastNodes[items[i][id]!], node_next = me._lastNodes[items[i + 1][id]!]
                if (node_this.nextElementSibling !== node_next) {
                    dom_modified = true
                    me.outer.insertBefore(node_this, node_next)
                }
            }

            me.all = items
            if (dom_modified)
                me.timeLastModifiedDomWise.val = Date.now()
        }
    }
    me.onUpdated(initial)
    return me
}
