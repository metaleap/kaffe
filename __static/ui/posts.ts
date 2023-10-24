import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as uibuddies from './buddies.js'

export type UiCtlPosts = {
    DOM: HTMLElement
    getUser: (id: number) => yo.User | undefined
    posts: vanx.Reactive<yo.Post[]>
    update: (_: yo.Post[]) => void
}

export function create(getUser: (id: number) => yo.User | undefined): UiCtlPosts {
    const me: UiCtlPosts = {
        DOM: htm.div({ 'class': 'haxsh-posts' }),
        getUser: getUser,
        posts: vanx.reactive([] as yo.Post[]),
        update: (posts) => update(me, posts),
    }

    van.add(me.DOM, vanx.list(htm.div, me.posts, (it) => {
        const post = it.val, now = new Date().getTime()
        const post_by = me.getUser(post.By)! // TODO
        return htm.div({ 'class': 'post' },
            htm.div(uibuddies.buddyDomAttrs(post_by, now)),
            htm.div({ 'class': 'post-content' }, post.DtMade, post.Md),
        )
    }))
    return me
}

function update(me: UiCtlPosts, newOrUpdatedPosts: yo.Post[]) {
    vanx.replace(me.posts, (oldPosts: yo.Post[]) => {
        const ret = newOrUpdatedPosts
            .filter(post_upd => !oldPosts.some(post_old => (post_old.Id === post_upd.Id)))
            .concat(oldPosts.map(post_old => newOrUpdatedPosts.find(_ => (_.Id === post_old.Id)) ?? post_old))
        return ret
    })
}
