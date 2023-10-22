import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
import * as yo from './yo-sdk.js'

const htm = van.tags

let fetchRefreshSince: string | undefined
let fetchRefreshIntervalMs = 1234 // TODO: higher on-tab-blur

function onErr(err: any) { console.error(err) }

export function main() {
    const login_dialog = loginDialog()
    van.add(document.body,
        login_dialog,
    )
    setTimeout(fetchRefresh, 123)
}

async function fetchRefresh() {
    try {
        const recent_updates = await yo.apiRecentUpdates({ Since: fetchRefreshSince })
        fetchRefreshSince = recent_updates.Next
        if (recent_updates.Buddies || (recent_updates.Posts && recent_updates.Posts.length > 0))
            console.log(fetchRefreshSince, recent_updates.Buddies, recent_updates.Posts.length)
        // for (const post of recent_updates.Posts)
        //     console.log(fetchRefreshSince, post)
    } catch (_) { }
    setTimeout(fetchRefresh, fetchRefreshIntervalMs)
}

function loginDialog() {
    const on_btn_login = async () => {
        try {
            await yo.apiUserSignIn({ EmailAddr: in_user_name.value, PasswordPlain: in_password.value })
            alert("ok!")
        } catch (err) {
            const yo_err = err as yo.Err<yo.UserSignInErr>
            switch (yo_err.knownErr) {
                case '___yo_authLogin_WrongPassword':
                case '___yo_authLogin_AccountDoesNotExist':
                case '___yo_authLogin_EmailInvalid':
                    alert(yo_err.knownErr);
                    return;
            }
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
