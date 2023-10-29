import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as haxsh from '../haxsh.js'

import * as uibuddies from './buddies.js'

export type UiCtlUserPopup = {
    DOM: HTMLDialogElement
}

export const darkliteDefault = 88

export function darkliteCurrent(): number {
    return parseInt(localStorage.getItem('darklite') ?? darkliteDefault.toString())
}

export function setLiveDarklite(value?: string | number) {
    if (value === undefined)
        value = darkliteCurrent()
    document.getElementById('theme')!.innerHTML = `:root {--liteness: ${value}%;}`
}

export function create(user: yo.User): UiCtlUserPopup {
    const on_darklite_slider_change = () =>
        setLiveDarklite(htm_input_darklite.value)
    const is_self = (haxsh.userSelf.val) && (haxsh.userSelf.val.Id === user.Id),
        htm_input_nick = htm.input({ 'type': 'text', 'class': 'nick', 'value': user.Nick, 'placeholder': '(Nickname)', 'spellcheck': false, 'autocorrect': 'off' }),
        htm_input_btw = htm.input({ 'type': 'text', 'class': 'btw', 'value': user.Btw, 'placeholder': '(Your hover statement here)' }),
        htm_input_pic = htm.input({ 'type': 'file', 'name': 'picfile', 'id': 'picfile' }),
        htm_input_darklite = htm.input({ 'type': 'range', 'id': 'darklite', 'class': 'darklite', 'value': darkliteCurrent(), 'min': 0, 'max': 100, 'step': 1, 'onchange': on_darklite_slider_change })
    const save_changes = () => {
        const darklite = parseInt(htm_input_darklite.value)
        if (!isNaN(darklite))
            localStorage.setItem('darklite', htm_input_darklite.value)

        htm_input_nick.value = htm_input_nick.value.trim()
        htm_input_btw.value = htm_input_btw.value.trim()
        const pic_has_changed = (htm_input_pic.files?.length) && (htm_input_pic.files[0])
        const has_changed = pic_has_changed || (htm_input_nick.value !== user.Nick) || (htm_input_btw.value !== user.Btw)
        if (has_changed) {
            const form_data = new FormData()
            if (pic_has_changed)
                form_data.append('picfile', htm_input_pic.files![0])
            yo.apiUserUpdate({
                Id: user.Id, Changes: {
                    ...user,
                    Nick: htm_input_nick.value,
                    Btw: htm_input_btw.value,
                }, ChangedFields: ['Btw', 'Nick']
            }, pic_has_changed ? form_data : undefined)
            // yo.apiUserUpdate()
        }

        me.DOM.close()
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
    me.DOM.onclose = () => {
        console.log("ONCLOSE")
        if (is_self)
            setLiveDarklite()
        me.DOM.remove()
    }
    return me
}
