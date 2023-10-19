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

var checkSignedIn = Pair[Err, func(*Ctx) bool]{ErrUnauthorized, yoauth.CurrentlyLoggedIn}

func init() {
	Apis(ApiMethods{
		"userSignOut": Api(apiUserSignOut, PkgInfo).
			CouldFailWith(":" + yoauth.MethodPathLogout),
		"userSignUp": Api(apiUserSignUp, PkgInfo).
			CouldFailWith(":"+yoauth.MethodPathRegister, ":userSignIn"),
		"userSignIn": Api(apiUserSignIn, PkgInfo).
			CouldFailWith(":" + yoauth.MethodPathLogin),
		"userUpdate": Api(apiUserUpdate, PkgInfo,
			Fails{Err: ErrDbUpdExpectedIdGt0, If: UserUpdateId.LessOrEqual(0)},
		).PreCheck(checkSignedIn).
			CouldFailWith(":"+yodb.ErrSetDbUpdate, ErrDbNotStored, "NicknameAlreadyExists"),
		"userGet": Api(apiUserGet, PkgInfo),
	})
	PreApiHandling = append(PreApiHandling, Middleware{"setUserLastSeen", func(ctx *Ctx) {
		go setUserLastSeen(0, ctx.Get(yoauth.CtxKeyAuthId, yodb.I64(0)).(yodb.I64))
	}})
}

type User struct {
	Id      yodb.I64
	Created *yodb.DateTime

	LastSeen  *yodb.DateTime
	Auth      yodb.Ref[yoauth.UserAuth, yodb.RefOnDelCascade]
	PicFileId yodb.Text
	Nick      yodb.Text
	Btw       yodb.Text
	BtwDt     *yodb.DateTime
	Buddies   yodb.Arr[yodb.I64]
}

func apiUserSignUp(this *ApiCtx[yoauth.ApiAccountPayload, User]) {
	this.Ctx.DbTx()

	auth := Do(yoauth.ApiUserRegister, this.Ctx, this.Args)
	user := User{LastSeen: yodb.DtFrom(time.Now)}
	user.Auth.SetId(auth.Id)
	if user.Id = yodb.CreateOne(this.Ctx, &user); user.Id <= 0 {
		panic(ErrDbNotStored)
	}
	_ = Do(apiUserSignIn, this.Ctx, this.Args)
	this.Ret = &user
}

func apiUserSignOut(this *ApiCtx[Void, Void]) {
	_ = Do(yoauth.ApiUserLogout, this.Ctx, this.Args)
}

func apiUserSignIn(this *ApiCtx[yoauth.ApiAccountPayload, Void]) {
	Do(yoauth.ApiUserLogin, this.Ctx, this.Args)
}

func apiUserUpdate(this *ApiCtx[yodb.ApiUpdateArgs[User], Void]) {
	_, user_auth_id := yoauth.CurrentlyLoggedInUser(this.Ctx)
	this.Args.Changes.Id = this.Args.Id
	if user_auth_id != this.Args.Changes.Auth.Id() {
		panic(ErrUnauthorized)
	}
	if !UserUpdate(this.Ctx, &this.Args.Changes, this.Args.IncludingEmptyOrMissingFields) {
		panic(ErrDbNotStored)
	}
}

func apiUserGet(this *ApiCtx[struct {
	EmailAddr string
}, User]) {
	this.Ret = UserByEmailAddr(this.Ctx, this.Args.EmailAddr)
}

func setUserLastSeen(id yodb.I64, auth_id yodb.I64) {
	ctx := NewCtxNonHttp(time.Minute, "setUserLastSeen")
	defer ctx.OnDone(nil)
	upd := &User{LastSeen: yodb.DtFrom(time.Now), Id: id}
	if auth_id != 0 {
		upd.Auth.SetId(auth_id)
	}
	go UserUpdate(ctx, upd, false, UserLastSeen)
}

func UserUpdate(ctx *Ctx, upd *User, inclEmptyOrMissingFields bool, onlyFields ...UserField) bool {
	ctx.DbTx()
	if upd.Btw.Do(str.Trim); (upd.Btw != "") && (upd.BtwDt == nil) && ((len(onlyFields) == 0) || sl.Has(onlyFields, UserBtw)) {
		if upd.BtwDt = yodb.DtFrom(time.Now); (len(onlyFields) != 0) && !sl.Has(onlyFields, UserBtwDt) {
			onlyFields = append(onlyFields, UserBtwDt)
		}
	}
	if (len(onlyFields) == 0) || sl.Has(onlyFields, UserBuddies) {
		upd.Buddies.EnsureAllUnique()
	}
	if upd.Nick.Do(str.Trim); (upd.Nick != "") && ((len(onlyFields) == 0) || sl.Has(onlyFields, UserNick)) {
		if yodb.Exists[User](ctx, UserNick.Equal(upd.Nick).And(UserId.NotEqual(upd.Id))) {
			panic(ErrUserUpdate_NicknameAlreadyExists)
		}
	}
	return (yodb.Update[User](ctx, upd, nil, !inclEmptyOrMissingFields, sl.To(onlyFields, UserField.F)...) > 0)
}

func UserByEmailAddr(ctx *Ctx, emailAddr string) *User {
	return yodb.FindOne[User](ctx, UserAuth_EmailAddr.Equal(emailAddr))
}

func UserCur(ctx *Ctx) (ret *User) {
	if ret = ctx.Get(ctxKeyCurUser, nil).(*User); ret == nil {
		if _, user_auth_id := yoauth.CurrentlyLoggedInUser(ctx); user_auth_id != 0 {
			ret = yodb.FindOne[User](ctx, UserAuth.Equal(user_auth_id))
			ctx.Set(ctxKeyCurUser, ret)
		}
	}
	return
}
