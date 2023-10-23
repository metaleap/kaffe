import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
import * as yo from '../yo-sdk.js'

const htm = van.tags

export type UiCtlBuddies = {
    DOM: HTMLElement
    buddies: vanx.Reactive<yo.User[]>
    update: (_: yo.User[]) => void
}

export function create(): UiCtlBuddies {
    const me: UiCtlBuddies = {
        DOM: htm.div({ 'class': 'haxsh-buddies' }),
        buddies: vanx.reactive([] as yo.User[]),
        update: (buddies) => update(me, buddies),
    }

    van.add(me.DOM, vanx.list(htm.ul, me.buddies, (it) => {
        const buddy = it.val
        return htm.li({}, " Nick: ", buddy.Nick, " Last: ", buddy.LastSeen, " Btw: ", buddy.Btw)
    }))

    return me
}

function update(me: UiCtlBuddies, buddies: yo.User[]) {
    vanx.replace(me.buddies, (_: yo.User[]) => buddies)
}
