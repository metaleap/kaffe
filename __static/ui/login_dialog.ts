import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as kaffe from '../kaffe.js'


export function create(setSignUpOrPwdForgotNotice: (_: string) => void) {
    const is_from_signup_or_pwd_reset_mail = location.search.includes('needPwd')
    const on_btn_clicked = async () => {
        if (!(in_user_name.value = in_user_name.value.trim()))
            return
        in_user_name.disabled = true
        in_password.disabled = true
        in_password_2.disabled = true
        in_password.value = in_password.value.trim()
        in_password_2.value = in_password_2.value.trim()
        try {
            const is_signup_or_pwd_forgotten = (!in_password.value.length)
            if (is_signup_or_pwd_forgotten)
                await yo.apiUserSignUpOrForgotPassword({ NickOrEmailAddr: in_user_name.value })
            else
                await yo.apiUserSignInOrReset({ NickOrEmailAddr: in_user_name.value, PasswordPlain: in_password.value, Password2Plain: in_password_2.value })
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
            in_password_2.disabled = false
            if (!kaffe.knownErr<yo.UserSignInOrResetErr | yo.UserSignUpOrForgotPasswordErr>(err, (err) => {
                switch (err) {
                    case '___yo_authLoginOrFinalizePwdReset_AccountDoesNotExist':
                    case '___yo_authLoginOrFinalizePwdReset_WrongPassword':
                    case '___yo_authLoginOrFinalizePwdReset_EmailInvalid':
                    case '___yo_authLoginOrFinalizePwdReset_EmailRequiredButMissing':
                    case 'UserSignInOrReset_ExpectedPasswordAndNickOrEmailAddr':
                    case 'UserSignInOrReset_WrongPassword':
                    case '___yo_authRegister_EmailInvalid':
                    case '___yo_authRegister_EmailRequiredButMissing':
                    case 'UserSignUpOrForgotPassword_EmailInvalid':
                    case 'UserSignUpOrForgotPassword_EmailRequiredButMissing':
                        alert("There's surely a typo in there, please double-check and try again.")
                        return true
                }
                return false
            }))
                kaffe.onErrOther(err, true)
        }
    }

    const in_user_name = htm.input({ 'placeholder': "(your nick or email address)" })
    const in_password = htm.input({
        'type': 'password', 'placeholder':
            is_from_signup_or_pwd_reset_mail ? "(paste the auto-generated one-time code from your confirmation email)"
                : "(your account password: keep blank to sign up — OR if forgotten)"
    })
    const in_password_2 = htm.input({
        'type': 'password', 'placeholder':
            is_from_signup_or_pwd_reset_mail ? `(choose your preferred new sign-in password, min. ${yo.Cfg_YO_AUTH_PWD_MIN_LEN} characters)`
                : `(ONLY to change password: new one here, old one above, min. ${yo.Cfg_YO_AUTH_PWD_MIN_LEN} characters)`
    })
    const dialog = htm.dialog({ 'class': 'login-popup' }, htm.form({ 'onsubmit': () => false },
        htm.button({ 'type': 'submit', 'class': 'save', 'title': "Sign in or sign up now", 'onclick': _ => on_btn_clicked() }, "✅"),
        in_user_name,
        in_password,
        in_password_2,
    ))
    dialog.onclose = (evt) => {
        if (!kaffe.signUpOrPwdForgotNotice.val)
            setTimeout(() => {
                if (!kaffe.signUpOrPwdForgotNotice.val)
                    dialog.showModal()
            }, 1234)
    }
    return dialog
}
