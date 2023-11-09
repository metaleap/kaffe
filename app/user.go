package kaffe

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

const ctxKeyCurUser = "kaffeCurUser"

func init() {
	PostApiHandling = append(PostApiHandling, Middleware{"userSetLastSeen", func(ctx *Ctx) {
		if (ctx.Http.UrlPath == AppApiUrlPrefix+apiMethodNameUserUpdate) || (ctx.Http.UrlPath == AppApiUrlPrefix+apiMethodNameUserBuddiesAdd) {
			return // dont set last-seen for calls that already just did it
		}
		if _, user_auth_id := yoauth.CurrentlyLoggedInUser(ctx); user_auth_id > 0 {
			by_buddy_last_msg_check, _ := ctx.Get(ctxKeyByBuddyLastMsgCheck, nil).(yodb.JsonMap[*yodb.DateTime])
			go userSetLastSeen(user_auth_id, by_buddy_last_msg_check)
		}
	}})
}

type User struct {
	Id     yodb.I64
	DtMade *yodb.DateTime
	DtMod  *yodb.DateTime

	LastSeen              *yodb.DateTime
	Auth                  yodb.Ref[yoauth.UserAuth, yodb.RefOnDelCascade]
	PicFileId             yodb.Text
	Nick                  yodb.Text
	Btw                   yodb.Text
	Buddies               yodb.Arr[yodb.I64]
	byBuddyDtLastMsgCheck yodb.JsonMap[*yodb.DateTime]
	vip                   yodb.Bool // if true, posts & files stay around forever, dont get wiped after x days
	gravatarChecked       yodb.Bool

	BtwEmoji string // for API consumers, not in DB
	Offline  bool   // dito
}

func userUpdate(ctx *Ctx, upd *User, inclEmptyOrMissingFields bool, onlyFields ...UserField) {
	upd.Btw.Set(str.Trim)
	if (len(onlyFields) == 0) || sl.Has(onlyFields, UserBuddies) {
		upd.Buddies.EnsureAllUnique(nil)
	}
	if upd.Nick = yodb.Text(str.Replace(string(upd.Nick), str.Dict{"@": ""})); (len(onlyFields) == 0) || sl.Has(onlyFields, UserNick) {
		if upd.Nick.Set(str.Trim); upd.Nick == "" {
			panic(ErrUserUpdate_ExpectedNonEmptyNickname)
		}
		ctx.DbTx()
		if yodb.Exists[User](ctx, UserNick.Equal(upd.Nick).And(UserId.NotEqual(upd.Id))) {
			panic(ErrUserUpdate_NicknameAlreadyExists)
		}
	}
	if upd.LastSeen = yodb.DtNow(); len(onlyFields) > 0 {
		onlyFields = sl.With(onlyFields, UserLastSeen)
	}
	_ = yodb.Update[User](ctx, upd, nil, !inclEmptyOrMissingFields, UserFields(onlyFields...)...)
}

func userByEmailAddr(ctx *Ctx, emailAddr string) *User {
	return yodb.FindOne[User](ctx, UserAuth_EmailAddr.Equal(emailAddr)).augmentAfterLoaded()
}

func userByNickName(ctx *Ctx, nickName string) *User {
	return yodb.FindOne[User](ctx, UserNick.Equal(nickName)).augmentAfterLoaded()
}

func userById(ctx *Ctx, id yodb.I64) *User {
	user, _ := ctx.Get(ctxKeyCurUser, nil).(*User) // maybe `id` points to current-user anyway?
	if (user == nil) || (user.Id != id) {
		user = yodb.ById[User](ctx, id).augmentAfterLoaded()
	}
	return user
}

func userCur(ctx *Ctx) (ret *User) {
	if ret, _ = ctx.Get(ctxKeyCurUser, nil).(*User); ret == nil {
		_, user_auth_id := yoauth.CurrentlyLoggedInUser(ctx)
		if user_auth_id > 0 {
			ret = yodb.FindOne[User](ctx, UserAuth.Equal(user_auth_id)).augmentAfterLoaded()
			ctx.Set(ctxKeyCurUser, ret)
		}
	}
	return
}

func (me *User) augmentAfterLoaded() *User {
	if me != nil {
		me.Offline, me.BtwEmoji = true, emojiUnescaped(string(me.Btw))
		if me.LastSeen != nil {
			me.Offline = time.Since(*me.LastSeen.Time()) > (11 * time.Second)
		}
	}
	return me
}

func userSetLastSeen(auth_id yodb.I64, byBuddyDtLastMsgCheck yodb.JsonMap[*yodb.DateTime]) {
	ctx := NewCtxNonHttp(time.Second, false, "userSetLastSeen")
	defer func() { _ = recover(); ctx.OnDone(nil) }() // for total silence of this operation on errs even in dev-mode outputs (rare tho it is)
	ctx.DbNoLoggingInDevMode()
	ctx.TimingsNoPrintInDevMode = true
	ctx.ErrNoNotifyOf = []Err{ErrTimedOut}
	upd := &User{byBuddyDtLastMsgCheck: byBuddyDtLastMsgCheck}
	upd.Auth.SetId(auth_id)
	// upd.LastSeen = yodb.DtNow() // userUpdate call does it anyway
	only_fields := []UserField{UserLastSeen}
	if byBuddyDtLastMsgCheck != nil {
		only_fields = append(only_fields, userByBuddyDtLastMsgCheck)
	}
	userUpdate(ctx, upd, false, only_fields...)
}
