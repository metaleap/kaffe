package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	"yo/util/sl"
)

func userBuddies(ctx *Ctx, forUser *User) (buddiesAlready []*User, buddyRequestsMade []*User, buddyRequestsBy []*User) {
	query := q.InArr(forUser.Id, UserBuddies)
	if len(forUser.Buddies) > 0 {
		query = query.Or(UserId.In(forUser.Buddies.ToAnys()...))
	}
	for _, user := range yodb.FindMany[User](ctx, query, 0, nil, UserLastSeen.Desc(), UserDtMod.Desc()) {
		in_our_buddies, in_their_buddies := sl.Has(forUser.Buddies, user.Id), sl.Has(user.Buddies, forUser.Id)
		if in_our_buddies && in_their_buddies {
			buddiesAlready = append(buddiesAlready, user.augmentAfterLoaded())
		} else if in_their_buddies {
			buddyRequestsBy = append(buddyRequestsBy, user)
		} else if in_our_buddies {
			buddyRequestsMade = append(buddyRequestsMade, user)
		}
	}

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
	user := yodb.FindOne[User](ctx, UserNick.Equal(nickOrEmailAddr).Or(UserAuth_EmailAddr.Equal(nickOrEmailAddr)))
	if (user != nil) && !sl.Has(forUser.Buddies, user.Id) {
		userUpdate(ctx, &User{Id: forUser.Id, Buddies: sl.With(forUser.Buddies, user.Id)}, true, false, UserBuddies)
	}
	return user
}
