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
    update: (_: yo.Post[]) => yo.Post | undefined
    numFreshPosts: number
    isSending: State<boolean>
    isDeleting: State<number>
}

type PostAug = yo.Post & {
    _uxStrAgo: string
    _isFresh: boolean
    _isDel: boolean
}

export function create(): UiCtlPosts {
    const is_sending = van.state(false), is_deleting = van.state(0)
    let me: UiCtlPosts
    const htm_post = htm.div({
        'class': depends(() => 'post-content' + (haxsh.isSeeminglyOffline.val ? ' offline' : '') + (is_sending.val ? ' sending' : '')),
        'contenteditable': depends(() => (is_sending.val ? 'false' : 'true')),
        'autofocus': true, 'spellcheck': false, 'autocorrect': 'off', 'tabindex': 1, 'onkeydown': (evt: KeyboardEvent) => {
            if (['Enter', 'NumpadEnter'].includes(evt.code) && !(evt.shiftKey || evt.ctrlKey || evt.altKey || evt.metaKey)) {
                evt.preventDefault()
                evt.stopPropagation()
                return postSendNew(me)
            }
        },
    }, "")
    const button_disabled = () => {
        return (haxsh.isSeeminglyOffline.val || (is_deleting.val > 0) || is_sending.val)
    }
    me = {
        _htmPostInput: htm_post,
        isSending: is_sending,
        isDeleting: is_deleting,
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
                            'disabled': depends(button_disabled), 'onclick': (() => postSendNew(me)),
                        }),
                        htm.button({
                            'type': 'button', 'class': 'button attach', 'title': "Add Files", 'tabindex': 3,
                            'disabled': depends(button_disabled), 'onclick': () => { },
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
        const htm_post = htm.div({ 'class': depends(() => ('post-content' + ((me.isDeleting.val === (post.Id!)) ? ' deleting' : (post._isFresh ? ' fresh' : '')))) })
        htm_post.innerHTML = post.Htm || `(files: ${post.Files.join(", ")})`
        const post_by = haxsh.getUserByPost(post), post_dt = new Date(post.DtMade!)
        const is_own_post = (post_by?.Id === haxsh.userSelf.val?.Id) || false
        return htm.div({ 'class': 'post' },
            htm.div({ 'class': 'post-head' },
                htm.div(is_own_post ? uibuddies.userDomAttrsSelf() : uibuddies.userDomAttrsBuddy(post_by)),
                htm.div({ 'class': 'post-ago', 'title': post_dt.toLocaleDateString() + " — " + post_dt.toLocaleTimeString() }, post._uxStrAgo),
            ),
            htm.div({ 'class': 'post-buttons' },
                htm.button({
                    'type': 'button', 'class': 'button delete', 'title': "Delete", 'style': `visibility:${is_own_post ? 'visible' : 'hidden'}`,
                    'disabled': depends(button_disabled), 'onclick': () => postDelete(me, post.Id!),
                }),
            ),
            htm_post,
        )
    }))
    return me
}

async function postDelete(me: UiCtlPosts, postId: number) {
    const post_idx = me.posts.findIndex(_ => (_ && (_.Id === postId)))
    if (post_idx < 0)
        return
    me.isDeleting.val = postId
    const post = me.posts[post_idx]
    post._isDel = true
    await haxsh.deletePost(postId)
    update(me, me.posts)
    me.isDeleting.val = 0
}

async function postSendNew(me: UiCtlPosts) {
    if (haxsh.isSeeminglyOffline.val)
        return false
    let post_html = me._htmPostInput.innerHTML.replaceAll('&nbsp;', ' ').trim()
    while (post_html.startsWith('<br>'))
        post_html = post_html.substring('<br>'.length)
    while (post_html.endsWith('<br>'))
        post_html = post_html.substring(0, post_html.length - '<br>'.length)
    post_html = me._htmPostInput.innerHTML.replaceAll('&nbsp;', ' ').trim() //
    if ((post_html.length === 0) || (post_html.replaceAll('<br>', '').replaceAll('<p></p>', '').trim().length === 0))
        return false

    me.isSending.val = true
    const ok = await haxsh.sendNewPost(post_html)
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
    const new_posts_merged_with_old = all_new_posts.concat(me.posts.filter(_ => (!_._isDel)).map(post_old => {
        post_old._isFresh = false
        return (newOrUpdatedPosts.find(_ => (_.Id === post_old.Id))) ?? post_old
    }))
    const fresh_feed = new_posts_merged_with_old
        .map((post: yo.Post): PostAug => {
            const post_time = new Date(post.DtMod ?? (post.DtMade!)).getTime()
            let post_ago_str = util.timeAgoStr(post_time, now, true, "")
            if (post_ago_str === last_ago_str)
                post_ago_str = ""
            const user_cur = haxsh.userSelf
            const is_fresh = (me.posts.length > 0) && ((haxsh.browserTabInvisibleSince === 0)
                ? ((now - post_time) < freshnessDurationMsWhenVisible)
                : (post_time >= haxsh.browserTabInvisibleSince)) && ((!user_cur.val) || (post.By!) != (user_cur.val.Id))
            if (is_fresh)
                num_fresh++
            const ret: PostAug = { ...post, _uxStrAgo: post_ago_str, _isDel: false, _isFresh: is_fresh }
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
