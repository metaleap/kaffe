import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
import vanx from '../__yostatic/vanjs/van-x.js'
const htm = van.tags

import * as yo from './yo-sdk.js'
import * as youtil from '../__yostatic/util.js'
import * as uibuddies from './ui/buddies.js'

const none = void 0

const fetchBuddiesIntervalMs = 1234
let fetchPostsSinceDt: string | undefined
const fetchPostsIntervalMsWhenVisible = 2345
const fetchPostsIntervalMsWhenHidden = 4321
let fetchPostsIntervalMsCur = fetchPostsIntervalMsWhenVisible
let fetchesPaused = false // true while signed out

let uiDialogLogin = newUiLoginDialog()
let uiFeedPosts = newUiPostsFeed()
let uiBuddies: uibuddies.UiCtlBuddies = uibuddies.create()

function onErr(err: any) { console.error(JSON.stringify(err)) }
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

export function main() {
    document.onvisibilitychange = () => {
        fetchPostsIntervalMsCur = ((document.visibilityState == 'hidden') || (document.hidden))
            ? fetchPostsIntervalMsWhenHidden : fetchPostsIntervalMsWhenVisible
        document.title = fetchPostsIntervalMsCur.toString()
    }
    van.add(document.body,
        uiFeedPosts,
        uiBuddies.DOM,
        uiDialogLogin,
    )
    setTimeout(fetchBuddies, 234)
    setTimeout(fetchRefresh, 321)
}

async function fetchBuddies() {
    if (fetchesPaused)
        return

    try {
        const buddies = await yo.apiUserBuddies()
        console.log(Array.isArray(buddies.Result), Array.isArray(uiBuddies.buddies), buddies.Result.length, uiBuddies.buddies.length, youtil.deepEq(buddies.Result, uiBuddies.buddies, false), youtil.deepEq(buddies.Result, uiBuddies.buddies, true))
        if (!youtil.deepEq(buddies.Result, uiBuddies.buddies, false)) // cmp not-ignoring order by design (result always ordered by last-active)
            uiBuddies.update(buddies.Result)
    } catch (err) {
        if (!knownErr<yo.UserBuddiesErr>(err, handleKnownErrMaybe<yo.UserBuddiesErr>))
            onErr(err)
    }

    if (!fetchesPaused)
        setTimeout(fetchBuddies, fetchBuddiesIntervalMs)
}

async function fetchRefresh() {
    if (fetchesPaused)
        return
    try {
        const recent_updates = await yo.apiRecentUpdates({ Since: fetchPostsSinceDt ? fetchPostsSinceDt : none })
        fetchPostsSinceDt = recent_updates.Next

        if (recent_updates.Posts && recent_updates.Posts.length) { }
    } catch (err) {
        if (!knownErr<yo.RecentUpdatesErr>(err, handleKnownErrMaybe<yo.RecentUpdatesErr>))
            onErr(err)
    }
    if (!fetchesPaused)
        setTimeout(fetchRefresh, fetchPostsIntervalMsCur)
}

function newUiLoginDialog() {
    const on_btn_login = async () => {
        try {
            await yo.apiUserSignIn({ EmailAddr: in_user_name.value, PasswordPlain: in_password.value })
            alert("ok!")
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
                onErr(err)
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

function newUiPostsFeed() {
    return htm.ul({},
        htm.li({}, "post 1"),
        htm.li({}, "post 2"),
        htm.li({}, "post 3"),
        htm.li({}, "post 4"),
    )
}
