import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
const htm = van.tags

import * as yo from './yo-sdk.js'
import * as uibuddies from './ui/buddies.js'
import * as uiposts from './ui/posts.js'


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
export let selectedBuddies: yo.User[] = []
export let haveAnySelected = van.state(false)

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
    setTimeout(fetchBuddies, 234)
}

async function fetchBuddies() {
    if (fetchesPaused)
        return

    try {
        const buddies = (await yo.apiUserBuddies())!.Result ?? []
        uiBuddies.update(buddies)
        let user_self = userSelf.val
        if (!user_self)
            userSelf.val = (user_self = await yo.apiUserBy({ EmailAddr: yo.userEmailAddr }))
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

    if (!fetchesPaused)
        setTimeout(fetchBuddies, fetchBuddiesIntervalMs)
}

async function fetchPostsRecent(oneOff?: boolean) {
    if ((uiPosts.isDeleting.val > 0) || (fetchesPaused && !oneOff))
        return
    try {
        const recent_updates = await yo.apiPostsRecent({
            Since: fetchPostsSinceDt,
            OnlyBy: selectedBuddies.map(_ => _.Id),
        })
        isSeeminglyOffline.val = false
        fetchedPostsEverYet = true // even if empty, we have a non-error outcome and so set this
        if (uiPosts.isDeleting.val === 0) {
            fetchPostsSinceDt = recent_updates.Next
            uiposts.update(uiPosts, recent_updates?.Posts ?? [])
            browserTabTitleRefresh()
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


export function getUserByPost(post?: yo.Post) {
    const user_self = userSelf.val
    if ((!post) || (user_self && (user_self.Id === post.By)))
        return user_self
    return uiBuddies.buddies.find(_ => (_.Id === post.By))
}

export async function sendNewPost(html: string, files?: string[]) {
    const user_self = userSelf.val
    if (!user_self)
        return false
    let ok = false
    try {
        const resp = await yo.apiPostNew({
            By: user_self.Id,
            To: selectedBuddies.map(_ => _.Id),
            Files: files ?? [],
            Htm: html,
        })
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
    const user_self = userSelf.val
    const buddies_and_self = uiBuddies.buddies.concat(user_self ? [user_self] : [])
    const new_title = ((isSeeminglyOffline.val ? '(disconnected)' : ((uiPosts.numFreshPosts === 0) ? '' : `(${uiPosts.numFreshPosts})`))
        + ' ' + (buddies_and_self.map(_ => _.Nick).join(', '))).trim()
    if (new_title !== document.title)
        document.title = new_title
    const fav_icon_user = isSeeminglyOffline.val ? undefined : buddies_and_self.find(_ => (_.PicFileId !== ''))
    const fav_icon_href = uibuddies.userPicFileUrl(fav_icon_user, '☕', true), htm_favicon = document.getElementById('favicon') as HTMLLinkElement
    if (htm_favicon && htm_favicon.href && (htm_favicon.href !== fav_icon_href))
        htm_favicon.href = fav_icon_href
}

export function buddySelected(user: yo.User, toggleIsSelected?: boolean): boolean {
    let is_selected = selectedBuddies.some(_ => (_.Id === user.Id))
    if (toggleIsSelected) {
        if (is_selected)
            selectedBuddies = selectedBuddies.filter(_ => (_.Id !== user.Id))
        else
            selectedBuddies.push(user)
        haveAnySelected.val = (selectedBuddies.length > 0)
        is_selected = !is_selected
        uiposts.update(uiPosts, [], true)
        fetchPostsSinceDt = undefined
        fetchPostsRecent(true)
    }
    return is_selected
}


function onErrOther(err: any) {
    isSeeminglyOffline.val = true
    browserTabTitleRefresh()
    console.warn(`${err}`, err, JSON.stringify(err))
}
function knownErr<T extends string>(err: any, ifSo: (_: T) => boolean): boolean {
    const yo_err = err as yo.Err<T>
    return yo_err && yo_err.knownErr && (yo_err.knownErr.length > 0) && ifSo(yo_err.knownErr)
}
function handleKnownErrMaybe<T extends string>(err: T): boolean {
    switch (err) {
        case 'Unauthorized':
            fetchesPaused = true
            uiDialogLogin.showModal()
            return true
        case 'MissingOrExcessiveContentLength':
            alert("Your input has exceeded the server's maximum permissible payload size.")
            return true
    }
    return false
}
