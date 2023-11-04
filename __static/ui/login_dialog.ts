import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as haxsh from '../haxsh.js'


export function create(setSignUpOrPwdForgotNotice: (_: string) => void) {
    const on_btn_clicked = async () => {
        if (!(in_user_name.value = in_user_name.value.trim()))
            return
        in_user_name.disabled = true
        in_password.disabled = true
        in_password_new.disabled = true
        in_password.value = in_password.value.trim()
        in_password_new.value = in_password_new.value.trim()
        try {
            const is_signup_or_pwd_forgotten = (!in_password.value.length)
            if (is_signup_or_pwd_forgotten)
                await yo.apiUserSignUpOrForgotPassword({ NickOrEmailAddr: in_user_name.value })
            else
                await yo.apiUserSignIn({ NickOrEmailAddr: in_user_name.value, PasswordPlain: in_password.value })
            if (is_signup_or_pwd_forgotten) {
                const notice = `An email will be sent to '${in_user_name.value}' within minutes, with the link to complete the sign-up or password-reset.`
                setSignUpOrPwdForgotNotice(notice)
                alert(notice)
                dialog.close()
            } else
                location.reload()
        } catch (err) {
            in_user_name.disabled = false
            in_password.disabled = false
            in_password_new.disabled = false
            if (!haxsh.knownErr<yo.UserSignInErr | yo.UserSignUpOrForgotPasswordErr>(err, (err) => {
                switch (err) {
                    case '___yo_authLogin_AccountDoesNotExist':
                    case '___yo_authLogin_WrongPassword':
                    case '___yo_authLogin_EmailInvalid':
                    case '___yo_authLogin_EmailRequiredButMissing':
                    case 'UserSignIn_ExpectedPasswordAndNickOrEmailAddr':
                    case 'UserSignIn_WrongPassword':
                    case '___yo_authRegister_EmailInvalid':
                    case '___yo_authRegister_EmailRequiredButMissing':
                    case 'UserSignUpOrForgotPassword_EmailInvalid':
                    case 'UserSignUpOrForgotPassword_EmailRequiredButMissing':
                        alert("There's surely a typo in there, please double-check and try again.")
                        return true
                }
                return false
            }))
                haxsh.onErrOther(err, true)
        }
    }

    const in_user_name = htm.input({ 'placeholder': '(nick or email address)' })
    const in_password = htm.input({ 'type': 'password', 'placeholder': '(password: keep blank to sign up — or if forgotten)' })
    const in_password_new = htm.input({ 'type': 'password', 'placeholder': '(only to change password: new one here, old one above)' })
    const dialog = htm.dialog({ 'class': 'login-popup' },
        htm.form({ 'onsubmit': () => false },
            htm.button({ 'type': 'submit', 'class': 'save', 'title': "Sign in or sign up now", 'onclick': _ => on_btn_clicked() }, "✅"),
            in_user_name,
            in_password,
            in_password_new,
        ),
    )
    dialog.onclose = (evt) => {
        if (!haxsh.signUpOrPwdForgotNotice.val)
            setTimeout(() => {
                if (!haxsh.signUpOrPwdForgotNotice.val)
                    dialog.showModal()
            }, 1234)
    }
    return dialog
}
