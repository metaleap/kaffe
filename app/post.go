package haxsh

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	. "yo/srv"
	"yo/util/sl"
)

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

func postsFor(ctx *Ctx, forUser *User, dtFrom time.Time, dtUntil time.Time) (ret []*Post) {
	if year := dtFrom.Year(); (year < 2023) || ((year == 2023) && (dtFrom.Month() < 10)) || (dtUntil.Equal(dtFrom)) ||
		(dtUntil.Before(dtFrom)) || (dtUntil.Sub(dtFrom) > (time.Hour * 24 * 33)) {
		panic(ErrPostsForPeriod_ExpectedPeriodGreater0AndLess33Days)
	}
	query := dbQueryPostsForUser(forUser).And(PostDtMade.GreaterOrEqual(dtFrom)).And(PostDtMade.LessOrEqual(dtUntil))
	yodb.FindMany[Post](ctx, query, 0, PostDtMade.Desc())
	return
}

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

func dbQueryPostsForUser(forUser *User) q.Query {
	buddy_ids := forUser.Buddies.Anys()
	return PostBy.In(buddy_ids...).
		And(q.ArrIsEmpty(PostTo).Or(q.ArrHas(PostTo, forUser.Id))).
		And(PostRepl.Equal(nil).Or(PostRepl_By.In(append(buddy_ids, forUser.Id)...)))
}
