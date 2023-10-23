import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import vanx from '../../__yostatic/vanjs/van-x.js'
import * as yo from '../yo-sdk.js'

const htm = van.tags

export type UiBuddies = {
    _htmlDomNode: HTMLElement
    buddies: vanx.Reactive<yo.User[]>

}

export function create(): UiBuddies {
    const me: UiBuddies = {
        _htmlDomNode: htm.div({ 'class': 'haxsh-buddies' }),
        buddies: vanx.reactive([
            { Auth: 1234, Id: 1234, Btw: "Btw 1234", Nick: "user1234", PicFileId: "" } as yo.User,
            { Auth: 2345, Id: 2345, Btw: "Btw 2345", Nick: "user2345", PicFileId: "" } as yo.User,
            { Auth: 4321, Id: 4321, Btw: "Btw 4321", Nick: "user4321", PicFileId: "" } as yo.User,
            { Auth: 123, Id: 123, Btw: "Btw 123", Nick: "user123", PicFileId: "" } as yo.User,
            { Auth: 321, Id: 321, Btw: "Btw 321", Nick: "user321", PicFileId: "" } as yo.User,
        ])
    }

    van.add(me._htmlDomNode, vanx.list(htm.ul, me.buddies, (it) => {

        return htm.li({},)
    }))

    return me
}
