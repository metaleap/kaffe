import van, { State } from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags, depends = van.derive

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as haxsh from '../haxsh.js'
import * as util from '../util.js'

export type UiCtlBuddies = {
    DOM: HTMLElement
    buddies: vanx.Reactive<yo.User[]>
    buddyRequestsBy: State<yo.User[]>
    update: (_: yo.userBuddies_Out) => void
}

export function create(): UiCtlBuddies {
    const me: UiCtlBuddies = {
        DOM: htm.div({ 'class': 'haxsh-buddies' },
            htm.div({
                'class': depends(() => 'buddy-self' + ((haxsh.selectedBuddy.val === 0) ? ' selected' : '') + ((haxsh.buddyBadges[0].val) ? ' badged' : '') + ((haxsh.buddyBadgesAlt[0].val) ? ' badged-alt' : '')),
                'data-badge': depends(() => (haxsh.buddyBadges[0].val) || ""),
                'data-badge-alt': depends(() => (haxsh.buddyBadgesAlt[0].val) || ""),
                'onclick': () => {
                    if (!haxsh.isSeeminglyOffline.val)
                        haxsh.buddySelected(undefined, true)
                },
            }, htm.div(userDomAttrsSelf())),
        ),
        buddies: vanx.reactive([] as yo.User[]),
        buddyRequestsBy: van.state([] as yo.User[]),
        update: (_) => update(me, _),
    }

    van.add(me.DOM, vanx.list(() => htm.div({ 'class': 'buddies' }), me.buddies, (it) => {
        const item = htm.div({
            'class': depends(() => 'buddy' + (haxsh.isSeeminglyOffline.val ? ' offline' : '') + (haxsh.buddySelected(it.val) ? ' selected' : '') + ((haxsh.buddyBadges[it.val.Id!].val) ? ' badged' : '') + ((haxsh.buddyBadgesAlt[it.val.Id!].val) ? ' badged-alt' : '')),
            'data-badge': depends(() => (haxsh.buddyBadges[it.val.Id!].val) || ""),
            'data-badge-alt': depends(() => (haxsh.buddyBadgesAlt[it.val.Id!].val) || ""),
        }, htm.div(userDomAttrsBuddy(it.val)))
        item.onclick = () => {
            if (!haxsh.isSeeminglyOffline.val)
                haxsh.buddySelected(it.val, true)
        }
        return item
    }))
    van.add(me.DOM, htm.div({
        'style': depends(() => haxsh.userSelf.val ? '' : 'display:none'),
        'class': depends(() => 'buddy' + (me.buddyRequestsBy.val.length ? ' badged' : '')),
        'data-badge': depends(() => me.buddyRequestsBy.val.length || ""),
        'onclick': () => showBuddiesDialog(me),
    }, htm.div({ 'class': 'buddy-pic', 'title': "Manage buddies", 'style': `background-image: url('${userPicFileUrl(undefined, "👥")}')` })))
    return me
}

function isOffline(user: yo.User, now?: number) {
    const last_seen = new Date(user.LastSeen ?? (user.DtMod!)).getTime()
    return (((now ?? youtil.dtNow()) - last_seen) > 77777)
}

export function userPicFileUrl(user?: yo.User, fallBackToEmoji = "🦜", toRoundedSvgFavIcon = false) {
    if (user && !user.Auth)
        fallBackToEmoji = "👤"
    if (!(user && user.PicFileId))
        return util.svgTextIconDataHref(fallBackToEmoji)
    return '/_postfiles/' + user.PicFileId + (toRoundedSvgFavIcon ? '?picRounded=true' : '')
}

export function userDomAttrsBuddy(user?: yo.User, userIdHint?: number) {
    if (!user)
        return {
            'class': 'buddy-pic offline',
            'title': `(ex-buddy #${userIdHint ?? -1} — or bug)`,
            'style': `background-image: url('${userPicFileUrl()}')`,
        }
    if (!user.Auth) //
        return {
            'class': 'buddy-pic offline',
            'title': `${user.Nick} — (buddy request still pending)`,
            'style': `background-image: url('${userPicFileUrl(user)}')`,
        }
    return {
        'class': depends(() => 'buddy-pic' + ((haxsh.isSeeminglyOffline.val || isOffline(user)) ? ' offline' : '')),
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

function update(me: UiCtlBuddies, buddiesInfo: yo.userBuddies_Out) {
    me.buddyRequestsBy.val = buddiesInfo.BuddyRequestsBy ?? []
    const buddies = buddiesInfo.Buddies ?? []
    const move_selected_top = false // actually irritating ux-wise, so decided-against for now
    if (move_selected_top) {
        const is_selected: { [_: number]: boolean } = {}
        for (let i = 0; i < buddies.length; i++) {
            const buddy = buddies[i]
            is_selected[buddy.Id!] = haxsh.buddySelected(buddy)
            if ((i > 0) && (is_selected[buddy.Id!])) {
                for (let j = 0; j < i; j++) {
                    const earlier = buddies[j]
                    if (!is_selected[earlier.Id!]) {
                        buddies[j] = buddy
                        buddies[i] = earlier
                        break
                    }
                }
            }
        }
    }
    if (!youtil.deepEq(buddies, me.buddies.filter(_ => true), false, false))
        vanx.replace(me.buddies, (_: yo.User[]) => buddies)
}

async function showBuddiesDialog(me: UiCtlBuddies) {
    const add_new_buddy = async () => {
        let nick_or_email_addr = prompt("New buddy's nick or email address?", "")
        if (nick_or_email_addr && nick_or_email_addr.length && (nick_or_email_addr = nick_or_email_addr.trim())) {
            if (me.buddies.filter(_ => true).some(_ => _.Nick === nick_or_email_addr))
                alert(`A buddy request for '${nick_or_email_addr}' had already been placed at an earlier time, but your eagerness is commendable.`)
            else
                try {
                    const result = await yo.apiUserBuddiesAdd({ NickOrEmailAddr: nick_or_email_addr })
                    if (!result.Done)
                        alert(`No user '${nick_or_email_addr}' was found. Typo?`)
                    else
                        alert(`Done. Until '${nick_or_email_addr}' confirms, they'll appear offline in your buddy list.`)
                } catch (err) {
                    if (!haxsh.knownErr<yo.UserBuddiesAddErr>(err, haxsh.handleKnownErrMaybe<yo.UserBuddiesAddErr>))
                        haxsh.onErrOther(err, true)
                }
        }
    }
    const confirm_new_buddy = async (htmCheckbox: HTMLInputElement, user: yo.User) => {
        htmCheckbox.checked = false
        if (!confirm(`Sure to become buddies with '${user.Nick}' now? (Just checking in case your cat triggered this...)`))
            return
        htmCheckbox.style.cursor = 'wait'
        htmCheckbox.disabled = true
        try {
            await haxsh.reloadUserSelf()
            const user_self = haxsh.userSelf?.val
            if (!user_self)
                return
            if ((!user_self.Buddies) || !user_self.Buddies.some(_ => (_ === user.Id!))) {
                await yo.apiUserUpdate({
                    Id: user_self.Id, Changes: { Buddies: [user.Id!].concat(user_self.Buddies ?? []) }, ChangedFields: ['Buddies']
                }, new FormData())
            }
            htmCheckbox.checked = true
            alert(`You're now buddies with '${user.Nick ?? '?'}', go chat them up!`)
            await haxsh.reloadUserSelf()
        } catch (err) {
            if (!haxsh.knownErr<yo.UserUpdateErr>(err, haxsh.handleKnownErrMaybe<yo.UserUpdateErr>))
                haxsh.onErrOther(err, true)
        } finally {
            htmCheckbox.style.removeProperty('cursor')
        }
    }
    const dialog = htm.dialog({ 'class': 'buddies-popup' },
        htm.button({ 'type': 'button', 'class': 'addnew', 'title': "Add a buddy...", 'onclick': _ => add_new_buddy() }, "➕"),
        htm.button({ 'type': 'button', 'class': 'close', 'title': "Close", 'onclick': _ => dialog.close() }, "❎"),
        htm.div({}, htm.h3({}, "Who wants to be buddies:")), (!me.buddyRequestsBy.val.length)
        ? htm.div({}, "No buddy requests received lately. ", htm.a({ 'onclick': () => add_new_buddy() }, " Add a buddy..."))
        : htm.ul({}, ...me.buddyRequestsBy.val.map(buddy_to_be => {
            let htm_checkbox = htm.input({ 'id': 'chk_confirm_' + buddy_to_be.Id, 'type': 'checkbox', 'onclick': () => confirm_new_buddy(htm_checkbox, buddy_to_be) })
            return htm.li({},
                htm.div(userDomAttrsBuddy(buddy_to_be)),
                htm.b(buddy_to_be.Nick),
                htm.hr(),
                htm.span({},
                    htm_checkbox,
                    htm.label({ 'for': 'chk_confirm_' + buddy_to_be.Id }, "Become buddies now?")
                ),
            )
        }))
    )
    dialog.onclose = _ => dialog.remove()

    van.add(me.DOM, dialog)
    dialog.showModal()
}
