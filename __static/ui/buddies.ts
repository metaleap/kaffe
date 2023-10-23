import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
import * as yo from '../yo-sdk.js'
import * as util from '../util.js'

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
        const buddy = it.val, now = new Date().getTime()
        if (!buddy.LastSeen)
            buddy.LastSeen = buddy.DtMod!
        const is_offline = (now - Date.parse(buddy.LastSeen)) > 77777
        return htm.li({
            'class': is_offline ? 'offline' : '',
            'title': `${buddy.Nick}${(!buddy.Btw) ? '' : (' — ' + buddy.Btw)}`,
            'style': `background-image: url('${buddy.PicFileId ? ("/__static/mockfiles/" + buddy.PicFileId) : util.emoIconDataHref('👤')}')`
        },)
    }))

    return me
}

function update(me: UiCtlBuddies, buddies: yo.User[]) {
    vanx.replace(me.buddies, (_: yo.User[]) => buddies)
}
