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
    const me: UiCtlUserPopup = {
        DOM: htm.dialog({ 'class': 'user-popup' },
            htm.div({},
                htm.input({ 'type': 'text', 'class': 'nick', 'value': user.Nick, 'placeholder': '(Nickname)', 'spellcheck': false, 'autocorrect': 'off' }),
                htm.div({ 'class': 'pic', 'style': `background-image:url('${uibuddies.userPicFileUrl(user)}')` })),
            htm.div({},
                htm.input({ 'type': 'text', 'class': 'btw', 'value': user.Btw, 'placeholder': '(Your hover message/slogan/thought here)' }),
            ),
            htm.div({}),
            htm.div({},
                htm.label({ 'for': 'darklite' }, "UI Dark/Light:"),
                htm.input({ 'type': 'range', 'id': 'darklite', 'class': 'darklite', 'value': 88, 'min': 0, 'max': 100, 'step': 1 }),
            ),
        )
    }
    return me
}
