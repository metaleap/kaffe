package haxsh

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	. "yo/srv"
	. "yo/util"
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
	Posts []*Post
	Since *yodb.DateTime
	Next  *yodb.DateTime
}

func dtBeforeApp(dt time.Time) bool {
	year := dt.Year()
	return (year < 2023) || ((year == 2023) && (dt.Month() < 10))
}

func postsFor(ctx *Ctx, forUser *User, dtFrom time.Time, dtUntil time.Time) (ret []*Post) {
	if dtBeforeApp(dtFrom) || (dtUntil.Equal(dtFrom)) || (dtUntil.Before(dtFrom)) || (dtUntil.Sub(dtFrom) > (time.Hour * 24 * 33)) {
		panic(ErrPostsForPeriod_ExpectedPeriodGreater0AndLess33Days)
	}
	query := dbQueryPostsForUser(forUser).And(PostDtMade.GreaterOrEqual(dtFrom)).And(PostDtMade.LessOrEqual(dtUntil))
	yodb.FindMany[Post](ctx, query, 0, PostDtMade.Desc())
	return
}

func postsRecent(ctx *Ctx, forUser *User, since *yodb.DateTime) *RecentUpdates {
	const max_posts_to_fetch_if_just_checked = 2
	if (since != nil) && (since.Time().After(time.Now()) || dtBeforeApp(*since.Time())) {
		since = nil
	}
	is_first_fetch_in_session, max_posts_to_fetch := (since == nil), max_posts_to_fetch_if_just_checked
	if is_first_fetch_in_session {
		since = yodb.DtFrom(time.Now().AddDate(0, 0, -1))
	}
	since = forUser.DtMade
	max_posts_to_fetch = If((since.SinceNow() > 11*time.Hour), 123,
		If((since.SinceNow() > time.Hour), 44, If(is_first_fetch_in_session, 22, 2)))

	ret := &RecentUpdates{Since: since, Next: yodb.DtNow()} // the below outside the ctor to ensure Next is set before hitting the DB
	ret.Posts = yodb.FindMany[Post](ctx,
		PostDtMod.GreaterOrEqual(since).And(dbQueryPostsForUser(forUser)),
		max_posts_to_fetch, PostDtMade.Desc())
	return ret
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
