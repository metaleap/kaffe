package kaffe

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

const ctxKeyByBuddyLastMsgCheck = "by_buddy_last_msg_check"

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

func postsForMonthUtc(ctx *Ctx, forUser *User, period YearAndMonth, onlyThoseBy []yodb.I64) (ret []*Post) {
	if forUser == nil {
		return
	}
	first_of_that_month := time.Date(int(period.Year), time.Month(period.Month), 1, 0, 0, 0, 0, time.UTC)
	first_of_following_month := first_of_that_month.AddDate(0, 1, 0)
	query := dbQueryPostsForUser(forUser, onlyThoseBy).
		And(PostDtMade.GreaterOrEqual(first_of_that_month)).
		And(PostDtMade.LessThan(first_of_following_month))
	return yodb.FindMany[Post](ctx, query, 0, nil, PostDtMade.Desc())
}

type YearAndMonth struct {
	Year  uint16
	Month uint8
}

func postMonthsUtc(ctx *Ctx, forUser *User, with []yodb.I64) (ret []YearAndMonth) {
	if forUser == nil {
		return
	}
	now_year, now_month, _ := time.Now().Date()
	query := dbQueryPostsForUser(forUser, with)
	post_earliest := yodb.FindOne[Post](ctx, query.And(PostDtMade.GreaterOrEqual(forUser.DtMade)), PostDtMade.Asc())
	// first, gather all months since user sign-up
	if post_earliest != nil {
		year, month, _ := post_earliest.DtMade.Time().Date()
		ret = append(ret, YearAndMonth{Year: uint16(year), Month: uint8(month)})
		for {
			if month++; month > time.December {
				year, month = year+1, time.January
			}
			if (year > now_year) || ((year == now_year) && (month > now_month)) {
				break
			}
			ret = append(ret, YearAndMonth{Year: uint16(year), Month: uint8(month)})
		}
	}
	// now, drop the ones without any posts
	var idxs_to_drop []int
	for i := 1; i < len(ret)-1; i++ {
		first_of_that_month := time.Date(int(ret[i].Year), time.Month(ret[i].Month), 1, 0, 0, 0, 0, time.UTC)
		first_of_following_month := first_of_that_month.AddDate(0, 1, 0)
		have_post_in_month := yodb.Exists[Post](ctx, query.
			And(PostDtMade.GreaterOrEqual(first_of_that_month)).
			And(PostDtMade.LessThan(first_of_following_month)))
		if !have_post_in_month {
			idxs_to_drop = append(idxs_to_drop, i)
		}
	}
	ret = sl.Reversed(sl.WithoutIdxs(ret, idxs_to_drop...))
	return
}

func postsRecent(ctx *Ctx, forUser *User, since *yodb.DateTime, onlyThoseBy []yodb.I64) *PostsListResult {
	if forUser == nil {
		return nil
	}
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
			query := dbQueryPostsForUser(forUser, If(buddyId == 0, nil, sl.Of[yodb.I64]{buddyId}))
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
		ctx.Set(ctxKeyByBuddyLastMsgCheck, forUser.byBuddyDtLastMsgCheck)
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

func postDelete(ctx *Ctx, post *Post) bool {
	if len(post.Files) > 0 {
		yodb.CreateOne[fileDelReq](ctx, &fileDelReq{FileNames: post.Files})
	}
	return (yodb.Delete[Post](ctx, PostId.Equal(post.Id)) > 0)
}

func postNew(ctx *Ctx, post *Post, userById yodb.I64) yodb.I64 {
	var user_by *User
	if userById <= 0 {
		user_by = userCur(ctx)
	} else {
		user_by = yodb.ById[User](ctx, userById)
	}
	if user_by == nil { // no user cookie
		panic(ErrUnauthorized)
	}
	post.By.SetId(user_by.Id)

	post.Htm = yodb.Text(str.Replace(post.Htm.String(), str.Dict{
		"script": "sсriрt", // homoglyphs for c and p so no post has <script> or javascript://
		"style":  "stуlе",  // homoglyphs for y and e so no post changes everyone's stylesheet
	}))
	if len(post.To) > 0 {
		if sl.Any(post.To, func(it yodb.I64) bool { return !sl.Has(user_by.Buddies, it) }) {
			println(str.FmtV(user_by.Buddies), " VS. ", str.FmtV(post.To))
			ctx.DevModeNoCatch = true
			panic(Err__postNew_ExpectedOnlyBuddyRecipients)
		}
		post.To = sl.Sorted(sl.With(post.To, user_by.Id))
	}

	post_id := yodb.CreateOne(ctx, post)
	if (post.By.Id() != elizaUser.id) && sl.Has(post.To, elizaUser.id) {
		elizaReplyShortlyTo(post_id)
	}
	return post_id
}

func dbQueryPostsForUser(forUser *User, onlyThoseBy sl.Of[yodb.I64]) q.Query {
	is_room := (len(onlyThoseBy) > 0)
	if !is_room {
		onlyThoseBy = (sl.Of[yodb.I64])(forUser.Buddies)
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
