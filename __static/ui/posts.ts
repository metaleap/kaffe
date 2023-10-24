import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as uibuddies from './buddies.js'
import * as util from '../util.js'

export type UiCtlPosts = {
    DOM: HTMLElement
    getUser: (id: number) => yo.User | undefined
    posts: vanx.Reactive<PostAug[]>
    update: (_: yo.Post[]) => void
}

type PostAug = yo.Post & { _uxStrAgo: string }

export function create(getUser: (id: number) => yo.User | undefined): UiCtlPosts {
    const me: UiCtlPosts = {
        DOM: htm.div({ 'class': 'haxsh-posts' }),
        getUser: getUser,
        posts: vanx.reactive([] as PostAug[]),
        update: (posts) => update(me, posts),
    }

    van.add(me.DOM, vanx.list(htm.div, me.posts, (it) => {
        const post = it.val, now = new Date().getTime()
        const post_by = me.getUser(post.By)
        const dt = new Date(post.DtMade!)
        return htm.div({ 'class': 'post' },
            htm.div({ 'class': 'post-head' },
                htm.div(uibuddies.buddyDomAttrs(post_by, now)),
                htm.div({ 'class': 'post-ago', 'title': dt.toLocaleDateString() + " @ " + dt.toLocaleTimeString() }, post._uxStrAgo),
            ),
            htm.div({ 'class': 'post-content' }, post.Md || `(files: ${post.Files.join(", ")})`),
        )
    }))
    return me
}

function update(me: UiCtlPosts, newOrUpdatedPosts: yo.Post[]) {
    const now = new Date().getTime()
    let prev: PostAug | undefined, last_with_ago: PostAug | undefined
    const fresh_feed = newOrUpdatedPosts
        .filter(post_upd => !me.posts.some(post_old => (post_old.Id === post_upd.Id)))
        .concat(me.posts.map(post_old => newOrUpdatedPosts.find(_ => (_.Id === post_old.Id)) ?? post_old))
        .map((post: yo.Post): PostAug => {
            let ago_str = util.timeAgoStr(new Date(post.DtMade!).getTime(), now, true, "")
            if (last_with_ago && (ago_str === last_with_ago._uxStrAgo))
                ago_str = ""
            prev = {
                ...post,
                _uxStrAgo: ago_str
            }
            if (ago_str)
                last_with_ago = prev
            return prev
        })
    if (!youtil.deepEq(me.posts, fresh_feed, true, false))
        vanx.replace(me.posts, (_: PostAug[]) => fresh_feed)
}
