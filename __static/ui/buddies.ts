import van, { State } from '../../__yostatic/vanjs/van-1.2.6.js'
const htm = van.tags, depends = van.derive

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as youi from '../../__yostatic/youi.js'
import * as kaffe from '../kaffe.js'
import * as util from '../util.js'

export type UiCtlBuddies = {
    DOM: HTMLElement
    buddies: youi.DomLive<yo.User>
    buddyRequestsBy: State<yo.User[]>
    update: (_: yo.__userBuddies_Out) => void
}

export function create(): UiCtlBuddies {
    const me: UiCtlBuddies = {
        update: (_) => update(me, _),
        buddyRequestsBy: van.state([] as yo.User[]),
        buddies: youi.domLive<yo.User>(htm.div({ 'class': 'buddies' }), [], (it) => {
            const item = htm.div({
                'class': depends(() => 'buddy' + (kaffe.isSeeminglyOffline.val ? ' offline' : '') + (kaffe.buddySelected(it) ? ' selected' : '') + ((kaffe.buddyBadges[it.Id!].val) ? ' badged' : '') + ((kaffe.buddyBadgesAlt[it.Id!].val) ? ' badged-alt' : '')),
                'data-badge': depends(() => (kaffe.buddyBadges[it.Id!].val) || ""),
                'data-badge-alt': depends(() => (kaffe.buddyBadgesAlt[it.Id!].val) || ""),
            }, htm.div(userDomAttrsBuddy(it)))
            item.onclick = () => {
                if (!kaffe.isSeeminglyOffline.val)
                    kaffe.buddySelected(it, true)
            }
            return item
        }),
        DOM: htm.div({ 'class': 'kaffe-buddies' },
            htm.div({
                'class': depends(() => 'buddy-self' + ((kaffe.selectedBuddy.val === 0) ? ' selected' : '') + (kaffe.isSeeminglyOffline.val ? ' offline' : '') + ((kaffe.buddyBadges[0].val) ? ' badged' : '') + ((kaffe.buddyBadgesAlt[0].val) ? ' badged-alt' : '')),
                'data-badge': depends(() => (kaffe.buddyBadges[0].val) || ""),
                'data-badge-alt': depends(() => (kaffe.buddyBadgesAlt[0].val) || ""),
                'onclick': () => {
                    if (!kaffe.isSeeminglyOffline.val)
                        kaffe.buddySelected(undefined, true)
                },
            }, htm.div(userDomAttrsSelf())),
        ),
    }

    van.add(me.DOM,
        me.buddies.domNode,
        htm.div({
            'style': depends(() => kaffe.userSelf.val ? '' : 'display:none'),
            'class': depends(() => 'buddy' + (me.buddyRequestsBy.val.length ? ' badged' : '') + (kaffe.isSeeminglyOffline.val ? ' offline' : '')),
            'data-badge': depends(() => me.buddyRequestsBy.val.length || ""),
            'onclick': () => { if (!kaffe.isSeeminglyOffline.val) showBuddiesDialog(me) },
        },
            htm.div({ 'class': depends(() => 'buddy-pic' + (kaffe.isSeeminglyOffline.val ? ' offline' : '')), 'title': "Manage buddies", 'style': `background-image: url('${userPicFileUrl(undefined, "👥")}')` })),
    )
    return me
}


export function userPicFileUrl(user?: yo.User, fallBackToEmoji = "🦜", toRoundedSvgFavIcon = false) {
    if (user && !user.Account)
        fallBackToEmoji = "👤"
    if (!(user && user.PicFileId))
        return util.svgTextIconDataHref(fallBackToEmoji)
    return '/_postfiles/' + user.PicFileId + (toRoundedSvgFavIcon ? ('?picRounded=' + user.PicFileId + '.svg') : '')
}

export function userDomAttrsBuddy(user?: yo.User, userIdHint?: number) {
    if (!user)
        return {
            'class': 'buddy-pic offline',
            'title': `(ex-buddy #${userIdHint ?? -1} — or bug)`,
            'style': `background-image: url('${userPicFileUrl()}')`,
        }
    if (!user.Account) //
        return {
            'class': 'buddy-pic offline',
            'title': `${user.Nick} — (buddy request still pending)`,
            'style': `background-image: url('${userPicFileUrl(user)}')`,
        }
    return {
        'class': depends(() => {
            return 'buddy-pic' + ((kaffe.isSeeminglyOffline.val || user.Offline) ? ' offline' : '')
        }),
        'title': `${user.Nick}${((!user.Btw) ? '' : (' — ' + user.Btw))}`,
        'style': `background-image: url('${userPicFileUrl(user)}')`,
    }
}

export function userDomAttrsSelf() {
    return {
        'class': depends(() => 'buddy-pic self' + (kaffe.isSeeminglyOffline.val ? ' offline' : '')),
        'title': depends(() => {
            const user_self = kaffe.userSelf.val
            return (!user_self) ? "(you)" : `${user_self.Nick}${((!user_self.Btw) ? '' : (' — ' + user_self.Btw))}`
        }),
        'style': depends(() => {
            const user_self = kaffe.userSelf.val
            return `background-image: url('${userPicFileUrl(user_self)}')`
        }),
    }
}

function update(me: UiCtlBuddies, buddiesInfo: yo.__userBuddies_Out) {
    me.buddyRequestsBy.val = buddiesInfo.BuddyRequestsBy ?? []
    const buddies = buddiesInfo.Buddies ?? []
    const offline_buddies = buddies.filter(_ => _.Offline)
    let have_changes = false
    for (const user of offline_buddies) {
        const old = me.buddies.all.find(_ => (_.Id === user.Id))
        if (old && (old.Offline !== user.Offline))
            have_changes = true
    }

    const move_selected_top = false // actually irritating ux-wise, so decided-against for now
    if (move_selected_top) {
        const is_selected: { [_: number]: boolean } = {}
        for (let i = 0; i < buddies.length; i++) {
            const buddy = buddies[i]
            is_selected[buddy.Id!] = kaffe.buddySelected(buddy)
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

    if (have_changes || !youtil.deepEq(buddies, me.buddies.all.filter(_ => true), false, false))
        me.buddies.replaceWith(buddies)
}

async function showBuddiesDialog(me: UiCtlBuddies) {
    const add_new_buddy = async () => {
        let nick_or_email_addr = prompt("New buddy's nick or email address?", "")
        if (nick_or_email_addr && nick_or_email_addr.length && (nick_or_email_addr = nick_or_email_addr.trim())) {
            if (me.buddies.all.filter(_ => true).some(_ => _.Nick === nick_or_email_addr))
                alert(`A buddy request for '${nick_or_email_addr}' had already been placed at an earlier time, but your eagerness is commendable.`)
            else
                try {
                    const result = await yo.api__userBuddiesAdd({ NickOrEmailAddr: nick_or_email_addr })
                    if (!result.Done)
                        alert(`No user '${nick_or_email_addr}' was found. Typo?`)
                    else
                        alert(`Done. Until '${nick_or_email_addr}' confirms, they'll appear offline in your buddy list.`)
                } catch (err) {
                    if (!kaffe.knownErr<yo.__userBuddiesAddErr>(err, kaffe.handleKnownErrMaybe<yo.__userBuddiesAddErr>))
                        kaffe.onErrOther(err, true)
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
            await kaffe.reloadUserSelf()
            const user_self = kaffe.userSelf?.val
            if (!user_self)
                return
            if ((!user_self.Buddies) || !user_self.Buddies.some(_ => (_ === user.Id!))) {
                await yo.api__userUpdate({
                    Id: user_self.Id, Changes: { Buddies: [user.Id!].concat(user_self.Buddies ?? []) }, ChangedFields: ['Buddies']
                }, new FormData())
            }
            htmCheckbox.checked = true
            alert(`You're now buddies with '${user.Nick ?? '?'}', go chat them up!`)
            await kaffe.reloadUserSelf()
        } catch (err) {
            if (!kaffe.knownErr<yo.__userUpdateErr>(err, kaffe.handleKnownErrMaybe<yo.__userUpdateErr>))
                kaffe.onErrOther(err, true)
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
