import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
const htm = van.tags

import * as yo from './yo-sdk.js'
import * as uibuddies from './ui/buddies.js'
import * as uiposts from './ui/posts.js'

const none = void 0

const fetchBuddiesIntervalMs = 12345
let fetchPostsSinceDt: string | undefined
const fetchPostsIntervalMsWhenVisible = 2345
const fetchPostsIntervalMsWhenHidden = 4321
let fetchPostsIntervalMsCur = fetchPostsIntervalMsWhenVisible
let fetchesPaused = false // true while signed out
let fetchedPostsEverYet = false

let uiDialogLogin = newUiLoginDialog()
let uiBuddies: uibuddies.UiCtlBuddies = uibuddies.create()
let uiPosts: uiposts.UiCtlPosts = uiposts.create((post: yo.Post) => {
    const buddy = uiBuddies.buddies.find(_ => (_.Id === post.By))
    if (buddy) {
        if (post.DtMade! > buddy.LastSeen!)
            buddy.LastSeen = post.DtMade!
        if (post.DtMod! > buddy.LastSeen!)
            buddy.LastSeen = post.DtMod!
    }
    return buddy
})

export function main() {
    document.onvisibilitychange = () => {
        fetchPostsIntervalMsCur = ((document.visibilityState == 'hidden') || (document.hidden))
            ? fetchPostsIntervalMsWhenHidden : fetchPostsIntervalMsWhenVisible
        document.title = fetchPostsIntervalMsCur.toString()
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
        const buddies = await yo.apiUserBuddies()
        if (buddies && buddies.Result) {
            uiBuddies.update(buddies.Result)
            if (!fetchedPostsEverYet)
                setTimeout(fetchPosts, 123)
        }
    } catch (err) {
        if (!knownErr<yo.UserBuddiesErr>(err, handleKnownErrMaybe<yo.UserBuddiesErr>))
            onErrOther(err)
    }

    if (!fetchesPaused)
        setTimeout(fetchBuddies, fetchBuddiesIntervalMs)
}

async function fetchPosts() {
    if (fetchesPaused)
        return
    try {
        const recent_updates = await yo.apiPostsRecent({ Since: fetchPostsSinceDt ? fetchPostsSinceDt : none })
        fetchedPostsEverYet = true // even if empty, we have a non-error outcome and so set this
        fetchPostsSinceDt = recent_updates.Next

        if (recent_updates.Posts && recent_updates.Posts.length)
            uiPosts.update(recent_updates.Posts)
    } catch (err) {
        if (!knownErr<yo.PostsRecentErr>(err, handleKnownErrMaybe<yo.PostsRecentErr>))
            onErrOther(err)
    }
    if (!fetchesPaused)
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

    const in_user_name = htm.input({ 'value': 'foo16@bar.baz' })
    const in_password = htm.input({ 'type': 'password', 'value': 'foobar' })
    return htm.dialog({},
        htm.form({},
            in_user_name,
            in_password,
            htm.button({ 'onclick': on_btn_login, 'type': 'button' }, "Login"),
        ),
    )
}


function onErrOther(err: any) { console.error(err) }
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
