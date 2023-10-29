import van, { State } from '../__yostatic/vanjs/van-1.2.3.debug.js'
const htm = van.tags

import * as yo from './yo-sdk.js'
import * as uibuddies from './ui/buddies.js'
import * as uiposts from './ui/posts.js'
import * as uiuserpopup from './ui/user_popup.js'


const fetchBuddiesIntervalMs = 4321
const fetchPostsDeletedIntervalMs = 6789
const fetchPostsIntervalMsWhenVisible = 2345
const fetchPostsIntervalMsWhenHidden = 4321
let fetchPostsIntervalMsCur = fetchPostsIntervalMsWhenVisible
let fetchPostsSinceDt: string | undefined
let fetchesPaused = false // true while signed out
let fetchedPostsEverYet = false
export let userSelf = van.state(undefined as (yo.User | undefined))
export let browserTabInvisibleSince = 0
export let isSeeminglyOffline = van.state(false)
export let selectedBuddy: State<number> = van.state(0)
export let buddyBadges: { [_: number]: State<string> } = { 0: van.state("") }

let uiDialogLogin = newUiLoginDialog()
let uiBuddies: uibuddies.UiCtlBuddies = uibuddies.create()
let uiPosts: uiposts.UiCtlPosts = uiposts.create()

export function main() {
    document.onvisibilitychange = () => {
        const is_hidden = ((document.visibilityState === 'hidden') || document.hidden), now = new Date().getTime()
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
    )
    uiuserpopup.setLiveDarklite()
    setTimeout(fetchBuddies, 234)
}

export async function reloadUserSelf() {
    const user_self = await yo.apiUserBy({ EmailAddr: yo.userEmailAddr })
    userSelf.val = user_self
    return user_self
}

async function fetchBuddies(oneOff?: boolean) {
    if (fetchesPaused && !oneOff)
        return
    try {
        const buddies = (await yo.apiUserBuddies())!.Result ?? []
        for (const user of buddies)
            if (!buddyBadges[user.Id!])
                buddyBadges[user.Id!] = van.state("")
        uiBuddies.update(buddies)
        if (!userSelf.val)
            reloadUserSelf()
        isSeeminglyOffline.val = false
        browserTabTitleRefresh()
        if (!fetchedPostsEverYet) {
            setTimeout(fetchPostsRecent, 123)
            setTimeout(fetchPostsDeleted, fetchPostsDeletedIntervalMs)
        }
    } catch (err) {
        if (!knownErr<yo.UserBuddiesErr>(err, handleKnownErrMaybe<yo.UserBuddiesErr>))
            onErrOther(err)
    }

    if ((!fetchesPaused) && !oneOff)
        setTimeout(fetchBuddies, fetchBuddiesIntervalMs)
}

async function fetchPostsRecent(oneOff?: boolean) {
    if ((uiPosts.isDeleting.val > 0) || (fetchesPaused && !oneOff))
        return
    try {
        const recent_updates = await yo.apiPostsRecent({
            Since: fetchPostsSinceDt,
            OnlyBy: selectedBuddy.val ? [selectedBuddy.val] : [],
        })
        isSeeminglyOffline.val = false
        fetchedPostsEverYet = true // even if empty, we have a non-error outcome and so set this
        if (uiPosts.isDeleting.val === 0) {
            fetchPostsSinceDt = recent_updates.NextSince
            uiposts.update(uiPosts, recent_updates?.Posts ?? [])
            browserTabTitleRefresh()
        }
        if (recent_updates.UnreadCounts)
            for (const buddy_id_str in recent_updates.UnreadCounts) {
                const buddy_id = (buddy_id_str === "") ? 0 : parseInt(buddy_id_str), num_unread = recent_updates.UnreadCounts[buddy_id_str]
                const state = buddyBadges[buddy_id], badge_text = ((num_unread <= 0) ? "" : num_unread.toString())
                if (state)
                    state.val = badge_text
                else
                    buddyBadges[buddy_id] = van.state(badge_text)
            }
    } catch (err) {
        if (!knownErr<yo.PostsRecentErr>(err, handleKnownErrMaybe<yo.PostsRecentErr>))
            onErrOther(err)
    }
    if ((!fetchesPaused) && !oneOff)
        setTimeout(fetchPostsRecent, fetchPostsIntervalMsCur)
}

async function fetchPostsDeleted() {
    if (fetchesPaused)
        return
    const post_ids = uiPosts.posts.filter(_ => true).map(_ => _.Id!)
    if (post_ids.length) try {
        const post_ids_deleted = (await yo.apiPostsDeleted({ OutOfPostIds: post_ids })).DeletedPostIds
        isSeeminglyOffline.val = false
        if (post_ids_deleted && post_ids_deleted.length)
            uiposts.update(uiPosts, uiPosts.posts.filter(_ => !post_ids_deleted.includes(_.Id!)), false, post_ids_deleted)
    } catch (err) {
        if (!knownErr<yo.PostsDeletedErr>(err, handleKnownErrMaybe<yo.PostsDeletedErr>))
            onErrOther(err)
    }
    if (!fetchesPaused)
        setTimeout(fetchPostsDeleted, fetchPostsDeletedIntervalMs)
}

function newUiLoginDialog() {
    const on_btn_login = async () => {
        try {
            await yo.apiUserSignIn({ EmailAddr: in_user_name.value, PasswordPlain: in_password.value })
            isSeeminglyOffline.val = false
            location.reload()
        } catch (err) {
            if (!knownErr<yo.UserSignInErr>(err, (err) => {
                switch (err) {
                    case '___yo_authLogin_WrongPassword':
                    case '___yo_authLogin_AccountDoesNotExist':
                    case '___yo_authLogin_EmailInvalid':
                        alert(err)
                        return true
                }
                return false
            }))
                onErrOther(err)
        }
    }

    const in_user_name = htm.input({ 'value': 'foo789@bar.baz' })
    const in_password = htm.input({ 'type': 'password', 'value': 'foobar' })
    return htm.dialog({},
        htm.form({},
            in_user_name,
            in_password,
            htm.button({ 'onclick': on_btn_login, 'type': 'button' }, "Login"),
        ),
    )
}

export function userById(id: number) {
    if (id <= 0)
        return undefined
    const user_self = userSelf.val
    if (user_self && (user_self.Id === id))
        return user_self
    return uiBuddies.buddies.find(_ => (_.Id === id))
}
export function userByPost(post?: yo.Post) {
    const user_self = userSelf.val
    if ((!post) || (user_self && (user_self.Id === post.By)))
        return user_self
    return uiBuddies.buddies.find(_ => (_.Id === post.By))
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
        const resp = await yo.apiPostNew({
            By: user_self.Id,
            To: (!selectedBuddy.val) ? [] : [selectedBuddy.val],
            Htm: html,
        }, form_data)
        isSeeminglyOffline.val = false
        ok = (resp.Result > 0)
    } catch (err) {
        if (!knownErr<yo.PostNewErr>(err, handleKnownErrMaybe<yo.PostNewErr>))
            onErrOther(err)
    }
    if (ok)
        fetchPostsRecent(true) // async but here we dont care to await
    return ok
}

export async function deletePost(id: number) {
    let ok = false
    try {
        await yo.apiPostDelete({ Id: id })
        isSeeminglyOffline.val = false
        ok = true
    } catch (err) {
        if (!knownErr<yo.PostDeleteErr>(err, handleKnownErrMaybe<yo.PostDeleteErr>))
            onErrOther(err)
    }
    if (ok)
        fetchPostsRecent(true) // async but here we dont care to await
    return ok
}

function browserTabTitleRefresh() {
    const user_self = userSelf.val, user_buddy = userById(selectedBuddy.val)
    const buddies_and_self = (user_buddy && user_self) ? [user_buddy, user_self]
        : uiBuddies.buddies.concat(user_self ? [user_self] : [])
    const new_title = ((isSeeminglyOffline.val ? '(disconnected)' : ((uiPosts.numFreshPosts === 0) ? '' : `(${uiPosts.numFreshPosts})`))
        + ' ' + (buddies_and_self.map(_ => _.Nick).join(', '))).trim()
    if (new_title !== document.title)
        document.title = new_title
    const fav_icon_user = isSeeminglyOffline.val ? undefined : buddies_and_self.find(_ => (_.PicFileId !== ''))
    const fav_icon_href = uibuddies.userPicFileUrl(fav_icon_user, '☕', true), htm_favicon = document.getElementById('favicon') as HTMLLinkElement
    if (htm_favicon && htm_favicon.href && (htm_favicon.href !== fav_icon_href))
        htm_favicon.href = fav_icon_href
}

export function buddySelected(user?: yo.User, ensureIsSelected?: boolean): boolean {
    let is_selected = (selectedBuddy.val === ((user?.Id) ?? 0))
    if (ensureIsSelected)
        if (!is_selected) {
            selectedBuddy.val = ((user?.Id) ?? 0)
            buddyBadges[selectedBuddy.val].val = ""
            is_selected = !is_selected
            uiposts.update(uiPosts, [], true)
            fetchPostsSinceDt = undefined
            fetchPostsRecent(true)
            fetchBuddies(true)
        } else  // already was selected, so the click/tap shows user card
            userShowPopup(user)
    return is_selected
}

export function userShowPopup(user?: yo.User) {
    if (!user)
        user = userSelf.val
    if (user) {
        const popup = uiuserpopup.create(user ?? userSelf.val)
        van.add(document.body, popup.DOM)
        popup.DOM.showModal()
    }
}


export function onErrOther(err: any, showAlert?: boolean) {
    if (isSeeminglyOffline.val = !showAlert)
        browserTabTitleRefresh()
    const err_json = JSON.stringify(err), err_str_1 = err.toString(), err_str_2 = `${err}`,
        err_msg = err.message ? err.message :
            ((err_str_1 && err_str_1 !== '[object Object]') ? err_str_1 :
                ((err_str_2 && err_str_2 !== '[object Object]') ? err_str_2 : err_json))
    if (showAlert)
        alert(err_msg)
    else
        console.warn(err, err_json, err_msg)
}
export function knownErr<T extends string>(err: any, ifSo: (_: T) => boolean): boolean {
    const yo_err = err as yo.Err<T>
    return yo_err && yo_err.knownErr && (yo_err.knownErr.length > 0) && ifSo(yo_err.knownErr)
}
export function handleKnownErrMaybe<T extends string>(err: T): boolean {
    console.log(JSON.stringify(err))
    switch (err) {
        case 'Unauthorized':
            fetchesPaused = true
            uiDialogLogin.showModal()
            return true
        case 'MissingOrExcessiveContentLength':
            alert("To share something over " + yo.reqMaxReqPayloadSizeMb + "MB, host it elsewhere and share the link instead.")
            return true
        case 'UserUpdate_NicknameAlreadyExists':
            alert(`Nickname already taken — but, look... '${userSelf.val?.Nick}' ain't so shabby either!`)
            return true
    }
    return false
}
