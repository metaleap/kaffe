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
			Fails{Err: "ExpectedNonEmptyPost", If: PostMd.Equal("").And(q.ArrIsEmpty(PostFiles))},
			Fails{Err: "RepliedToPostDoesNotExist", If: PostRepl.LessThan(0)},
			Fails{Err: "ExpectedOnlyBuddyRecipients", If: q.ArrAreAnyIn(PostTo, q.OpLeq, 0)},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),
		"recentUpdates": apiRecentUpdates.
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),
	})
}

type Post struct {
	Id     yodb.I64
	DtMade *yodb.DateTime
	DtMod  *yodb.DateTime

	By    yodb.Ref[User, yodb.RefOnDelCascade]
	To    yodb.Arr[yodb.I64]
	Md    yodb.Text
	Files yodb.Arr[yodb.Text]
	Repl  yodb.Ref[Post, yodb.RefOnDelCascade]
}

type RecentUpdates struct {
	Posts   []*Post
	Buddies bool
	Since   *yodb.DateTime
	Next    *yodb.DateTime
}

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
		if sl.Any(post.To, func(it yodb.I64) bool { return !sl.Has(it, user.Buddies) }) {
			panic(ErrPostNew_ExpectedOnlyBuddyRecipients)
		}
	}

	return yodb.CreateOne(ctx, post)
}

func getRecentUpdates(ctx *Ctx, forUser *User, since *yodb.DateTime) *RecentUpdates {
	const max_posts_to_fetch_if_just_checked = 2
	is_first_fetch_in_session, max_posts_to_fetch := (since == nil), max_posts_to_fetch_if_just_checked
	if is_first_fetch_in_session {
		if since = forUser.LastSeen; since == nil {
			since = forUser.DtMod
		}
	}
	max_posts_to_fetch = If((since.SinceNow() > 11*time.Hour), 123,
		If((since.SinceNow() > time.Hour), 44, If(is_first_fetch_in_session, 22, 2)))
	buddy_ids := forUser.Buddies.Anys()

	ret := &RecentUpdates{Since: since, Next: yodb.DtFrom(time.Now)} // the below outside the ctor to ensure Next is set before hitting the DB
	if (max_posts_to_fetch > max_posts_to_fetch_if_just_checked) || ((time.Now().UnixNano() % 2) == 0) {
		ret.Buddies = yodb.Exists[User](ctx, // any buddies modified themselves?
			UserId.In(buddy_ids...).And(UserDtMod.GreaterOrEqual(since)))
	}
	ret.Posts = yodb.FindMany[Post](ctx,
		PostDtMod.GreaterOrEqual(since).
			And(PostBy.In(buddy_ids...)).
			And(q.ArrIsEmpty(PostTo).Or(q.ArrHas(PostTo, forUser.Id))).
			And(PostRepl.Equal(nil).Or(PostRepl_By.In(append(buddy_ids, forUser.Id)...))),
		max_posts_to_fetch, PostDtMade.Desc())
	return ret
}
