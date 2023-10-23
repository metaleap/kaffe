package haxsh

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

const ctxKeyCurUser = "haxshCurUser"

func init() {
	Apis(ApiMethods{
		"userSignOut": apiUserSignOut.
			CouldFailWith(":" + yoauth.MethodPathLogout),
		"userSignUp": apiUserSignUp.
			CouldFailWith(":"+yoauth.MethodPathRegister, ":userSignIn"),
		"userSignIn": apiUserSignIn.
			CouldFailWith(":" + yoauth.MethodPathLogin),
		"userBy": apiUserBy.Checks(
			Fails{Err: "ExpectedEitherNickNameOrEmailAddr", If: UserByEmailAddr.Equal("").And(UserByNickName.Equal(""))},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),
		"userUpdate": apiUserUpdate.Checks(
			Fails{Err: ErrDbUpdExpectedIdGt0, If: UserUpdateId.LessOrEqual(0)},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized).
			CouldFailWith(":"+yodb.ErrSetDbUpdate, "NicknameAlreadyExists"),
		"userBuddies": apiUserBuddies.
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),
	})
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

var apiUserSignIn = api(func(this *ApiCtx[yoauth.ApiAccountPayload, Void]) {
	Do(yoauth.ApiUserLogin, this.Ctx, this.Args)
})

var apiUserSignUp = api(func(this *ApiCtx[yoauth.ApiAccountPayload, User]) {
	this.Ctx.DbTx()

	auth := Do(yoauth.ApiUserRegister, this.Ctx, this.Args)
	user := User{LastSeen: yodb.DtNow()}
	user.Auth.SetId(auth.Id)
	user.Id = yodb.CreateOne(this.Ctx, &user)
	// _ = Do(apiUserSignIn, this.Ctx, this.Args)
	this.Ret = &user
})

var apiUserSignOut = api(func(this *ApiCtx[Void, Void]) {
	_ = Do(yoauth.ApiUserLogout, this.Ctx, this.Args)
})

var apiUserBy = api(func(this *ApiCtx[struct {
	EmailAddr string
	NickName  string
}, User]) {
	if this.Args.NickName != "" {
		this.Ret = userByNickName(this.Ctx, this.Args.NickName)
	} else {
		this.Ret = userByEmailAddr(this.Ctx, this.Args.EmailAddr)
	}
})

var apiUserUpdate = api(func(this *ApiCtx[yodb.ApiUpdateArgs[User, UserField], Void]) {
	_, user_auth_id := yoauth.CurrentlyLoggedInUser(this.Ctx)
	this.Args.Changes.Id = this.Args.Id
	if user_auth_id != this.Args.Changes.Auth.Id() {
		panic(ErrUnauthorized)
	}
	userUpdate(this.Ctx, &this.Args.Changes, true, (len(this.Args.ChangedFields) > 0), this.Args.ChangedFields...)
})

var apiUserBuddies = api(func(this *ApiCtx[Void, Return[[]*User]]) {
	this.Ret.Result = userBuddies(this.Ctx, userCur(this.Ctx))
})

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
	return yodb.FindMany[User](ctx, UserId.In(forUser.Buddies.Anys()...), 0, UserDtMod.Desc())
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
