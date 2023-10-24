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
}

type RecentUpdates struct {
	Posts []*Post
	Since *yodb.DateTime
	Next  *yodb.DateTime
}

func postsFor(ctx *Ctx, forUser *User, dtFrom time.Time, dtUntil time.Time) (ret []*Post) {
	year := dtFrom.Year()
	if (year < 2023) || ((year == 2023) && (dtFrom.Month() < 10)) || (dtUntil.Equal(dtFrom)) || (dtUntil.Before(dtFrom)) || (dtUntil.Sub(dtFrom) > (time.Hour * 24 * 33)) {
		panic(ErrPostsForPeriod_ExpectedPeriodGreater0AndLess33Days)
	}
	query := dbQueryPostsForUser(forUser).And(PostDtMade.GreaterOrEqual(dtFrom)).And(PostDtMade.LessOrEqual(dtUntil))
	return yodb.FindMany[Post](ctx, query, 0, PostDtMade.Desc())
}

func postsRecent(ctx *Ctx, forUser *User, since *yodb.DateTime) *RecentUpdates {
	if (since != nil) && (since.Time().After(time.Now()) || since.Time().Before(*forUser.DtMade.Time())) {
		since = nil
	}

	ret := &RecentUpdates{Since: forUser.DtMade, Next: yodb.DtNow()} // the below outside the ctor to ensure Next is set before hitting the DB
	query_posts_for_user := dbQueryPostsForUser(forUser)
	if since == nil {
		since = yodb.DtFrom(time.Now().AddDate(0, 0, -1))
	}
	ret.Posts = yodb.FindMany[Post](ctx, query_posts_for_user.And(PostDtMod.GreaterOrEqual(since)), 0, PostDtMade.Desc())
	if (since == nil) && (len(ret.Posts) == 0) {
		ret.Posts = yodb.FindMany[Post](ctx, dbQueryPostsForUser(forUser), 11, PostDtMade.Desc())
	}
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
		And(q.ArrIsEmpty(PostTo).Or(q.ArrHas(PostTo, forUser.Id)))
}
