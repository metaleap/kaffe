import van, { State } from '../../__yostatic/vanjs/van-1.2.6.js'
const htm = van.tags, depends = van.derive

import * as yo from '../yo-sdk.js'
import * as youtil from '../../__yostatic/util.js'
import * as youi from '../../__yostatic/youi.js'
import * as kaffe from '../kaffe.js'
import * as uibuddies from './buddies.js'
import * as util from '../util.js'

const freshnessDurationMsWhenVisible = 3456

export type UiCtlPosts = {
    DOM: HTMLElement
    _htmPostInput: HTMLElement
    posts: youi.DomLive<PostAug>
    numFreshPosts: number
    isSending: State<boolean>
    isRequestingDeletion: State<number>
    upFilesNative: (File | null)[]
    upFilesOwn: youi.DomLive<UpFile>
}

type PostAug = yo.Post & {
    _uxStrAgo: string
    _isFresh: boolean
    _isDel: boolean
}

export function create(): UiCtlPosts {
    const is_sending = van.state(false), is_deleting = van.state(0), is_empty = van.state(true)
    let me: UiCtlPosts
    const files_to_post: UpFile[] = []
    const htm_post_entry = htm.div({
        'class': depends(() => 'post-content' + (kaffe.isSeeminglyOffline.val ? ' offline' : '') + (is_sending.val ? ' sending' : '') + (is_empty.val ? ' empty' : '')),
        'contenteditable': depends(() => ((is_sending.val || kaffe.isArchiveBrowsing.val || !kaffe.userSelf.val) ? 'false' : 'true')),
        'autofocus': true, 'spellcheck': false, 'autocorrect': 'off', 'tabindex': 1,
        'title': depends(() =>
            kaffe.isSeeminglyOffline.val ? "You seem to be offline, or our backend is. Or hax0rs, or a meteor strike on The Cloud, or TEOTWAWKI, or a stray cosmic ray... but it's probably your router resetting."
                : ((!kaffe.userSelf.val) ? (kaffe.signUpOrPwdForgotNotice.val || "Sign in or sign up to resume confabulations:") : (kaffe.isArchiveBrowsing.val
                    ? "Browsing archives. To chat, switch back to 'Fresh'."
                    : (kaffe.selectedBuddy.val
                        ? `Chat with ${kaffe.userById(kaffe.selectedBuddy.val)?.Nick || "?"}`
                        : "This goes to all buddies. (For 1-to-1 chat, select a buddy on the right.)")))),
        'oninput': () => {
            is_empty.val = (htm_post_entry.innerHTML === "") || (htm_post_entry.innerHTML === "<br>")
        },
        'onkeydown': (evt: KeyboardEvent) => {
            if (['Enter', 'NumpadEnter'].includes(evt.code) && !(evt.shiftKey || evt.ctrlKey || evt.altKey || evt.metaKey)) {
                evt.preventDefault()
                evt.stopPropagation()
                sendNew(me, me.upFilesOwn.all, is_empty)
                return false
            }
        },
    }, "")
    const button_disabled = () => (kaffe.isSeeminglyOffline.val || (is_deleting.val > 0) || is_sending.val || !kaffe.userSelf.val)
    const htm_input_file = htm.input({ 'type': 'file', 'multiple': true, 'onchange': () => onFilesAdded(me, htm_input_file) })
    const up_files_own = youi.domLive<UpFile>(htm.div({ 'class': 'kaffe-post-files' }), [], (_: UpFile) => {
        const icon = _.type.includes('/') ? (fileContentTypeIcons[_.type.substring(0, _.type.indexOf('/'))]) : ""
        return htm.a({ 'class': 'kaffe-post-file', 'title': `${_.name + '\n'}${(_.size / (1024 * 1024)).toFixed(3)}MB` },
            (icon ? htm.div({}, icon) : undefined),
            htm.span({}, _.name,
                htm.span({}, _.type || '(unknown type)',
                    htm.button({
                        'class': 'button delete', 'type': 'button', 'title': 'Remove', 'disabled': depends(button_disabled), 'onclick': () => removeUpFile(me, _)
                    }))))
    }, 'idx')
    me = {
        _htmPostInput: htm_post_entry,
        isSending: is_sending,
        isRequestingDeletion: is_deleting,
        upFilesNative: [],
        upFilesOwn: up_files_own,
        DOM: htm.div({ 'class': 'kaffe-posts' },
            htm.div({ 'class': 'self-post' },
                htm.div({ 'class': 'post' },
                    htm.div({ 'class': 'post-head' },
                        htm.div(uibuddies.userDomAttrsSelf()),
                        htm.div({ 'class': 'post-ago' }, ""),
                    ),
                    htm.div({ 'class': 'post-buttons' },
                        htm.button({
                            'type': 'button', 'class': 'button send', 'title': "Send", 'tabindex': 2,
                            'disabled': depends(button_disabled), 'onclick': (() => sendNew(me, files_to_post, is_empty)),
                        }),
                        htm.button({
                            'type': 'button', 'class': 'button attach', 'title': `Add files (max ${yo.reqMaxReqPayloadSizeMb}MB per post).

Don't share privacy-sensitive/highly-personal stuff (if you care), we don't protect uploaded files all that strongly, other than blocking off unauthenticated public/anonymous access attempts.`, 'tabindex': 3,
                            'disabled': depends(button_disabled), 'onclick': () => htm_input_file.click(),
                        }),
                    ),
                    htm_input_file,
                    htm.div({},
                        htm_post_entry,
                        up_files_own.domNode,
                    ),
                ),
            ),
        ),
        numFreshPosts: 0,
        posts: youi.domLive<PostAug>(htm.div({ 'class': 'feed' }), [], (post) => {
            let inner_html = post.Htm ?? ''
            const post_files = (post.Files ?? []).filter(_ => !_.includes('__yodata__'))
            if (post_files && post_files.length) {
                const htm_files = htm.div({ 'class': 'kaffe-post-files' },
                    ...post_files.map((file_name_full, idx) => {
                        const idx_sep = file_name_full.indexOf("__yo__")
                        const file_name_show = (idx_sep < 0) ? file_name_full : file_name_full.substring(idx_sep + "__yo__".length)
                        const file_content_type = post.FileContentTypes![idx],
                            file_url = `/_postfiles/${encodeURIComponent(file_name_full)}`
                        const htm_file_link = htm.a({ 'class': 'kaffe-post-file', 'target': '_blank', 'href': file_url, 'data-filename': file_name_show, 'data-filetype': file_content_type })
                        if (file_content_type !== "") {
                            const icon = fileContentTypeIcons[file_content_type.substring(0, file_content_type.indexOf('/'))]
                            van.add(htm_file_link, (icon !== fileContentTypeIcons['image']) ? htm.div({}, icon)
                                : htm.div({ 'class': 'image', 'style': `background-image:url('${file_url}')` }))
                        }
                        van.add(htm_file_link, htm.span({}, file_name_show, (file_content_type === "") ? undefined : htm.span({}, file_content_type)))
                        return htm_file_link
                    })
                )
                inner_html += htm_files.outerHTML
            }

            const htm_post = htm.div({ 'class': depends(() => ('post-content' + ((post._isDel || (me.isRequestingDeletion.val === (post.Id!))) ? ' deleting' : (post._isFresh ? ' fresh' : '')))) })
            htm_post.innerHTML = inner_html
            const htm_file_link_nodes = htm_post.querySelectorAll("a[data-filetype]")
            if (htm_file_link_nodes)
                htm_file_link_nodes.forEach((a) => {
                    if (a.getAttribute('data-filetype')!.startsWith('image/'))
                        (a as HTMLAnchorElement).onclick = () => openMediaPopup(a.getAttribute('href')!, a.getAttribute('data-filename')!, false)
                    else if (a.getAttribute('data-filetype')!.startsWith('video/'))
                        (a as HTMLAnchorElement).onclick = () => openMediaPopup(a.getAttribute('href')!, a.getAttribute('data-filename')!, true)
                })

            const post_by = kaffe.userByPost(post), post_dt = new Date(post.DtMade!)
            const is_own_post = (post_by?.Id === kaffe.userSelf.val?.Id) || false,
                dt_str = post_dt.toLocaleDateString() + " at " + post_dt.toLocaleTimeString()
            return htm.div({ 'class': 'post', 'title': dt_str },
                htm.div({ 'class': 'post-head' },
                    htm.div({
                        ...is_own_post ? uibuddies.userDomAttrsSelf() : uibuddies.userDomAttrsBuddy(post_by, post.By),
                        'onclick': () => kaffe.userShowPopup(is_own_post ? undefined : post_by),
                    }),
                    htm.div({ 'class': 'post-ago', 'title': dt_str }, post._uxStrAgo),
                ),
                htm.div({ 'class': 'post-buttons' },
                    htm.button({
                        'type': 'button', 'class': 'button delete', 'title': "Delete", 'style': `visibility:${is_own_post ? 'visible' : 'hidden'}`,
                        'disabled': depends(button_disabled), 'onclick': () => deletePost(me, post.Id!),
                    }),
                ),
                htm_post,
            )
        })
    }
    van.add(me.DOM, me.posts.domNode)
    return me
}

function openMediaPopup(fileUrl: string, fileNameShow: string, isVideo: boolean) {
    const dialog = htm.dialog({ 'class': 'media-popup' },
        htm.button({ 'type': 'button', 'class': 'close', 'title': "Close", 'onclick': _ => dialog.close() }, "❎"),
        (!isVideo)
            ? htm.img({ 'title': fileNameShow, 'src': fileUrl, 'alt': fileNameShow })
            : htm.video({ 'title': fileNameShow, 'src': fileUrl, 'controls': true, 'loop': true, 'playsinline': true }),
    )
    dialog.onclose = () => dialog.remove()
    van.add(document.body, dialog)
    dialog.showModal()
    return false
}

function hasUpFiles(me: UiCtlPosts) {
    return me.upFilesNative.some(_ => _ !== null)
}

type UpFile = { name: string, type: string, idx: number, size: number, lastModified: number }

function onFilesAdded(me: UiCtlPosts, htmInputFile: HTMLInputElement) {
    const files = me.upFilesOwn.all
    const likely_dupls: string[] = []
    for (let i = 0; i < htmInputFile.files!.length; i++) {
        const file = htmInputFile.files!.item(i)
        if (file) {
            if (files.some(_ => (_.name === file.name) && (_.type === file.type) && (youtil.fEq(_.size, file.size)) && (youtil.fEq(_.lastModified, file.lastModified))))
                likely_dupls.push(file.name)
            files.push({ idx: me.upFilesNative.length, name: file.name, type: file.type, size: file.size, lastModified: file.lastModified })
            me.upFilesNative.push(file)
        }
    }
    if (likely_dupls.length)
        alert("Detected probable (but not byte-by-byte-compared) duplicate files, please double-check and remove any duplicates:\n\n· " + likely_dupls.join('\n· '))

    me.upFilesOwn.replaceWith(files)
}

function removeUpFile(me: UiCtlPosts, upFile: UpFile) {
    me.upFilesNative[upFile.idx] = null
    me.upFilesOwn.replaceWith(me.upFilesOwn.all.filter(_ => (_.idx !== upFile.idx)))
}

async function deletePost(me: UiCtlPosts, postId: number) {
    if (!confirm("Sure to delete?"))
        return
    const post_idx = me.posts.all.findIndex(_ => (_ && (_.Id === postId)))
    if (post_idx < 0)
        return
    me.isRequestingDeletion.val = postId
    const post = me.posts.all[post_idx]
    if (post)
        post._isDel = true
    await kaffe.deletePost(postId)
    update(me, me.posts.all)
    me.isRequestingDeletion.val = 0
}

async function sendNew(me: UiCtlPosts, upFilesOwn: UpFile[], isEmptyStateToSet: State<boolean>) {
    const post_html = htmlToSend(me)
    if ((!(post_html && post_html.length)) && !hasUpFiles(me))
        return

    me.isSending.val = true
    const ok = await kaffe.sendNewPost(post_html, me.upFilesNative.filter(_ => (_ ? true : false)).map(_ => (_ as File)))
    me.isSending.val = false
    if (ok) {
        me._htmPostInput.innerHTML = ''
        me.upFilesNative = []
        me.upFilesOwn.replaceWith([])
        window.scrollTo(0, 0)
        isEmptyStateToSet.val = true
    }
    me._htmPostInput.focus()
}

function htmlToSend(me: UiCtlPosts, ignoreOffline?: boolean) {
    if (kaffe.isSeeminglyOffline.val && !ignoreOffline)
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
    const now = Date.now(), old_posts = me.posts.all.filter(_ => true)
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
            const user_cur = kaffe.userSelf
            const is_fresh = (old_posts.length > 0) && ((kaffe.browserTabInvisibleSince === 0)
                ? ((now - post_time) < freshnessDurationMsWhenVisible)
                : (post_time >= kaffe.browserTabInvisibleSince)) && ((!user_cur.val) || (post.By!) != (user_cur.val.Id))
            if (is_fresh)
                num_fresh++
            const ret: PostAug = { ...post, _uxStrAgo: post_ago_str, _isDel: false, _isFresh: is_fresh }
            if (post_ago_str !== "")
                last_ago_str = post_ago_str
            return ret
        })
    me.numFreshPosts = num_fresh
    if (!youtil.deepEq(old_posts, fresh_feed, true, false))
        me.posts.replaceWith(fresh_feed)
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
