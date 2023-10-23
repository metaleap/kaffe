package haxsh

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	. "yo/srv"
	"yo/util/sl"
	"yo/util/str"
)

const ctxKeyCurUser = "haxshCurUser"

func init() {
	PreApiHandling = append(PreApiHandling, Middleware{"userSetLastSeen", func(ctx *Ctx) {
		go userSetLastSeen(ctx.Get(yoauth.CtxKeyAuthId, yodb.I64(0)).(yodb.I64))
	}})
}

type User struct {
	Id     yodb.I64
	DtMade *yodb.DateTime
	DtMod  *yodb.DateTime

	LastSeen  *yodb.DateTime
	Auth      yodb.Ref[yoauth.UserAuth, yodb.RefOnDelCascade]
	PicFileId yodb.Text
	Nick      yodb.Text
	Btw       yodb.Text
	Buddies   yodb.Arr[yodb.I64]
}

func userUpdate(ctx *Ctx, upd *User, byCurUserInCtx bool, inclEmptyOrMissingFields bool, onlyFields ...UserField) {
	ctx.DbTx()
	upd.Btw.Do(str.Trim)
	if (len(onlyFields) == 0) || sl.Has(UserBuddies, onlyFields) {
		upd.Buddies.EnsureAllUnique(nil)
	}
	if upd.Nick.Do(str.Trim); (upd.Nick != "") && ((len(onlyFields) == 0) || sl.Has(UserNick, onlyFields)) {
		if yodb.Exists[User](ctx, UserNick.Equal(upd.Nick).And(UserId.NotEqual(upd.Id))) {
			panic(ErrUserUpdate_NicknameAlreadyExists)
		}
	}
	if byCurUserInCtx {
		upd.LastSeen = yodb.DtNow()
	}
	if 0 == yodb.Update[User](ctx, upd, nil, !inclEmptyOrMissingFields, sl.To(onlyFields, UserField.F)...) {
		panic("nochanges in " + str.From(onlyFields) + "?" + str.From(upd) + "vs." + str.From(userCur(ctx)))
	}
}

func userBuddies(ctx *Ctx, forUser *User) []*User {
	return yodb.FindMany[User](ctx, UserId.In(forUser.Buddies.Anys()...), 0, UserLastSeen.Desc())
}

func userByEmailAddr(ctx *Ctx, emailAddr string) *User {
	return yodb.FindOne[User](ctx, UserAuth_EmailAddr.Equal(emailAddr))
}

func userByNickName(ctx *Ctx, nickName string) *User {
	return yodb.FindOne[User](ctx, UserNick.Equal(nickName))
}

func userById(ctx *Ctx, id yodb.I64) *User {
	user_cur, _ := ctx.Get(ctxKeyCurUser, nil).(*User)
	if (user_cur != nil) && (user_cur.Id == id) {
		return user_cur
	}
	return yodb.ById[User](ctx, id)
}

func userCur(ctx *Ctx) (ret *User) {
	if ret, _ = ctx.Get(ctxKeyCurUser, nil).(*User); ret == nil {
		_, user_auth_id := yoauth.CurrentlyLoggedInUser(ctx)
		if user_auth_id != 0 {
			ret = yodb.FindOne[User](ctx, UserAuth.Equal(user_auth_id))
			ctx.Set(ctxKeyCurUser, ret)
		}
	}
	return
}

func userSetLastSeen(auth_id yodb.I64) {
	if auth_id == 0 {
		return
	}
	ctx := NewCtxNonHttp(time.Minute, "userSetLastSeen")
	defer ctx.OnDone(nil)
	ctx.TimingsNoPrintInDevMode = true
	upd := &User{LastSeen: yodb.DtNow()}
	upd.Auth.SetId(auth_id)
	userUpdate(ctx, upd, true, false, UserLastSeen)
}
