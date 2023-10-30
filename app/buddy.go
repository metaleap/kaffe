package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	. "yo/util"
)

func userBuddies(ctx *Ctx, forUser *User, normalizeLastSeenByMinute bool) (buddiesAlready []*User, buddyRequests []*User) {
	buddiesAlready = yodb.FindMany[User](ctx, UserId.In(forUser.Buddies.ToAnys()...).And(
		q.InArr(forUser.Id, UserBuddies)), 0, nil, UserLastSeen.Desc(), UserDtMod.Desc())
	buddyRequests = yodb.FindMany[User](ctx, UserId.In(forUser.Buddies.ToAnys()...).And(
		q.InArr(forUser.Id, UserBuddies).Not()), 0, nil, UserDtMade.Desc())
	if normalizeLastSeenByMinute {
		for _, buddy := range buddiesAlready {
			if buddy.LastSeen != nil {
				buddy.LastSeen.Set(DtAtZeroSecsUtc)
			}
		}
	}
	for _, buddy_request := range buddyRequests {
		buddy_request.Auth.SetId(0)
		buddy_request.Btw = "(Buddy request still pending)"
		buddy_request.Buddies = nil
		buddy_request.DtMod = buddy_request.DtMade
		buddy_request.LastSeen = buddy_request.DtMade
		buddy_request.PicFileId = ""
	}
	return
}
