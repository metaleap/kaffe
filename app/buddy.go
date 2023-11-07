package haxsh

import (
	"math"
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	. "yo/util"
	"yo/util/sl"
)

func userBuddies(ctx *Ctx, forUser *User, normalizeLastSeenBySecond bool) (buddiesAlready []*User, buddyRequestsMade []*User, buddyRequestsBy []*User) {
	query := q.InArr(forUser.Id, UserBuddies)
	if len(forUser.Buddies) > 0 {
		query = query.Or(UserId.In(forUser.Buddies.ToAnys()...))
	}
	for _, user := range yodb.FindMany[User](ctx, query, 0, nil, UserDtMod.Desc()) {
		in_our_buddies, in_their_buddies := sl.Has(forUser.Buddies, user.Id), sl.Has(user.Buddies, forUser.Id)
		if in_our_buddies && in_their_buddies {
			buddiesAlready = append(buddiesAlready, user.augmentAfterLoaded())
		} else if in_their_buddies {
			buddyRequestsBy = append(buddyRequestsBy, user)
		} else if in_our_buddies {
			buddyRequestsMade = append(buddyRequestsMade, user)
		}
	}

	if normalizeLastSeenBySecond {
		for _, buddy := range buddiesAlready {
			if buddy.LastSeen != nil {
				buddy.LastSeen.Set(DtAtZeroNanosUtc)
			}
		}
	}
	buddiesAlready = sl.SortedPer(buddiesAlready, func(lhs *User, rhs *User) int {
		if lhs.Offline && !rhs.Offline {
			return math.MaxInt
		} else if rhs.Offline && !lhs.Offline {
			return math.MinInt
		}
		lhs_last_seen, rhs_last_seen := *lhs.LastSeen, *rhs.LastSeen
		lhs_last_seen.Set(DtAtZeroSecsUtc)
		rhs_last_seen.Set(DtAtZeroSecsUtc)
		return int(rhs_last_seen.Time().UnixNano() - lhs_last_seen.Time().UnixNano())
	})

	for _, buddy_request := range buddyRequestsMade {
		buddy_request.Auth.SetId(0)
		buddy_request.Btw = ""
		buddy_request.Buddies = nil
		buddy_request.DtMod = buddy_request.DtMade
		buddy_request.LastSeen = buddy_request.DtMade
		buddy_request.PicFileId = ""
	}
	buddyRequestsMade = sl.SortedPer(buddyRequestsMade, func(lhs *User, rhs *User) int { return int(rhs.Id - lhs.Id) })
	return
}

func userAddBuddy(ctx *Ctx, forUser *User, nickOrEmailAddr string) *User {
	buddy_to_be := yodb.FindOne[User](ctx, UserNick.Equal(nickOrEmailAddr).Or(UserAuth_EmailAddr.Equal(nickOrEmailAddr)))
	if (buddy_to_be != nil) && !sl.Has(forUser.Buddies, buddy_to_be.Id) {
		userUpdate(ctx, &User{Id: forUser.Id, Buddies: sl.With(forUser.Buddies, buddy_to_be.Id)}, false, UserBuddies)
	}
	return buddy_to_be
}
