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
    numFreshPosts: number
    isSending: State<boolean>
    isDeleting: State<number>
    upFilesNative: (File | null)[]
}

type PostAug = yo.Post & {
    _uxStrAgo: string
    _isFresh: boolean
    _isDel: boolean
}

export function create(): UiCtlPosts {
    const is_sending = van.state(false), is_deleting = van.state(0)
    let me: UiCtlPosts
    const files_to_post: vanx.Reactive<UpFile[]> = vanx.reactive([] as UpFile[])
    const htm_post_entry = htm.div({
        'class': depends(() => 'post-content' + (haxsh.isSeeminglyOffline.val ? ' offline' : '') + (is_sending.val ? ' sending' : '')),
        'contenteditable': depends(() => (is_sending.val ? 'false' : 'true')),
        'autofocus': true, 'spellcheck': false, 'autocorrect': 'off', 'tabindex': 1, 'onkeydown': (evt: KeyboardEvent) => {
            if (['Enter', 'NumpadEnter'].includes(evt.code) && !(evt.shiftKey || evt.ctrlKey || evt.altKey || evt.metaKey)) {
                evt.preventDefault()
                evt.stopPropagation()
                return sendNew(me, files_to_post)
            }
        },
    }, "")
    const button_disabled = () => (haxsh.isSeeminglyOffline.val || (is_deleting.val > 0) || is_sending.val)
    const htm_input_file = htm.input({ 'type': 'file', 'multiple': true, 'onchange': () => onFilesAdded(me, files_to_post, htm_input_file) })
    me = {
        _htmPostInput: htm_post_entry,
        isSending: is_sending,
        isDeleting: is_deleting,
        upFilesNative: [],
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
                            'disabled': depends(button_disabled), 'onclick': (() => sendNew(me, files_to_post)),
                        }),
                        htm.button({
                            'type': 'button', 'class': 'button attach', 'title': "Add Files", 'tabindex': 3,
                            'disabled': depends(button_disabled), 'onclick': () => htm_input_file.click(),
                        }),
                    ),
                    htm_input_file,
                    htm.div({},
                        htm_post_entry,
                        vanx.list(() => { return htm.div({ 'class': 'haxsh-post-files' }) }, files_to_post, (_) => {
                            const icon = _.val.type.includes('/') ? (fileContentTypeIcons[_.val.type.substring(0, _.val.type.indexOf('/'))]) : ""
                            // van.add(htm_file, (icon !== fileContentTypeIcons['image']) ? htm.div({}, icon)
                            //     : htm.div({ 'class': 'image', 'style': `background-image:url('${file_url}')` }))
                            return htm.a({ 'class': 'haxsh-post-file', 'title': `${_.val.name + '\n'}${(_.val.size / (1024 * 1024)).toFixed(3)}MB` },
                                (icon ? htm.div({}, icon) : undefined),
                                htm.span({}, _.val.name,
                                    htm.span({}, _.val.type || '(unknown type)',
                                        htm.button({
                                            'class': 'button delete', 'type': 'button', 'title': 'Remove', 'onclick': () => removeUpFile(me, files_to_post, _.val)
                                        }))))
                        }),
                    ),
                ),
            ),
        ),
        numFreshPosts: 0,
        posts: vanx.reactive([] as PostAug[]),
    }

    van.add(me.DOM, vanx.list(() => htm.div({ 'class': 'feed' }), me.posts, (it) => {
        const post = it.val
        let inner_html = post.Htm ?? ''
        if (post.Files && post.Files.length) {
            const htm_files = htm.div({ 'class': 'haxsh-post-files' },
                ...post.Files.map((file_name_full, idx) => {
                    const idx_sep = file_name_full.indexOf("__yo__")
                    const file_name_show = (idx_sep < 0) ? file_name_full : file_name_full.substring(idx_sep + "__yo__".length)
                    const file_content_type = post.FileContentTypes![idx],
                        file_url = `/_postfiles/${encodeURIComponent(file_name_full)}`
                    const htm_file = htm.a({ 'class': 'haxsh-post-file', 'target': '_blank', 'href': file_url })
                    if (file_content_type !== "") {
                        if (file_content_type.startsWith('image/'))
                            htm_file.setAttribute('onclick', `mvd.innerHTML="<img alt='${file_name_show}' title='${file_name_show}' src='${file_url}'>";mvd.showModal();return false`)
                        else if (file_content_type.startsWith('video/'))
                            htm_file.setAttribute('onclick', `mvd.innerHTML="<video controls='true' loop='true' playsinline='true' title='${file_name_show}' src='${file_url}'>";mvd.showModal();return false`)
                        const icon = fileContentTypeIcons[file_content_type.substring(0, file_content_type.indexOf('/'))]
                        van.add(htm_file, (icon !== fileContentTypeIcons['image']) ? htm.div({}, icon)
                            : htm.div({ 'class': 'image', 'style': `background-image:url('${file_url}')` }))
                    }
                    van.add(htm_file, htm.span({}, file_name_show, (file_content_type === "") ? undefined : htm.span({}, file_content_type)))
                    return htm_file
                })
            )
            inner_html += htm_files.outerHTML
        }

        const post_by = haxsh.getUserByPost(post), post_dt = new Date(post.DtMade!)
        const htm_post = htm.div({ 'class': depends(() => ('post-content' + ((me.isDeleting.val === (post.Id!)) ? ' deleting' : (post._isFresh ? ' fresh' : '')))) })
        htm_post.innerHTML = inner_html
        const is_own_post = (post_by?.Id === haxsh.userSelf.val?.Id) || false
        return htm.div({ 'class': 'post' },
            htm.div({ 'class': 'post-head' },
                htm.div(is_own_post ? uibuddies.userDomAttrsSelf() : uibuddies.userDomAttrsBuddy(post_by)),
                htm.div({ 'class': 'post-ago', 'title': post_dt.toLocaleDateString() + " — " + post_dt.toLocaleTimeString() }, post._uxStrAgo),
            ),
            htm.div({ 'class': 'post-buttons' },
                htm.button({
                    'type': 'button', 'class': 'button delete', 'title': "Delete", 'style': `visibility:${is_own_post ? 'visible' : 'hidden'}`,
                    'disabled': depends(button_disabled), 'onclick': () => deletePost(me, post.Id!),
                }),
            ),
            htm_post,
        )
    }))
    return me
}

function hasUpFiles(me: UiCtlPosts) {
    return me.upFilesNative.some(_ => _ !== null)
}

type UpFile = { name: string, type: string, idx: number, size: number, lastModified: number }

function onFilesAdded(me: UiCtlPosts, upFilesOwn: vanx.Reactive<UpFile[]>, htmInputFile: HTMLInputElement) {
    vanx.replace(upFilesOwn, (prevFiles: UpFile[]) => {
        const ret = prevFiles.filter(_ => true)
        const likely_dupls: string[] = []
        for (let i = 0; i < htmInputFile.files!.length; i++) {
            const file = htmInputFile.files!.item(i)
            if (file) {
                if (ret.some(_ => (_.name === file.name) && (_.type === file.type) && (youtil.fEq(_.size, file.size)) && (youtil.fEq(_.lastModified, file.lastModified))))
                    likely_dupls.push(file.name)
                ret.push({ idx: me.upFilesNative.length, name: file.name, type: file.type, size: file.size, lastModified: file.lastModified })
                me.upFilesNative.push(file)
            }
        }
        if (likely_dupls.length)
            alert("Detected probable (but not byte-by-byte-compared) duplicate files, please double-check and remove any duplicates:\n\n· " + likely_dupls.join('\n· '))
        return ret
    })
}

function removeUpFile(me: UiCtlPosts, upFilesOwn: vanx.Reactive<UpFile[]>, upFile: UpFile) {
    me.upFilesNative[upFile.idx] = null
    vanx.replace(upFilesOwn, (prevFiles: UpFile[]) => {
        return prevFiles.filter(_ => (_.idx !== upFile.idx))
    })
}

async function deletePost(me: UiCtlPosts, postId: number) {
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

async function sendNew(me: UiCtlPosts, upFilesOwn: vanx.Reactive<UpFile[]>) {
    const post_html = htmlToSend(me)
    if ((!(post_html && post_html.length)) && !hasUpFiles(me))
        return false

    me.isSending.val = true
    const ok = await haxsh.sendNewPost(post_html, me.upFilesNative.filter(_ => (_ ? true : false)).map(_ => (_ as File)))
    me.isSending.val = false
    if (ok) {
        me._htmPostInput.innerHTML = ''
        me.upFilesNative = []
        vanx.replace(upFilesOwn, () => [])
        window.scrollTo(0, 0)
    }
    me._htmPostInput.focus()
    return false
}

function htmlToSend(me: UiCtlPosts) {
    if (haxsh.isSeeminglyOffline.val)
        return ""
    let post_html = me._htmPostInput.innerHTML
    {   // firefox-only (seemingly) quirks:
        post_html = post_html.replaceAll('&nbsp;', ' ').trim()
        while (post_html.startsWith('<br>'))
            post_html = post_html.substring('<br>'.length).trim()
        while (post_html.endsWith('<br>'))
            post_html = post_html.substring(0, post_html.length - '<br>'.length).trim()
    }
    return ((post_html.length === 0) || (post_html.replaceAll('<br>', '').replaceAll('<p></p>', '').trim().length === 0))
        ? "" : post_html
}

export function update(me: UiCtlPosts, newOrUpdatedPosts: yo.Post[], clearOld?: boolean, sansIds: number[] = []) {
    let num_fresh = 0, last_ago_str = ""
    const now = new Date().getTime(), old_posts = me.posts.filter(_ => true)
    const all_new_posts: yo.Post[] = newOrUpdatedPosts.filter(post_upd =>
        !old_posts.some(post_old => (post_old.Id === post_upd.Id)))
    const new_posts_merged_with_old = clearOld ? all_new_posts : all_new_posts.concat(old_posts
        .filter(_ => (!_._isDel) && ((!sansIds.length) || !sansIds.includes(_.Id!)))
        .map(post_old => {
            post_old._isFresh = false
            return (newOrUpdatedPosts.find(_ => (_.Id === post_old.Id))) ?? post_old
        })
    )
    const fresh_feed = new_posts_merged_with_old
        .map((post: yo.Post): PostAug => {
            const post_time = new Date(post.DtMod ?? (post.DtMade!)).getTime()
            let post_ago_str = util.timeAgoStr(post_time, now, true, "")
            if (post_ago_str === last_ago_str)
                post_ago_str = ""
            const user_cur = haxsh.userSelf
            const is_fresh = (old_posts.length > 0) && ((haxsh.browserTabInvisibleSince === 0)
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
    if (!youtil.deepEq(old_posts, fresh_feed, true, false))
        vanx.replace(me.posts, (_: PostAug[]) => fresh_feed)
    if (fresh_feed.length > 0)
        return fresh_feed[0]
    return undefined
}

const fileContentTypeIcons: { [_: string]: string } = {
    'audio': "🎧",
    'video': "🎥",
    'image': "🖼️",
    'text': "📝",
    'application': "⚙️", // 📦 ⚙️ 🧰 🛠️ 🔧
}
