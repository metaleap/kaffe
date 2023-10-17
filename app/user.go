package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
)

const ctxKeyCurUser = "haxshCurUser"

var checkSignedIn = Pair[Err, func(*Ctx) bool]{ErrUnauthorized, yoauth.CurrentlyLoggedIn}

func init() {
	yodb.Ensure[User, UserField]("", nil)
	yodb.Ensure[Post, PostField]("", nil)

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
			CouldFailWith(":"+yodb.ErrSetDbUpdate, ErrDbNotStored),
	})
}

type User struct {
	Id      yodb.I64
	Created *yodb.DateTime

	Auth     yodb.Ref[yoauth.UserAuth, yodb.RefOnDelCascade]
	NickName yodb.Text
	Btw      yodb.Text
	BwtDt    *yodb.DateTime
	Buddies  yodb.Arr[int64]
}

type Post struct {
	Id      yodb.I64
	Created *yodb.DateTime

	Md    yodb.Text
	Files yodb.Arr[struct {
		Id   string
		Name string
	}]

	User yodb.Ref[User, yodb.RefOnDelCascade]
	Rcpt yodb.Arr[yodb.I64]
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
	yodb.Update[User](this.Ctx, &this.Args.Changes, this.Args.IncludingEmptyOrMissingFields, nil)
}

func CurUser(ctx *Ctx) (ret *User) {
	if ret = ctx.Get(ctxKeyCurUser, nil).(*User); ret == nil {
		if _, user_auth_id := yoauth.CurrentlyLoggedInUser(ctx); user_auth_id != 0 {
			ret = yodb.FindOne[User](ctx, UserColAuth.Equal(user_auth_id))
			ctx.Set(ctxKeyCurUser, ret)
		}
	}
	return
}
