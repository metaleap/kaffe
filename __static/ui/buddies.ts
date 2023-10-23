import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
import * as yo from '../yo-sdk.js'

const htm = van.tags

export type UiCtlBuddies = {
    DOM: HTMLElement
    buddies: vanx.Reactive<yo.User[]>
}

export function create(): UiCtlBuddies {
    const me: UiCtlBuddies = {
        DOM: htm.div({ 'class': 'haxsh-buddies' }),
        buddies: vanx.reactive([
            { Id: 1234, Btw: "Btw 1234", Nick: "user1234", PicFileId: "", Auth: 1234 } as yo.User,
            { Id: 2345, Btw: "Btw 2345", Nick: "user2345", PicFileId: "", Auth: 2345 } as yo.User,
            { Id: 4321, Btw: "Btw 4321", Nick: "user4321", PicFileId: "", Auth: 4321 } as yo.User,
            { Id: 123, Btw: "Btw 123", Nick: "user123", PicFileId: "", Auth: 123 } as yo.User,
            { Id: 321, Btw: "Btw 321", Nick: "user321", PicFileId: "", Auth: 321 } as yo.User,
        ])
    }

    setInterval(() => {
        const idx = Math.floor(me.buddies.length * Math.random())
        const buddy = me.buddies[idx]
        buddy.Btw = new Date().toLocaleTimeString()
    }, 1234)

    van.add(me.DOM, vanx.list(htm.ul, me.buddies, (it) => {
        const buddy = it.val
        return htm.li({}, buddy.Nick, buddy.LastSeen, buddy.Btw)
    }))

    return me
}
