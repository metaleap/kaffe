package haxsh

import (
	"time"
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
)

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

		"postsRecent": apiPostsRecent.
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),

		"postsForPeriod": apiPostsForPeriod.Checks(
			Fails{Err: "ExpectedPeriodGreater0AndLess33Days", If: PostsForPeriodFrom.Equal(nil).Or(PostsForPeriodUntil.Equal(nil)).Or(PostsForPeriodUntil.LessOrEqual(PostsForPeriodFrom))},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),

		"postNew": apiPostNew.Checks(
			Fails{Err: "ExpectedNonEmptyPost", If: PostMd.Equal("").And(q.ArrIsEmpty(PostFiles))},
			Fails{Err: "RepliedToPostDoesNotExist", If: PostRepl.LessThan(0)},
			Fails{Err: "ExpectedOnlyBuddyRecipients", If: q.ArrAreAnyIn(PostTo, q.OpLeq, 0)},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),
	})
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
	} else if this.Args.EmailAddr != "" {
		this.Ret = userByEmailAddr(this.Ctx, this.Args.EmailAddr)
	} else {
		panic(ErrUserBy_ExpectedEitherNickNameOrEmailAddr)
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
	this.Ret.Result = userBuddies(this.Ctx, userCur(this.Ctx), true)
})

var apiPostsRecent = api(func(this *ApiCtx[struct {
	Since *yodb.DateTime
}, RecentUpdates]) {
	user_cur := userCur(this.Ctx)
	if user_cur == nil {
		panic(ErrUnauthorized)
	}
	this.Ret = postsRecent(this.Ctx, user_cur, this.Args.Since)
})

type ApiArgPeriod struct {
	From  *time.Time
	Until *time.Time
}

var apiPostsForPeriod = api(func(this *ApiCtx[ApiArgPeriod, Void]) {
	if (this.Args.From == nil) || (this.Args.Until == nil) {
		panic(ErrPostsForPeriod_ExpectedPeriodGreater0AndLess33Days)
	}
	postsFor(this.Ctx, userCur(this.Ctx), *this.Args.From, *this.Args.Until)
})

var apiPostNew = api(func(this *ApiCtx[Post, Return[yodb.I64]]) {
	this.Ret.Result = postNew(this.Ctx, this.Args, true)
})
