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
            htm.div({ 'class': 'header' },
                user.Nick, htm.div({ 'class': 'pic', 'style': `background-image:url('${uibuddies.userPicFileUrl(user)}')` })),
            htm.div({ 'class': 'btw' }, user.Btw),
        )
    }
    return me
}
