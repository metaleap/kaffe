import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags, depends = van.derive

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
            htm.div({
                'class': depends(() => 'buddy-self' + ((haxsh.selectedBuddy.val === 0) ? ' selected' : '') + ((haxsh.buddyBadges[0].val) ? ' badged' : '')),
                'data-badge': depends(() => (haxsh.buddyBadges[0].val) || ""),
                'onclick': () => {
                    if (!haxsh.isSeeminglyOffline.val) {
                        haxsh.buddySelected(undefined, true)
                    }
                },
            },
                htm.div(userDomAttrsSelf()),
            ),
        ),
        buddies: vanx.reactive([] as yo.User[]),
        update: (buddies) => update(me, buddies),
    }

    van.add(me.DOM, vanx.list(() => htm.div({ 'class': 'buddies' }), me.buddies, (it) => {
        const item = htm.div({
            'class': depends(() => 'buddy' + (haxsh.isSeeminglyOffline.val ? ' offline' : '') + (haxsh.buddySelected(it.val) ? ' selected' : '') + ((haxsh.buddyBadges[it.val.Id].val) ? ' badged' : '')),
            'data-badge': depends(() => (haxsh.buddyBadges[it.val.Id].val) || ""),
        },
            htm.div(userDomAttrsBuddy(it.val, new Date().getTime())))
        item.onclick = () => {
            if (!haxsh.isSeeminglyOffline.val) {
                haxsh.buddySelected(it.val, true)
            }
        }
        return item
    }))
    return me
}

function isOffline(user: yo.User, now?: number) {
    const last_seen = new Date(user.LastSeen ?? (user.DtMod!)).getTime()
    return (((now ?? youtil.dtNow()) - last_seen) > 77777)
}

export function userPicFileUrl(user?: yo.User, fallBackToEmoji = '👤', toRoundedSvgFavIcon = false) {
    if (!(user && user.PicFileId))
        return util.svgTextIconDataHref(fallBackToEmoji)
    return '/_postfiles/' + user.PicFileId + (toRoundedSvgFavIcon ? '?picRounded=true' : '')
}

export function userDomAttrsBuddy(user?: yo.User, now?: number) {
    if (!user)
        return {
            'class': 'buddy-pic offline',
            'title': "(ex-buddy — or bug)",
            'style': `background-image: url('${userPicFileUrl()}')`,
        }
    return {
        'class': depends(() => 'buddy-pic' + ((haxsh.isSeeminglyOffline.val || isOffline(user, now)) ? ' offline' : '')),
        'title': `${user.Nick}${((!user.Btw) ? '' : (' — ' + user.Btw))}`,
        'style': `background-image: url('${userPicFileUrl(user)}')`,
    }
}

export function userDomAttrsSelf() {
    return {
        'class': depends(() => 'buddy-pic self' + (haxsh.isSeeminglyOffline.val ? ' offline' : '')),
        'title': depends(() => {
            const user_self = haxsh.userSelf.val
            return (!user_self) ? "(you)" : `${user_self.Nick}${((!user_self.Btw) ? '' : (' — ' + user_self.Btw))}`
        }),
        'style': depends(() => {
            const user_self = haxsh.userSelf.val
            return `background-image: url('${userPicFileUrl(user_self)}')`
        }),
    }
}

function update(me: UiCtlBuddies, buddies: yo.User[]): number {
    const now = new Date().getTime()
    const is_selected: { [_: number]: boolean } = {}
    for (let i = 0; i < buddies.length; i++) {
        const buddy = buddies[i]
        is_selected[buddy.Id] = haxsh.buddySelected(buddy)
        if ((i > 0) && (is_selected[buddy.Id])) {
            for (let j = 0; j < i; j++) {
                const earlier = buddies[j]
                if (!is_selected[earlier.Id]) {
                    buddies[j] = buddy
                    buddies[i] = earlier
                    break
                }
            }
        }
    }
    if (!youtil.deepEq(buddies, me.buddies.filter(_ => true), false, false))
        vanx.replace(me.buddies, (_: yo.User[]) => buddies)
    return buddies.filter(_ => !isOffline(_, now)).length
}
