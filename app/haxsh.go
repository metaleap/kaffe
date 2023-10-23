package haxsh

import (
	"time"

	. "yo/ctx"
	yodb "yo/db"
	. "yo/srv"
	. "yo/util"
)

var devModeInitMockUsers func()

func init() {
	AppApiUrlPrefix = "_/"
	AppSideStaticRePathFor = func(requestPath string) string {
		return "__static/haxsh.html"
	}
}

func Init() {
	yodb.Ensure[User, UserField]("", nil, false,
		yodb.ReadOnly[UserField]{UserAuth},
		yodb.Unique[UserField]{UserAuth, UserNick},
		yodb.NoUpdTrigger[UserField]{UserLastSeen},
	)
	yodb.Ensure[Post, PostField]("", nil, false,
		yodb.ReadOnly[PostField]{PostBy, PostRepl},
		yodb.Index[PostField]{PostBy, PostTo},
	)
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}

func fetchRecentUpdates(ctx *Ctx, forUser *User, since *yodb.DateTime) *RecentUpdates {
	const max_posts_to_fetch_if_just_checked = 2
	is_first_fetch_in_session, max_posts_to_fetch := (since == nil), max_posts_to_fetch_if_just_checked
	if is_first_fetch_in_session {
		if since = forUser.LastSeen; since == nil {
			since = forUser.DtMod
		}
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
