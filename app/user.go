package haxsh

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
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
	})
}

type User struct {
	Id      yodb.I64
	Created *yodb.DateTime

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
	var user User
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
	_ = Do(yoauth.ApiUserLogin, this.Ctx, this.Args)
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

func UserUpdate(ctx *Ctx, upd *User, inclEmptyOrMissingFields bool) bool {
	ctx.DbTx()
	if upd.Btw.Do(str.Trim); (upd.Btw != "") && (upd.BtwDt == nil) {
		upd.BtwDt = yodb.DtFrom(time.Now)
	}
	upd.Buddies.EnsureAllUnique()
	if upd.Nick.Do(str.Trim); upd.Nick != "" {
		if other := yodb.FindOne[User](ctx, UserColNick.Equal(upd.Nick)); (other != nil) && ((other.Id != upd.Id) || (other.Auth.Id() != upd.Auth.Id())) {
			panic(ErrUserUpdate_NicknameAlreadyExists)
		}
	}
	return (yodb.Update[User](ctx, upd, inclEmptyOrMissingFields, nil) > 0)
}

func UserByEmailAddr(ctx *Ctx, emailAddr string) (ret *User) {
	// TODO: UserColAuth.Where(yoauth.UserAuthColEmailAddr.Equal(emailAddr))
	// or alt: UserColAuthEmailAddr.Equal(emailAddr)
	if user_auth := yodb.FindOne[yoauth.UserAuth](ctx, yoauth.UserAuthColEmailAddr.Equal(emailAddr)); user_auth != nil {
		ret = yodb.FindOne[User](ctx, UserFieldAuth.Equal(user_auth.Id))
	}
	return
}

func UserCur(ctx *Ctx) (ret *User) {
	if ret = ctx.Get(ctxKeyCurUser, nil).(*User); ret == nil {
		if _, user_auth_id := yoauth.CurrentlyLoggedInUser(ctx); user_auth_id != 0 {
			ret = yodb.FindOne[User](ctx, UserColAuth.Equal(user_auth_id))
			ctx.Set(ctxKeyCurUser, ret)
		}
	}
	return
}
