import van, { State } from '../../__yostatic/vanjs/van-1.2.3.debug.js'
import * as vanx from '../../__yostatic/vanjs/van-x.js'
const htm = van.tags, depends = van.derive

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as haxsh from '../haxsh.js'
import * as uibuddies from './buddies.js'
import * as util from '../util.js'

const freshnessDurationMsWhenVisible = 3456

export type UiCtlPosts = {
    DOM: HTMLElement
    _htmPostInput: HTMLElement
    posts: vanx.Reactive<PostAug[]>
    getPostAuthor: (post?: yo.Post) => yo.User | undefined
    update: (_: yo.Post[]) => yo.Post | undefined
    doSendPost: (html: string, files?: string[]) => Promise<boolean>
    numFreshPosts: number
    isSending: State<boolean>
}

type PostAug = yo.Post & {
    _uxStrAgo: string
    _isFresh: boolean
}

export function create(
    getPostAuthor: (post?: yo.Post) => yo.User | undefined,
    doSendPost: (html: string, files?: string[]) => Promise<boolean>,
): UiCtlPosts {
    let me: UiCtlPosts
    const htm_post = htm.div({
        'class': depends(() => 'post-content' + (haxsh.isSeeminglyOffline.val ? ' offline' : '') + ((me && me.isSending.val) ? ' sending' : '')),
        'contenteditable': depends(() => ((me && me.isSending.val) ? 'false' : 'true')),
        'autofocus': true, 'spellcheck': false, 'autocorrect': 'off', 'tabindex': 1, 'onkeydown': (evt: KeyboardEvent) => {
            if (['Enter', 'NumpadEnter'].includes(evt.code) && !(evt.shiftKey || evt.ctrlKey || evt.altKey || evt.metaKey)) {
                evt.preventDefault()
                evt.stopPropagation()
                return sendPost(me)
            }
        },
    }, "")
    const button_state = () => (haxsh.isSeeminglyOffline.val || (me?.isSending.val) || false /* falsy madness sometimes bites =) */)
    me = {
        _htmPostInput: htm_post,
        doSendPost: doSendPost,
        getPostAuthor: getPostAuthor,
        isSending: van.state(false),
        DOM: htm.div({ 'class': 'haxsh-posts' },
            htm.div({ 'class': 'self-post' },
                htm.div({ 'class': 'post' },
                    htm.div({ 'class': 'post-head' },
                        htm.div(uibuddies.userDomAttrsSelf()),
                        htm.div({ 'class': 'post-ago' }, ""),
                    ),
                    htm.div({ 'class': 'post-buttons' },
                        htm.button({
                            'type': 'button', 'class': 'button send', 'title': "Send", 'tabindex': 2,
                            'disabled': depends(button_state), 'onclick': (() => sendPost(me)),
                        }),
                        htm.button({
                            'type': 'button', 'class': 'button attach', 'title': "Add Files", 'tabindex': 3,
                            'disabled': depends(button_state), 'onclick': () => { },
                        }),
                    ),
                    htm_post,
                ),
            ),
        ),
        numFreshPosts: 0,
        posts: vanx.reactive([] as PostAug[]),
        update: (posts) => update(me, posts),
    }

    van.add(me.DOM, vanx.list(() => htm.div({ 'class': 'feed' }), me.posts, (it) => {
        const post = it.val
        const htm_post = htm.div({ 'class': 'post-content' + (post._isFresh ? ' fresh' : '') })
        htm_post.innerHTML = post.Htm || `(files: ${post.Files.join(", ")})`
        const post_by = me.getPostAuthor(post), post_dt = new Date(post.DtMade!)
        const is_own_post = (post_by?.Id === haxsh.userSelf.val?.Id) || false
        return htm.div({ 'class': 'post' },
            htm.div({ 'class': 'post-head' },
                htm.div(is_own_post ? uibuddies.userDomAttrsSelf() : uibuddies.userDomAttrsBuddy(post_by)),
                htm.div({ 'class': 'post-ago', 'title': post_dt.toLocaleDateString() + " — " + post_dt.toLocaleTimeString() }, post._uxStrAgo),
            ),
            htm.div({ 'class': 'post-buttons' },
                htm.button({
                    'type': 'button', 'class': 'button edit', 'title': "Edit", 'style': `visibility:${is_own_post ? 'visible' : 'hidden'}`,
                    'disabled': depends(button_state), 'onclick': () => { },
                }),
            ),
            htm_post,
        )
    }))
    return me
}

async function sendPost(me: UiCtlPosts) {
    if (haxsh.isSeeminglyOffline.val)
        return false
    me.isSending.val = true
    let post_html = me._htmPostInput.innerHTML.trim()
    while (post_html.startsWith('<br>'))
        post_html = post_html.substring('<br>'.length)
    while (post_html.endsWith('<br>'))
        post_html = post_html.substring(0, post_html.length - '<br>'.length)
    if ((post_html.length === 0) || (post_html.replaceAll('<br>', '').replaceAll('<p></p>', '').length === 0))
        return false

    const ok = await me.doSendPost(post_html)
    me.isSending.val = false
    if (ok) {
        me._htmPostInput.innerHTML = ''
        window.scrollTo(0, 0)
    }
    me._htmPostInput.focus()
    return false
}

function update(me: UiCtlPosts, newOrUpdatedPosts: yo.Post[]) {
    let num_fresh = 0, last_ago_str = ""
    const now = new Date().getTime()
    const all_new_posts: yo.Post[] = newOrUpdatedPosts.filter(post_upd =>
        !me.posts.some(post_old => (post_old.Id === post_upd.Id)))
    const merged_with_others = all_new_posts.concat(me.posts.map(post_old => {
        post_old._isFresh = false
        return (newOrUpdatedPosts.find(_ => (_.Id === post_old.Id))) ?? post_old
    }))
    const fresh_feed = merged_with_others
        .map((post: yo.Post): PostAug => {
            const post_time = new Date(post.DtMod ?? (post.DtMade!)).getTime()
            let post_ago_str = util.timeAgoStr(post_time, now, true, "")
            if (post_ago_str === last_ago_str)
                post_ago_str = ""
            const is_fresh = (me.posts.length > 0) && ((haxsh.browserTabInvisibleSince === 0)
                ? ((now - post_time) < freshnessDurationMsWhenVisible)
                : (post_time >= haxsh.browserTabInvisibleSince))
            if (is_fresh)
                num_fresh++
            const ret = { ...post, _uxStrAgo: post_ago_str, _isFresh: is_fresh }
            if (post_ago_str !== "")
                last_ago_str = post_ago_str
            return ret
        })
    me.numFreshPosts = num_fresh
    if (!youtil.deepEq(me.posts, fresh_feed, true, false))
        vanx.replace(me.posts, (_: PostAug[]) => fresh_feed)
    if (fresh_feed.length > 0)
        return fresh_feed[0]
    return undefined
}
