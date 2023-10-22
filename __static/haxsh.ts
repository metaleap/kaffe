import van from '../__yostatic/vanjs/van-1.2.3.debug.js'
import * as yo from './yo-sdk.js'

const htm = van.tags

function onErr(err: any) { console.error(err) }

let loginDialog = () => {
    const in_user_name = htm.input({})
    const in_password = htm.input({ 'type': 'password' })
    return htm.dialog({ 'open': true },
        in_user_name,
        in_password,
        htm.button({
            'onclick': async () => {
                try {
                    await yo.apiUserSignIn({ EmailAddr: in_user_name.value, PasswordPlain: in_password.value })
                    alert("ok!")
                } catch (err) {
                    // alert(JSON.stringify(err))
                    // switch ((err as yo.Err<yo.UserSignInErr>).knownErr) {
                    //     case '___yo_authLogin_WrongPassword':
                    //     case '___yo_authLogin_AccountDoesNotExist':
                    //     case '___yo_authLogin_EmailInvalid':
                    //         alert(err);
                    //         return;
                    // }
                    onErr(err)
                }
                return false
            }
        }, "Login")
    )
}

export function main() {
    const login_dialog = loginDialog()
    van.add(document.body,
        login_dialog,
    )
}
