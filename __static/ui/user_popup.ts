import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags, depends = van.derive

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as haxsh from '../haxsh.js'
import * as util from '../util.js'

import * as uibuddies from './buddies.js'

export type UiCtlUserPopup = {
    DOM: HTMLDialogElement
}

export function create(user: yo.User): UiCtlUserPopup {
    const on_darklite_slider_change = () => {
        document.getElementById('theme')!.innerHTML = `:root {--liteness: ${htm_input_darklite.value}%;}`
    }
    const is_self = (haxsh.userSelf.val) && (haxsh.userSelf.val.Id === user.Id),
        darklite = parseInt(localStorage.getItem("darklight") ?? "88"),
        htm_input_nick = htm.input({ 'type': 'text', 'class': 'nick', 'value': user.Nick, 'placeholder': '(Nickname)', 'spellcheck': false, 'autocorrect': 'off' }),
        htm_input_btw = htm.input({ 'type': 'text', 'class': 'btw', 'value': user.Btw, 'placeholder': '(Your hover message/slogan/thought here)' }),
        htm_input_pic = htm.input({ 'type': 'file' }),
        htm_input_darklite = htm.input({ 'type': 'range', 'id': 'darklite', 'class': 'darklite', 'value': darklite, 'min': 0, 'max': 100, 'step': 1, 'onchange': on_darklite_slider_change })
    const save_changes = () => {

    }
    const me: UiCtlUserPopup = {
        DOM: htm.dialog({ 'class': 'user-popup' },
            htm.button({ 'type': 'button', 'class': 'close', 'onclick': _ => me.DOM.close() }, "🗙"),
            (!is_self) ? undefined : htm.button({ 'type': 'button', 'class': 'save', 'onclick': save_changes }, "✅"),
            htm.div({},
                is_self ? htm_input_nick : htm.span({ 'class': 'nick' }, user.Nick),
                htm.div({ 'class': 'pic', 'style': `background-image:url('${uibuddies.userPicFileUrl(user)}');cursor:${is_self ? 'pointer' : 'default'}`, 'onclick': _ => (is_self) ? htm_input_pic.click() : false })),
            htm.div({},
                is_self ? htm_input_btw : htm.span({ 'class': 'btw' }, user.Btw),
            ),
            htm.div({}),
            (!is_self) ? undefined : htm.div({},
                htm.label({ 'for': 'darklite' }, "UI Dark/Light:"),
                htm_input_darklite,
            ),
        )
    }
    return me
}
