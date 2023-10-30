package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	. "yo/util"
	"yo/util/sl"
)

func userBuddies(ctx *Ctx, forUser *User, normalizeLastSeenByMinute bool) (buddiesAlready []*User, buddyRequestsMade []*User, buddyRequestsBy []*User) {
	for _, user := range yodb.FindMany[User](ctx,
		UserId.In(forUser.Buddies.ToAnys()...).Or(q.InArr(forUser.Id, UserBuddies)),
		0, nil, UserLastSeen.Desc(), UserDtMod.Desc(),
	) {
		in_our_buddies, in_their_buddies := sl.Has(user.Id, forUser.Buddies), sl.Has(forUser.Id, user.Buddies)
		if in_our_buddies && in_their_buddies {
			buddiesAlready = append(buddiesAlready, user)
		} else if in_their_buddies {
			buddyRequestsBy = append(buddyRequestsBy, user)
		} else if in_our_buddies {
			buddyRequestsMade = append(buddyRequestsMade, user)
		}
	}

	if normalizeLastSeenByMinute {
		for _, buddy := range buddiesAlready {
			if buddy.LastSeen != nil {
				buddy.LastSeen.Set(DtAtZeroSecsUtc)
			}
		}
	}
	for _, buddy_request := range buddyRequestsMade {
		buddy_request.Auth.SetId(0)
		buddy_request.Btw = "(Buddy request still pending)"
		buddy_request.Buddies = nil
		buddy_request.DtMod = buddy_request.DtMade
		buddy_request.LastSeen = buddy_request.DtMade
		buddy_request.PicFileId = ""
	}
	return
}

func userAddBuddy(ctx *Ctx, forUser *User, nickOrEmailAddr string) *User {
	user := yodb.FindOne[User](ctx, UserNick.Equal(nickOrEmailAddr).Or(UserAuth_EmailAddr.Equal(nickOrEmailAddr)))
	if (user != nil) && !sl.Has(user.Id, forUser.Buddies) {
		userUpdate(ctx, &User{Buddies: sl.With(forUser.Buddies, user.Id)}, true, false, UserBuddies)
	}
	return user
}
