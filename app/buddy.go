package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	. "yo/util"
)

func userBuddies(ctx *Ctx, forUser *User, normalizeLastSeenByMinute bool) (buddiesAlready []*User, buddyRequestsMade []*User, buddyRequestsBy []*User) {
	buddiesAlready = yodb.FindMany[User](ctx, UserId.In(forUser.Buddies.ToAnys()...).And(
		q.InArr(forUser.Id, UserBuddies)), 0, nil, UserLastSeen.Desc(), UserDtMod.Desc())
	buddyRequestsMade = yodb.FindMany[User](ctx, UserId.In(forUser.Buddies.ToAnys()...).And(
		q.InArr(forUser.Id, UserBuddies).Not()), 0, UserFields(UserId, UserNick, UserDtMade), UserDtMade.Desc())
	buddyRequestsBy = yodb.FindMany[User](ctx, q.InArr(UserBuddies, forUser.Id), 0, nil)
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
