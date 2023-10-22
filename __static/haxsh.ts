import van from '../__yostatic/vanjs/van-1.2.3.debug.js'

const htm = van.tags

let loginDialog = () => {
    const in_user_name = htm.input({})
    const in_password = htm.input({ 'type': 'password' })
    return htm.dialog({},
        in_user_name,
        in_password,
        htm.button({
            'onclick': () => {
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
