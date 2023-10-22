import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
import * as yo from './yo-sdk.js'

const htm = van.tags

function onErr(err: any) { console.error(err) }

export function main() {
    const login_dialog = loginDialog()
    van.add(document.body,
        login_dialog,
    )
}

let loginDialog = () => {
    const on_btn_login = async () => {
        try {
            const prom = yo.apiUserSignIn({ EmailAddr: in_user_name.value, PasswordPlain: in_password.value })
            console.log("prom")
            await prom
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

    const in_user_name = htm.input({ 'value': 'foo1@bar.baz' })
    const in_password = htm.input({ 'type': 'password', 'value': 'foobar' })
    return htm.dialog({ 'open': true },
        htm.form({},
            in_user_name,
            in_password,
            htm.button({ 'onclick': on_btn_login, 'type': 'button' }, "Login"),
        ),
    )
}
