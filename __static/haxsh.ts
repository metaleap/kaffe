import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
import vanx from '../__yostatic/vanjs/van-x.js'
import * as yo from './yo-sdk.js'

import * as uibuddies from './ui/buddies.js'

const htm = van.tags

let fetchRefreshSince: string | undefined
let fetchRefreshIntervalMsWhenVisible = 2345
let fetchRefreshIntervalMsWhenHidden = 4321
let fetchRefreshIntervalMsWhenCur = fetchRefreshIntervalMsWhenVisible
let fetchPaused = false // true while signed out

let dialog_login = uiLoginDialog()
let feed_posts = uiPostsFeed()

function onErr(err: any) { console.error(err) }
function knownErr<T extends string>(err: any, ifSo: (_: T) => void): boolean {
    const yo_err = err as yo.Err<T>
    return (yo_err && yo_err.knownErr && (yo_err.knownErr.length > 0))
}

export function main() {
    document.onvisibilitychange = () => {
        fetchRefreshIntervalMsWhenCur = ((document.visibilityState == 'hidden') || (document.hidden))
            ? fetchRefreshIntervalMsWhenHidden : fetchRefreshIntervalMsWhenVisible
        document.title = fetchRefreshIntervalMsWhenCur.toString()
    }
    van.add(document.body,
        feed_posts,
        uibuddies.create().DOM,
        dialog_login,
    )
    setTimeout(fetchRefresh, 321)
}

async function fetchRefresh() {
    if (fetchPaused)
        return
    try {
        const recent_updates = await yo.apiRecentUpdates({ Since: fetchRefreshSince })
        fetchRefreshSince = recent_updates.Next
        if (recent_updates.Buddies || (recent_updates.Posts && recent_updates.Posts.length > 0))
            console.log(fetchRefreshSince, recent_updates.Buddies, recent_updates.Posts.length)
        // for (const post of recent_updates.Posts)
        //     console.log(fetchRefreshSince, post)
    } catch (err) {
        if (!knownErr<yo.RecentUpdatesErr>(err, (err) => {
            switch (err) {
                case 'Unauthorized':
                    fetchPaused = true
                    dialog_login.showModal()
            }
        }))
            onErr(err)
    }
    if (!fetchPaused)
        setTimeout(fetchRefresh, fetchRefreshIntervalMsWhenCur)
}

function uiLoginDialog() {
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
                        alert(err);
                        return;
                }
            }))
                onErr(err)
        }
    }

    const in_user_name = htm.input({ 'value': 'foo4874@bar.baz' })
    const in_password = htm.input({ 'type': 'password', 'value': 'foobar' })
    return htm.dialog({},
        htm.form({},
            in_user_name,
            in_password,
            htm.button({ 'onclick': on_btn_login, 'type': 'button' }, "Login"),
        ),
    )
}

function uiPostsFeed() {
    return htm.ul({},
        htm.li({}, "post 1"),
        htm.li({}, "post 2"),
        htm.li({}, "post 3"),
        htm.li({}, "post 4"),
    )
}
