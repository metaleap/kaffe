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
        htm_input_nick = htm.input({ 'type': 'text', 'class': 'nick', 'value': user.Nick!, 'placeholder': '(Nick)', 'spellcheck': false, 'autocorrect': 'off' }),
        htm_input_btw = htm.input({ 'type': 'text', 'class': 'btw', 'value': user.Btw ?? '', 'placeholder': '(Your hover statement here)', 'spellcheck': false, 'autocorrect': 'off' }),
        htm_input_pic = htm.input({ 'type': 'file', 'name': 'picfile', 'id': 'picfile', 'accept': 'image/*' }),
        htm_div_pic = htm.div({ 'class': 'buddy-pic', 'style': `background-image:url('${uibuddies.userPicFileUrl(user)}');cursor:${is_self ? 'pointer' : 'default'}`, 'onclick': _ => (is_self) ? htm_input_pic.click() : false }),
        htm_input_darklite = htm.input({ 'type': 'range', 'id': 'darklite', 'class': 'darklite', 'value': darkliteCurrent(), 'min': 0, 'max': 100, 'step': 1, 'onchange': on_darklite_slider_change })
    htm_input_pic.onchange = _ => {
        if (!(htm_input_pic.files && htm_input_pic.files.length && htm_input_pic.files[0]))
            return
        if (!(htm_input_pic.files[0].type && htm_input_pic.files[0].type.startsWith('image/'))) {
            alert(`That '${htm_input_pic.files[0].name}' may have merit, but we'll need a picture file here ...mmkay?`)
            htm_input_pic.value = ''
            return
        } else if (htm_input_pic.files[0].size > (1024 * 1024 * yo.reqMaxReqPayloadSizeMb)) {
            alert(`Pretty as heck! But over the ${yo.reqMaxReqPayloadSizeMb}MB limit. What else ye got?`)
            htm_input_pic.value = ''
            return
        }
        const file_reader = new FileReader()
        file_reader.onload = evt =>
            htm_div_pic.style.backgroundImage = `url('${evt.target?.result}')`
        file_reader.readAsDataURL(htm_input_pic.files[0])
    }
    const save_changes = async () => {
        const darklite = parseInt(htm_input_darklite.value)
        if (!isNaN(darklite))
            localStorage.setItem('darklite', htm_input_darklite.value)
        htm_input_nick.value = htm_input_nick.value.trim()
        htm_input_btw.value = htm_input_btw.value.trim()
        const pic_has_changed = (htm_input_pic.files?.length) && (htm_input_pic.files[0])
        const has_changed = pic_has_changed || (htm_input_nick.value !== user.Nick) || (htm_input_btw.value !== user.Btw)
        let did_save = false
        if (has_changed) {
            const form_data = new FormData()
            if (pic_has_changed)
                form_data.append('picfile', htm_input_pic.files![0])
            try {
                await yo.apiUserUpdate({
                    Id: user.Id, Changes: {
                        Nick: htm_input_nick.value,
                        Btw: htm_input_btw.value,
                    }, ChangedFields: ['Btw', 'Nick']
                }, form_data)
                did_save = true
                haxsh.reloadUserSelf()
            } catch (err) {
                if (!haxsh.knownErr<yo.UserUpdateErr>(err, haxsh.handleKnownErrMaybe<yo.UserUpdateErr>))
                    haxsh.onErrOther(err, true)
            }
        }
        if (did_save || !has_changed)
            me.DOM.close()
    }

    const me: UiCtlUserPopup = {
        DOM: htm.dialog({ 'class': 'user-popup' },
            htm.button({ 'type': 'button', 'class': 'close', 'title': "Close", 'onclick': _ => me.DOM.close() }, "❎"),
            (!is_self) ? undefined : htm.button({ 'type': 'button', 'class': 'save', 'title': "Save changes", 'onclick': save_changes }, "✅"),
            htm.div({},
                is_self ? htm_input_nick : htm.span({ 'class': 'nick' }, user.Nick),
                htm_div_pic,
            ),
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
        if (is_self)
            setLiveDarklite()
        me.DOM.remove()
    }
    return me
}
