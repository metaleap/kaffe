package haxsh

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
)

func init() {
	Apis(ApiMethods{
		"postNew": apiPostNew.Checks(
			Fails{Err: "ExpectedNonEmptyPost", If: PostMd.Equal("").And(PostFiles.Len().Equal(0))},
			Fails{Err: "RepliedToPostDoesNotExist", If: PostRepl.LessThan(0)},
			Fails{Err: "InvalidItemInFiles", If: PostFiles.ArrAny(q.Equal, "").Equal(true)},
			Fails{Err: "ExpectedOnlyBuddyRecipients", If: PostTo.ArrAny(q.LessOrEqual, 0).Equal(true)},
		).
			FailIf(ErrUnauthorized, yoauth.CurrentlyNotLoggedIn),
		"recentUpdates": apiRecentUpdates.
			FailIf(ErrUnauthorized, yoauth.CurrentlyNotLoggedIn),
	})
}

type Post struct {
	Id     yodb.I64
	DtMade *yodb.DateTime
	DtMod  *yodb.DateTime

	By    yodb.Ref[User, yodb.RefOnDelCascade]
	To    yodb.Arr[yodb.I64]
	Md    yodb.Text
	Files yodb.Arr[FileRef]
	Repl  yodb.Ref[Post, yodb.RefOnDelCascade]
}

type FileRef string

var apiPostNew = api(func(this *ApiCtx[Post, Return[yodb.I64]]) {
	this.Ret.Result = postNew(this.Ctx, this.Args, true)
})

var apiRecentUpdates = api(func(this *ApiCtx[struct {
	Since *yodb.DateTime
}, RecentUpdates]) {
	user_cur := userCur(this.Ctx)
	if user_cur == nil {
		panic(ErrUnauthorized)
	}
	this.Ret = getRecentUpdates(this.Ctx, user_cur, this.Args.Since)
})

func postNew(ctx *Ctx, post *Post, byCurUserInCtx bool) yodb.I64 {
	ctx.DbTx()

	var user *User
	post_by_user_id := post.By.Id()
	if byCurUserInCtx || (post_by_user_id <= 0) {
		user = userCur(ctx)
		post_by_user_id = user.Id
		post.By.SetId(user.Id)
	}
	if post_by_user_id <= 0 {
		panic(ErrUnauthorized)
	}
	if user == nil {
		if user = userById(ctx, post_by_user_id); user == nil {
			panic(ErrUnauthorized)
		}
	}

	if in_reply_to := post.Repl.Id(); (in_reply_to > 0) && !yodb.Exists[Post](ctx, PostId.Equal(in_reply_to)) {
		panic(ErrPostNew_RepliedToPostDoesNotExist)
	}

	if len(post.To) > 0 {
		if sl.Any(post.To, func(it yodb.I64) bool { return !sl.Has(user.Buddies, it) }) {
			panic(ErrPostNew_ExpectedOnlyBuddyRecipients)
		}
	}

	return yodb.CreateOne(ctx, post)
}

type RecentUpdates struct {
	Posts   []*Post
	Buddies bool
	Since   *yodb.DateTime
	Next    *yodb.DateTime
}

func getRecentUpdates(ctx *Ctx, forUser *User, since *yodb.DateTime) *RecentUpdates {
	if since == nil {
		if since = forUser.LastSeen; since == nil {
			since = forUser.DtMod
		}
	}
	buddy_ids := append(forUser.Buddies.Anys(), forUser.Id)
	ret := &RecentUpdates{Next: yodb.DtFrom(time.Now)} // the below outside the ctor to ensure Next is set before hitting the DB
	ret.Buddies = yodb.Exists[User](ctx, UserId.In(buddy_ids...).And(UserDtMod.GreaterOrEqual(since)))
	ret.Posts = yodb.FindMany[Post](ctx, PostDtMod.GreaterOrEqual(since).And(PostBy.In(buddy_ids...)), 1234, PostDtMade.Desc())
	return ret
}
