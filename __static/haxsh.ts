import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
const htm = van.tags

import * as yo from './yo-sdk.js'
import * as uibuddies from './ui/buddies.js'
import * as uiposts from './ui/posts.js'

const none = void 0

const fetchBuddiesIntervalMs = 4321
let fetchPostsSinceDt: string | undefined
const fetchPostsIntervalMsWhenVisible = 2345
const fetchPostsIntervalMsWhenHidden = 4321
let fetchPostsIntervalMsCur = fetchPostsIntervalMsWhenVisible
let fetchesPaused = false // true while signed out
let fetchedPostsEverYet = false
export let userSelf = van.state(undefined as (yo.User | undefined))
export let browserTabInvisibleSince = 0

let uiDialogLogin = newUiLoginDialog()
let uiBuddies: uibuddies.UiCtlBuddies = uibuddies.create()
let uiPosts: uiposts.UiCtlPosts = uiposts.create(getUserByPost, sendPost)

export function main() {
    document.onvisibilitychange = () => {
        const is_hidden = ((document.visibilityState === 'hidden') || document.hidden), now = new Date().getTime()
        const became_visible = (!is_hidden) && (browserTabInvisibleSince !== 0)
        browserTabInvisibleSince = (!is_hidden) ? 0 : ((browserTabInvisibleSince === 0) ? now : browserTabInvisibleSince)
        fetchPostsIntervalMsCur = is_hidden ? fetchPostsIntervalMsWhenHidden : fetchPostsIntervalMsWhenVisible
        if (became_visible && (uiPosts.numFreshPosts === 0))
            fetchPosts(true)
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
        browserTabTitleRefresh()
        if (!fetchedPostsEverYet)
            setTimeout(fetchPosts, 123)
    } catch (err) {
        if (!knownErr<yo.UserBuddiesErr>(err, handleKnownErrMaybe<yo.UserBuddiesErr>))
            onErrOther(err)
    }

    if (!fetchesPaused)
        setTimeout(fetchBuddies, fetchBuddiesIntervalMs)
}

async function fetchPosts(oneOff?: boolean) {
    if (fetchesPaused && !oneOff)
        return
    try {
        const recent_updates = await yo.apiPostsRecent({ Since: fetchPostsSinceDt ? fetchPostsSinceDt : none })
        fetchedPostsEverYet = true // even if empty, we have a non-error outcome and so set this
        fetchPostsSinceDt = recent_updates.Next
        uiPosts.update(recent_updates?.Posts ?? [])
        browserTabTitleRefresh()
    } catch (err) {
        if (!knownErr<yo.PostsRecentErr>(err, handleKnownErrMaybe<yo.PostsRecentErr>))
            onErrOther(err)
    }
    if ((!fetchesPaused) && !oneOff)
        setTimeout(fetchPosts, fetchPostsIntervalMsCur)
}

function newUiLoginDialog() {
    const on_btn_login = async () => {
        try {
            await yo.apiUserSignIn({ EmailAddr: in_user_name.value, PasswordPlain: in_password.value })
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


function getUserByPost(post?: yo.Post) {
    const user_self = userSelf.val
    if ((!post) || (user_self && (user_self.Id === post.By)))
        return user_self
    return uiBuddies.buddies.find(_ => (_.Id === post.By))
}

async function sendPost(html: string, files?: string[]) {
    const user_self = userSelf.val
    if (!user_self)
        return false
    const resp = await yo.apiPostNew({
        By: user_self.Id,
        To: [],
        Files: files ?? [],
        Htm: html,
    })
    const ok = (resp.Result > 0)
    if (ok)
        fetchPosts(true) // async but here we dont care to await
    return ok
}

function browserTabTitleRefresh() {
    const user_self = userSelf.val
    const buddies_and_self = uiBuddies.buddies.concat(user_self ? [user_self] : [])
    const new_title = (((uiPosts.numFreshPosts === 0) ? '' : `(${uiPosts.numFreshPosts}) `) + buddies_and_self.map(_ => _.Nick).join(', ') || '...')
    if (new_title !== document.title)
        document.title = new_title
    const fav_icon_user = buddies_and_self.find(_ => (_.PicFileId !== ''))
    const fav_icon_href = uibuddies.userPicFileUrl(fav_icon_user, '☕', true), htm_favicon = document.getElementById('favicon') as HTMLLinkElement
    if (htm_favicon && htm_favicon.href && (htm_favicon.href !== fav_icon_href))
        htm_favicon.href = fav_icon_href
}


function onErrOther(err: any) { console.error(`${err}`, err, JSON.stringify(err)) }
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
    }
    return false
}
