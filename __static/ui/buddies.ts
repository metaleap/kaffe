import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as haxsh from '../haxsh.js'
import * as util from '../util.js'

export type UiCtlBuddies = {
    DOM: HTMLElement
    buddies: vanx.Reactive<yo.User[]>
    update: (_: yo.User[]) => number
}

export function create(): UiCtlBuddies {
    const me: UiCtlBuddies = {
        DOM: htm.div({ 'class': 'haxsh-buddies' },
            htm.div({ 'class': 'buddy-self' },
                htm.div(userDomAttrsSelf()),
            ),
        ),
        buddies: vanx.reactive([] as yo.User[]),
        update: (buddies) => update(me, buddies),
    }

    van.add(me.DOM, vanx.list(htm.ul, me.buddies, (it) => {
        return htm.li(userDomAttrsBuddy(it.val, new Date().getTime()))
    }))
    return me
}

function isOffline(user: yo.User, now: number) {
    const last_seen = new Date(user.LastSeen ?? (user.DtMod!)).getTime()
    return ((now - last_seen) > 77777)
}

export function userPicFileUrl(user?: yo.User, fallBackToEmoji = '👤', toRoundedSvgFavIcon = false) {
    if (!(user && user.PicFileId))
        return util.svgTextIconDataHref(fallBackToEmoji)
    return '/__static/mockfiles/' + user.PicFileId + (toRoundedSvgFavIcon ? '?picRounded=true' : '')
}

export function userDomAttrsBuddy(user: yo.User | undefined, now: number) {
    if (!user)
        return {
            'class': 'buddy-pic offline',
            'title': "(ex-buddy — or bug)",
            'style': `background-image: url('${userPicFileUrl()}')`
        }
    const is_offline = isOffline(user, now)
    return {
        'class': 'buddy-pic' + (is_offline ? ' offline' : ''),
        'title': `${user.Nick}${((!user.Btw) ? '' : (' — ' + user.Btw))}`,
        'style': `background-image: url('${userPicFileUrl(user)}')`
    }
}

export function userDomAttrsSelf() {
    return {
        'class': 'buddy-pic self',
        'title': van.derive(() => {
            const user_self = haxsh.userSelf.val
            return (!user_self) ? "(you)" : `${user_self.Nick}${((!user_self.Btw) ? '' : (' — ' + user_self.Btw))}`
        }),
        'style': van.derive(() => {
            const user_self = haxsh.userSelf.val
            return `background-image: url('${userPicFileUrl(user_self)}')`
        }),
    }
}

function update(me: UiCtlBuddies, buddies: yo.User[]): number {
    const now = new Date().getTime()
    if (!youtil.deepEq(buddies, me.buddies, false, false))
        vanx.replace(me.buddies, (_: yo.User[]) => buddies)
    return buddies.filter(_ => !isOffline(_, now)).length
}
