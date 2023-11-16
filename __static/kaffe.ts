import van, { State } from '../__yostatic/vanjs/van-1.2.6.js'
const htm = van.tags

import * as yo from './yo-sdk.js'
import * as youtil from '../__yostatic/util.js'
import * as uibuddies from './ui/buddies.js'
import * as uiposts from './ui/posts.js'
import * as uiuserpopup from './ui/user_popup.js'
import * as uilogindialog from './ui/login_dialog.js'


const fetchPostsIntervalMsWhenVisible = 2345
const fetchPostsIntervalMsWhenHidden = 4321
const fetchPostsDeletedIntervalMs = fetchPostsIntervalMsWhenHidden * 2
const fetchBuddiesIntervalMs = 4321
let fetchPostsIntervalMsCur = fetchPostsIntervalMsWhenVisible
let fetchPostsSinceDt: string | undefined
let fetchesPaused = false // true while signed out
let fetchedPostsEverYet = false
export let userSelf = van.state(undefined as (yo.User | undefined))
export let browserTabInvisibleSince = 0
export let isSeeminglyOffline = van.state(false)
export let isArchiveBrowsing = van.state(false)
export let selectedBuddy: State<number> = van.state(0)
export let buddyBadges: { [_: number]: State<string> } = { 0: van.state("") }
export let buddyBadgesAlt: { [_: number]: State<string> } = { 0: van.state("") }
export let signUpOrPwdForgotNotice = van.state("")
let firstOfMonth = new Date(new Date().getFullYear(), new Date().getUTCMonth(), 1, 0, 0, 0, 0).getTime()
let shouldReloadPostPeriods = true

let uiDialogLogin = uilogindialog.create(_ => signUpOrPwdForgotNotice.val = _)
let uiBuddies: uibuddies.UiCtlBuddies = uibuddies.create()
let uiPosts: uiposts.UiCtlPosts = uiposts.create()
let uiPeriodPicker: HTMLSelectElement =
    htm.select({
        'class': 'dtsel',
        'onchange': () => {
            isArchiveBrowsing.val = (uiPeriodPicker.selectedIndex > 0)
            uiposts.update(uiPosts, [], true)
            fetchPostsRecent(true) // no await needed
        },
    },
        htm.option({ 'value': '' },
            "⏶\xa0\xa0 Fresh")
    )

export function main() {
    document.onvisibilitychange = () => {
        const is_hidden = ((document.visibilityState === 'hidden') || document.hidden), now = Date.now()
        const became_visible = (!is_hidden) && (browserTabInvisibleSince !== 0)
        browserTabInvisibleSince = (!is_hidden) ? 0 : ((browserTabInvisibleSince === 0) ? now : browserTabInvisibleSince)
        fetchPostsIntervalMsCur = is_hidden ? fetchPostsIntervalMsWhenHidden : fetchPostsIntervalMsWhenVisible
        if (became_visible && (uiPosts.numFreshPosts === 0))
            fetchPostsRecent(true)
    }
    van.add(document.body,
        uiPosts.DOM,
        uiBuddies.DOM,
        uiDialogLogin,
        uiPeriodPicker,
    )
    uiuserpopup.setLiveDarklite()
    setTimeout(fetchBuddies, 123)
}

export async function reloadUserSelf() {
    if (yo.userEmailAddr) {
        userSelf.val = undefined // in case of failure, a later buddy-fetch will re-attempt a fresh reload — but only with this assignment in place
        const user_self = (await yo.api__userBy({ EmailAddr: yo.userEmailAddr })).Result
        isSeeminglyOffline.val = false
        userSelf.val = user_self
        buddyBadgesAlt[0].val = user_self.BtwEmoji ?? ""
    }
}

async function fetchBuddies(oneOff?: boolean) {
    if (fetchesPaused && !oneOff)
        return
    try {
        const result = await yo.api__userBuddies()
        isSeeminglyOffline.val = false
        if (!userSelf.val) // fetch only *after* the above because apiUserBy needs cur-user email-addr, which isn't cookied (but headered)
            reloadUserSelf() // no need to await really
        if (result.Buddies && result.Buddies.length)
            for (const user of result.Buddies) {
                if (!buddyBadges[user.Id!])
                    buddyBadges[user.Id!] = van.state("")
                if (!buddyBadgesAlt[user.Id!])
                    buddyBadgesAlt[user.Id!] = van.state("")
                buddyBadgesAlt[user.Id!].val = user.BtwEmoji ?? ""
            }
        uiBuddies.update(result)
        browserTabTitleRefresh()
        if (!fetchedPostsEverYet) {
            fetchPostsRecent()
            setTimeout(fetchPostsDeleted, fetchPostsDeletedIntervalMs)
        }
    } catch (err) {
        if (!knownErr<yo.__userBuddiesErr>(err, handleKnownErrMaybe<yo.__userBuddiesErr>))
            onErrOther(err)
    }

    if ((!fetchesPaused) && !oneOff)
        setTimeout(fetchBuddies, fetchBuddiesIntervalMs)
}

async function fetchPostsRecent(oneOff?: boolean) {
    if ((uiPosts.isRequestingDeletion.val > 0) || (fetchesPaused && !oneOff))
        return
    try {
        if (shouldReloadPostPeriods) {
            shouldReloadPostPeriods = false
            uiPeriodPicker.selectedIndex = 0
            while (uiPeriodPicker.options.length > 1)
                uiPeriodPicker.options.remove(1)
            isArchiveBrowsing.val = false
            const periods = (await yo.api__postMonthsUtc({ WithUserIds: selectedBuddy.val ? [selectedBuddy.val] : [] })).Periods ?? []
            for (const period of periods) {
                const dt = new Date(period.Year!, (period.Month!) - 1, 1, 0, 0, 0, 0)
                uiPeriodPicker.options.add(htm.option({ 'value': JSON.stringify(period) }, `${dt.getFullYear()} — ${dt.toLocaleDateString('default', { month: 'long' })}`))
            }
        }

        const fetch_archived_posts = (uiPeriodPicker.selectedIndex > 0)
        if (fetch_archived_posts) {
            const period: yo.YearAndMonth = JSON.parse(uiPeriodPicker.selectedOptions[0].value)
            const dt = new Date(period.Year!, (period.Month!) - 1, 1, 0, 0, 0, 0)
            const is_earlier_month_than_current = (dt.getTime() < firstOfMonth)
            if ((!oneOff) && (uiPosts.posts.all.length > 0) && is_earlier_month_than_current)
                return // the month selected is before the current month and was already fetched.
            else if (!is_earlier_month_than_current) // current-month while in archive view...
                oneOff = false // ...so ensuring interval activates at the end of this function (in case user just jumped here from an earlier month)
        }
        const result = fetch_archived_posts
            ? await yo.api__postsForMonthUtc({
                OnlyBy: selectedBuddy.val ? [selectedBuddy.val] : [],
                Period: JSON.parse(uiPeriodPicker.selectedOptions[0].value),
            })
            : await yo.api__postsRecent({
                OnlyBy: selectedBuddy.val ? [selectedBuddy.val] : [],
                Since: fetchPostsSinceDt,
            })
        isSeeminglyOffline.val = false
        fetchedPostsEverYet = true // even if empty, we have a non-error outcome and so set this
        if (uiPosts.isRequestingDeletion.val === 0) {
            fetchPostsSinceDt = result.NextSince
            uiposts.update(uiPosts, result?.Posts ?? [])
            browserTabTitleRefresh()
        }
        if (result.UnreadCounts)
            for (const buddy_id_str in result.UnreadCounts) {
                const buddy_id = (buddy_id_str === "") ? 0 : parseInt(buddy_id_str), num_unread = result.UnreadCounts[buddy_id_str]
                const state = buddyBadges[buddy_id], badge_text = ((num_unread <= 0) ? "" : num_unread.toString())
                if (state)
                    state.val = badge_text
                else
                    buddyBadges[buddy_id] = van.state(badge_text)
            }
    } catch (err) {
        if (!knownErr<yo.__postsRecentErr>(err, handleKnownErrMaybe<yo.__postsRecentErr>))
            onErrOther(err)
    }
    if ((!fetchesPaused) && !oneOff)
        setTimeout(fetchPostsRecent, fetchPostsIntervalMsCur)
}

async function fetchPostsDeleted() {
    if (fetchesPaused)
        return
    const post_ids = uiPosts.posts.all.filter(_ => true).map(_ => _.Id!)
    if (post_ids.length)
        try {
            const post_ids_deleted = (await yo.api__postsDeleted({ OutOfPostIds: post_ids })).DeletedPostIds
            isSeeminglyOffline.val = false
            if (post_ids_deleted && post_ids_deleted.length)
                uiposts.update(uiPosts, uiPosts.posts.all.filter(_ => !post_ids_deleted.includes(_.Id!)), false, post_ids_deleted)
        } catch (err) {
            if (!knownErr<yo.__postsDeletedErr>(err, handleKnownErrMaybe<yo.__postsDeletedErr>))
                onErrOther(err)
        }
    if (!fetchesPaused)
        setTimeout(fetchPostsDeleted, fetchPostsDeletedIntervalMs)
}

export function userById(id: number) {
    if (id <= 0)
        return undefined
    const user_self = userSelf.val
    if (user_self && (user_self.Id === id))
        return user_self
    return uiBuddies.buddies.all.find(_ => (_.Id === id))
}
export function userByPost(post: yo.Post) {
    const user_self = userSelf.val
    if (user_self && (user_self.Id === post.By))
        return user_self
    return uiBuddies.buddies.all.find(_ => (_.Id === post.By))
}

export async function sendNewPost(html: string, files?: File[]) {
    const user_self = userSelf.val
    if (!user_self)
        return false
    let ok = false
    try {
        const form_data = new FormData()
        if (files && files.length)
            for (const file of files)
                form_data.append('files', file)
        const resp = await yo.api__postNew({
            NewPost: {
                By: user_self.Id,
                To: (!selectedBuddy.val) ? [] : [selectedBuddy.val],
                Htm: html,
            }
        }, form_data)
        isSeeminglyOffline.val = false
        ok = (resp.Result > 0)
    } catch (err) {
        if (!knownErr<yo.__postNewErr>(err, handleKnownErrMaybe<yo.__postNewErr>))
            onErrOther(err)
    }
    if (ok)
        fetchPostsRecent(true) // async but here we dont care to await
    return ok
}

export async function deletePost(id: number) {
    let ok = false
    try {
        await yo.api__postDelete({ Id: id })
        isSeeminglyOffline.val = false
        ok = true
    } catch (err) {
        if (!knownErr<yo.__postDeleteErr>(err, handleKnownErrMaybe<yo.__postDeleteErr>))
            onErrOther(err)
    }
    if (ok)
        fetchPostsRecent(true) // async but here we dont care to await
    return ok
}

function browserTabTitleRefresh() {
    const user_self = userSelf.val, user_buddy = userById(selectedBuddy.val)
    const buddies_and_self = (user_buddy && user_self) ? [user_buddy, user_self]
        : uiBuddies.buddies.all.concat(user_self ? [user_self] : [])
    const new_title = ((isSeeminglyOffline.val ? '(disconnected)' : ((uiPosts.numFreshPosts === 0) ? '' : `(${uiPosts.numFreshPosts})`))
        + ' ' + (buddies_and_self.map(_ => _.Nick).join(', '))).trim()
    if (new_title !== document.title)
        document.title = new_title
    const fav_icon_user = isSeeminglyOffline.val ? undefined : buddies_and_self.find(_ => ((_.PicFileId !== '') || selectedBuddy.val))
    const fav_icon_href = uibuddies.userPicFileUrl(fav_icon_user, fav_icon_user ? undefined : '☕', true), htm_favicon = document.getElementById('favicon') as HTMLLinkElement
    if (htm_favicon && htm_favicon.href && (htm_favicon.href !== fav_icon_href))
        htm_favicon.href = fav_icon_href
}

export function buddySelected(user?: yo.User, ensureIsSelected?: boolean) {
    let is_selected = (selectedBuddy.val === ((user?.Id) ?? 0))
    if (ensureIsSelected)
        if ((!is_selected) && ((!user) || (user.Auth))) { // pending buddy-req `User`s have no fields other than `Nick` set, so their `Auth` will be 0
            selectedBuddy.val = ((user?.Id) ?? 0)
            fetchPostsSinceDt = undefined
            buddyBadges[selectedBuddy.val].val = ""
            is_selected = !is_selected
            uiposts.update(uiPosts, [], true)
            shouldReloadPostPeriods = true
            fetchPostsRecent(true) // no await needed
            fetchBuddies(true) // no await needed
        } else  // already was selected, so the click/tap shows user popup
            userShowPopup(user)
    return is_selected
}

export async function userSignOut(confirmFirst: boolean) {
    if (confirmFirst && !confirm("Sure to sign out now?"))
        return
    try {
        await yo.api__userSignOut({})
        location.reload()
    } catch (err) {
        if (confirm('Failed to successfully sign out (at the server side), you can clear the Cookies for this domain or: try again?'))
            userSignOut(false)
    }
}

export function userShowPopup(user?: yo.User) {
    if (!user)
        user = userSelf.val
    if (user) {
        const popup = uiuserpopup.create(user)
        van.add(document.body, popup.DOM)
        popup.DOM.showModal()
    }
}


export function onErrOther(err: any, showAlert?: boolean) {
    if (isSeeminglyOffline.val = !showAlert)
        browserTabTitleRefresh()
    const err_msg = youtil.errStr(err)
    if (showAlert)
        alert("Try again shortly, because this attempt errored with: " + err_msg)
    else
        console.warn(err, JSON.stringify(err), err_msg)
}
export function knownErr<T extends string>(err: any, ifSo: (_: T) => boolean): boolean {
    const yo_err = err as yo.Err<T>
    return yo_err && yo_err.knownErr && (yo_err.knownErr.length > 0) && ifSo(yo_err.knownErr)
}
export function handleKnownErrMaybe<T extends string>(err: T): boolean {
    switch (err) {
        case 'DbUpdate_ExpectedChangesForUpdate': // classical double-click-the-Save-button handling (although we disable them here)
            return true
        case 'Unauthorized':
            fetchesPaused = true
            uiDialogLogin.showModal()
            return true
        case 'MissingOrExcessiveContentLength':
            alert("To share something over " + yo.reqMaxReqMultipartSizeMb + "MB, host it elsewhere and share the link instead.")
            return true
        case 'UserUpdate_NicknameAlreadyExists':
            alert(`Nick already taken — but, look... '${userSelf.val?.Nick}' ain't so shabby either!`)
            return true
        case 'UserUpdate_ExpectedNonEmptyNickname':
            alert("That choice of nick does not reflect you: you're not empty.")
            return true
    }
    return false
}
