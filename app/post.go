package haxsh

import (
	"sync"
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
	Posts        []*Post
	NextSince    *yodb.DateTime
	UnreadCounts map[string]int64
}

func postsFor(ctx *Ctx, forUser *User, dtFrom time.Time, dtUntil time.Time, onlyThoseBy []yodb.I64) (ret []*Post) {
	year := dtFrom.Year()
	if (year < 2023) || ((year == 2023) && (dtFrom.Month() < 10)) || (dtUntil.Equal(dtFrom)) || (dtUntil.Before(dtFrom)) || (dtUntil.Sub(dtFrom) > (time.Hour * 24 * 33)) {
		panic(ErrPostsForPeriod_ExpectedPeriodGreater0AndLess33Days)
	}
	query := dbQueryPostsForUser(forUser, onlyThoseBy).And(PostDtMade.GreaterOrEqual(dtFrom)).And(PostDtMade.LessThan(dtUntil))
	return yodb.FindMany[Post](ctx, query, 0, nil, PostDtMade.Desc())
}

func postPeriods(ctx *Ctx, forUser *User, with []yodb.I64) (ret []time.Time) {
	now_year, now_month, _ := time.Now().Date()
	query := dbQueryPostsForUser(forUser, with)
	post_earliest := yodb.FindOne[Post](ctx, query.And(PostDtMade.GreaterOrEqual(forUser.DtMade)), PostDtMade.Asc())
	if post_earliest != nil {
		year, month, day := post_earliest.DtMade.Time().Date()
		ret = append(ret, time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
		for {
			if month++; month > time.December {
				year, month = year+1, time.January
			}
			if (year > now_year) || ((year == now_year) && (month > now_month)) {
				break
			}
			ret = append(ret, time.Date(year, month, 1, 0, 0, 0, 0, time.UTC))
		}
	}
	var idxs_to_drop []int
	for i := 1; i < len(ret)-1; i++ {
		have_post_in_month := yodb.Exists[Post](ctx, query.And(PostDtMade.GreaterOrEqual(ret[i])).And(PostDtMade.LessThan(ret[i+1])))
		if !have_post_in_month {
			idxs_to_drop = append(idxs_to_drop, i)
		}
	}
	ret = sl.Reversed(sl.WithoutIdxs(ret, idxs_to_drop...))
	return
}

func postsRecent(ctx *Ctx, forUser *User, since *yodb.DateTime, onlyThoseBy []yodb.I64) *PostsListResult {
	if (since != nil) && (since.Time().After(time.Now()) || since.Time().Before(*forUser.DtMade.Time())) {
		since = nil
	}

	ret := &PostsListResult{NextSince: yodb.DtNow(), UnreadCounts: map[string]int64{}} // `NextSince=now` must happen before hitting the DB
	query_posts_for_user := dbQueryPostsForUser(forUser, onlyThoseBy)

	no_since_given := (since == nil)
	if no_since_given {
		since = yodb.DtFrom(time.Now().AddDate(0, 0, -1))
	}
	ret.Posts = yodb.FindMany[Post](ctx, query_posts_for_user.And(PostDtMod.GreaterOrEqual(since)),
		If(time.Since(*since.Time()) > (23*time.Hour), 22, 0 /*TODO: adapt max more fluidly to since*/), nil, PostDtMade.Desc())
	if no_since_given && (len(ret.Posts) == 0) {
		ret.Posts = yodb.FindMany[Post](ctx, query_posts_for_user, 22, nil, PostDtMade.Desc())
	}

	{ // we also populate PostsListResult.UnreadCounts for all buddies
		var mut sync.Mutex
		do_count := func(buddyId yodb.I64, since *yodb.DateTime, onDone func()) {
			defer onDone()
			query := dbQueryPostsForUser(forUser, If(buddyId == 0, nil, sl.Slice[yodb.I64]{buddyId}))
			if since != nil {
				query = query.And(PostDtMade.GreaterOrEqual(since))
			}
			count, key := yodb.Count[Post](ctx, query, "", nil), If(buddyId == 0, "", str.FromI64(int64(buddyId), 10))
			mut.Lock()
			ret.UnreadCounts[key] = count
			mut.Unlock()
		}
		var work sync.WaitGroup
		if len(onlyThoseBy) > 0 {
			work.Add(1)
			go do_count(0, forUser.byBuddyDtLastMsgCheck[""], work.Done)
		}
		work.Add(len(forUser.Buddies))
		for _, buddy_id := range forUser.Buddies {
			go do_count(buddy_id, forUser.byBuddyDtLastMsgCheck[str.FromI64(int64(buddy_id), 10)], work.Done)
		}
		work.Wait()
	}

	// ensure later update of user.ByBuddyDtLastMsgCheck
	{
		for _, user_id := range onlyThoseBy {
			if user_id != forUser.Id {
				forUser.byBuddyDtLastMsgCheck[str.FromI64(int64(user_id), 10)] = ret.NextSince
			}
		}
		if len(onlyThoseBy) == 0 {
			forUser.byBuddyDtLastMsgCheck[""] = ret.NextSince
		}
		ctx.Set("by_buddy_last_msg_check", forUser.byBuddyDtLastMsgCheck)
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
		if sl.Any(post.To, func(it yodb.I64) bool { return !sl.Has(user.Buddies, it) }) {
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
	ret := PostBy.In(onlyThoseBy.ToAnys()...).
		And(PostBy.Equal(forUser.Id).Or(q.InArr(forUser.Id, PostBy_Buddies))).
		And(If(is_room,
			PostTo.Equal(sl.Sorted(onlyThoseBy)),
			q.ArrIsEmpty(PostTo),
		))
	return ret
}
