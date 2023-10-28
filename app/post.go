package haxsh

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

type Post struct {
	Id     yodb.I64
	DtMade *yodb.DateTime
	DtMod  *yodb.DateTime

	By    yodb.Ref[User, yodb.RefOnDelCascade]
	To    yodb.Arr[yodb.I64]
	Htm   yodb.Text
	Files yodb.Arr[yodb.Text]

	FileContentTypes []string
}

type PostsListResult struct {
	Posts     []*Post
	NextSince *yodb.DateTime
}

func postsFor(ctx *Ctx, forUser *User, dtFrom time.Time, dtUntil time.Time, onlyThoseBy []yodb.I64) (ret []*Post) {
	year := dtFrom.Year()
	if (year < 2023) || ((year == 2023) && (dtFrom.Month() < 10)) || (dtUntil.Equal(dtFrom)) || (dtUntil.Before(dtFrom)) || (dtUntil.Sub(dtFrom) > (time.Hour * 24 * 33)) {
		panic(ErrPostsForPeriod_ExpectedPeriodGreater0AndLess33Days)
	}
	query := dbQueryPostsForUser(forUser, onlyThoseBy).And(PostDtMade.GreaterOrEqual(dtFrom)).And(PostDtMade.LessOrEqual(dtUntil))
	return yodb.FindMany[Post](ctx, query, 0, nil, PostDtMade.Desc())
}

func postsRecent(ctx *Ctx, forUser *User, since *yodb.DateTime, onlyThoseBy []yodb.I64) *PostsListResult {
	if (since != nil) && (since.Time().After(time.Now()) || since.Time().Before(*forUser.DtMade.Time())) {
		since = nil
	}

	ret := &PostsListResult{NextSince: yodb.DtNow()} // NextSince=now must happen before hitting the DB
	query_posts_for_user := dbQueryPostsForUser(forUser, onlyThoseBy)
	if since == nil {
		since = yodb.DtFrom(time.Now().AddDate(0, 0, -1))
	}
	ret.Posts = yodb.FindMany[Post](ctx, query_posts_for_user.And(PostDtMod.GreaterOrEqual(since)),
		If(time.Since(*since.Time()) > (23*time.Hour), 11, 0), nil, PostDtMade.Desc())
	if (since == nil) && (len(ret.Posts) == 0) {
		ret.Posts = yodb.FindMany[Post](ctx, query_posts_for_user, 11, nil, PostDtMade.Desc())
	}
	if len(onlyThoseBy) > 0 {
		did_mut := false
		for _, user_id := range onlyThoseBy {
			if user_id != forUser.Id {
				did_mut, forUser.ByBuddyDtLastMsgCheck[str.FromI64(int64(user_id), 10)] = true, time.Time(*ret.NextSince)
			}
		}
		if did_mut {
			ctx.Set("by_buddy_last_msg_check", forUser.ByBuddyDtLastMsgCheck)
		}
	}
	return ret
}

func postsDeleted(ctx *Ctx, postIds []yodb.I64) (ret []yodb.I64) {
	existing_posts := yodb.FindMany[Post](ctx, PostId.In(sl.ToAnys(postIds)...), 0, []q.F{PostId.F()})
	for _, post_id := range postIds {
		if !sl.HasWhere(existing_posts, func(it *Post) bool { return it.Id == post_id }) {
			ret = append(ret, post_id)
		}
	}
	return
}

func postDelete(ctx *Ctx, postId yodb.I64) bool {
	return (yodb.Delete[Post](ctx, PostId.Equal(postId)) > 0)
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
		post.To = sl.Sorted(sl.With(post.To, user.Id))
	}

	return yodb.CreateOne(ctx, post)
}

func dbQueryPostsForUser(forUser *User, onlyThoseBy sl.Slice[yodb.I64]) q.Query {
	is_room := (len(onlyThoseBy) > 0)
	if !is_room {
		onlyThoseBy = (sl.Slice[yodb.I64])(forUser.Buddies)
	}
	onlyThoseBy = sl.With(onlyThoseBy, forUser.Id)
	ret := PostBy.In(onlyThoseBy.ToAnys()...).And(If(is_room,
		PostTo.Equal(sl.Sorted(onlyThoseBy)),
		q.ArrIsEmpty(PostTo),
	))
	return ret
}
