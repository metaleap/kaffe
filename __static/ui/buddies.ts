import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as util from '../util.js'

export type UiCtlBuddies = {
    DOM: HTMLElement
    buddies: vanx.Reactive<yo.User[]>
    update: (_: yo.User[]) => void
}

export function create(): UiCtlBuddies {
    const me: UiCtlBuddies = {
        DOM: htm.div({ 'class': 'haxsh-buddies' },
            htm.div(buddyDomAttrs(undefined, new Date().getTime(), true, "(loading...)")),
        ),
        buddies: vanx.reactive([] as yo.User[]),
        update: (buddies) => update(me, buddies),
    }

    van.add(me.DOM, vanx.list(htm.ul, me.buddies, (it) => {
        return htm.li(buddyDomAttrs(it.val, new Date().getTime()))
    }))
    return me
}

export function buddyDomAttrs(buddy: yo.User | undefined, now: number, isSelf = false, noneText = "(ex-buddy)") {
    if (!buddy)
        return {
            'class': 'buddy-pic' + (isSelf ? ' self' : ' offline'),
            'title': noneText,
            'style': `background-image: url('${util.emoIconDataHref('🦜')}')`
        }
    if (!buddy.LastSeen)
        buddy.LastSeen = buddy.DtMod
    const last_seen = new Date(buddy.LastSeen!).getTime()
    const is_offline = (now - last_seen) > 77777
    return {
        'class': 'buddy-pic' + (is_offline ? ' offline' : ''),
        'title': `${buddy.Nick}${((!buddy.Btw) ? '' : (' — ' + buddy.Btw))}`,
        'style': `background-image: url('${buddy.PicFileId ? ("/__static/mockfiles/" + buddy.PicFileId) : util.emoIconDataHref('👤')}')`
    }
}

function update(me: UiCtlBuddies, buddies: yo.User[]) {
    if (!youtil.deepEq(buddies, me.buddies, false))
        vanx.replace(me.buddies, (_: yo.User[]) => buddies)
}
