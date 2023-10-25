import van from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as haxsh from '../haxsh.js'
import * as uibuddies from './buddies.js'
import * as util from '../util.js'

export type UiCtlPosts = {
    DOM: HTMLElement
    _htmPostInput: HTMLElement
    posts: vanx.Reactive<PostAug[]>
    getPostAuthor: (post?: yo.Post) => yo.User | undefined
    update: (_: yo.Post[]) => yo.Post | undefined
    doSendPost: (html: string, files?: string[]) => Promise<boolean>
}

type PostAug = yo.Post & { _uxStrAgo: string }

export function create(
    getPostAuthor: (post?: yo.Post) => yo.User | undefined,
    doSendPost: (html: string, files?: string[]) => Promise<boolean>,
): UiCtlPosts {
    const htm_post = htm.div({
        'class': 'post-content', 'contenteditable': 'true', 'autofocus': true, 'spellcheck': false,
        'autocorrect': 'off', 'tabindex': 1, 'onkeydown': (evt: KeyboardEvent) => {
            if (['Enter', 'NumpadEnter'].includes(evt.code))
                sendPost(me)
        },
    }, "")
    const me: UiCtlPosts = {
        _htmPostInput: htm_post,
        doSendPost: doSendPost,
        getPostAuthor: getPostAuthor,
        DOM: htm.div({ 'class': 'haxsh-posts' },
            htm.div({ 'class': 'self-post' },
                htm.div({ 'class': 'post' },
                    htm.div({ 'class': 'post-head' },
                        htm.div(uibuddies.userDomAttrsSelf()),
                        htm.div({ 'class': 'post-ago' }, ""),
                    ),
                    htm.div({ 'class': 'post-buttons' },
                        htm.button({ 'type': 'button', 'class': 'button send', 'title': "Send", 'tabindex': 2, 'onclick': () => sendPost(me) },
                            "📨"),
                        htm.button({ 'type': 'button', 'class': 'button attach', 'title': "Add Files", 'tabindex': 3, 'onclick': () => { } },
                            "📎"),
                    ),
                    htm_post,
                ),
            ),
        ),
        posts: vanx.reactive([] as PostAug[]),
        update: (posts) => update(me, posts),
    }

    van.add(me.DOM, vanx.list(() => htm.div({ 'class': 'feed' }), me.posts, (it) => {
        const post = it.val, now = new Date().getTime()
        const htm_post = htm.div({ 'class': 'post-content' })
        htm_post.innerHTML = post.Htm || `(files: ${post.Files.join(", ")})`
        const post_by = me.getPostAuthor(post), post_dt = new Date(post.DtMade!)
        const is_own_post = (post_by?.Id === haxsh.userSelf.val?.Id) || false
        return htm.div({ 'class': 'post' },
            htm.div({ 'class': 'post-head' },
                htm.div(is_own_post ? uibuddies.userDomAttrsSelf() : uibuddies.userDomAttrsBuddy(post_by, now)),
                htm.div({ 'class': 'post-ago', 'title': post_dt.toLocaleDateString() + " @ " + post_dt.toLocaleTimeString() }, post._uxStrAgo),
            ),
            htm.div({ 'class': 'post-buttons' },
                htm.button({ 'type': 'button', 'class': 'button', 'title': "Actions" }, "☰"),
            ),
            htm_post,
        )
    }))
    return me
}

async function sendPost(me: UiCtlPosts) {
    const post_html = me._htmPostInput.innerHTML.trim()
    if (post_html.length === 0)
        return
    me._htmPostInput.contentEditable = 'false'
    me._htmPostInput.classList.add('sending')
    const ok = await me.doSendPost(post_html)
    me._htmPostInput.classList.remove('sending')
    me._htmPostInput.contentEditable = 'true'
    if (ok)
        me._htmPostInput.innerHTML = ''
    me._htmPostInput.focus()
}

function update(me: UiCtlPosts, newOrUpdatedPosts: yo.Post[]) {
    const now = new Date().getTime()
    let last_ago_str = ""
    const fresh_feed = newOrUpdatedPosts
        .filter(post_upd => !me.posts.some(post_old => (post_old.Id === post_upd.Id)))
        .concat(me.posts.map(post_old => newOrUpdatedPosts.find(_ => (_.Id === post_old.Id)) ?? post_old))
        .map((post: yo.Post): PostAug => {
            let post_ago_str = util.timeAgoStr(new Date(post.DtMade!).getTime(), now, true, "")
            if (post_ago_str === last_ago_str)
                post_ago_str = ""
            const ret = { ...post, _uxStrAgo: post_ago_str }
            if (post_ago_str !== "")
                last_ago_str = post_ago_str
            return ret
        })
    if (!youtil.deepEq(me.posts, fresh_feed, true, true)) {
        vanx.replace(me.posts, (_: PostAug[]) => fresh_feed)
    }
    if (fresh_feed.length > 0)
        return fresh_feed[0]
    return undefined
}
