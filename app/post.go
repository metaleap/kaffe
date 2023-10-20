package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
)

func init() {
	Apis(ApiMethods{
		"postNew": apiPostNew.Checks(
			Fails{Err: "ExpectedNonEmptyPost", If: PostMd.Equal("").And(PostFiles.Len().Equal(0))},
			Fails{Err: "RepliedToPostDoesNotExist", If: PostRepl.LessThan(0)},
			Fails{Err: "InvalidItemInFiles", If: PostFiles.ArrAny(q.Equal, "").Equal(true)},
			Fails{Err: "ExpectedOnlyBuddyRecipients", If: PostTo.ArrAny(q.LessOrEqual, 0).Equal(true)},
		).
			FailIf(ErrUnauthorized, yoauth.CurrentlyNotLoggedIn).
			CouldFailWith(ErrDbNotStored),
	})
}

type Post struct {
	Id      yodb.I64
	Created *yodb.DateTime

	by    yodb.Ref[User, yodb.RefOnDelCascade]
	To    yodb.Arr[yodb.I64]
	Md    yodb.Text
	Files yodb.Arr[FileRef]
	Repl  yodb.Ref[Post, yodb.RefOnDelCascade]
}

type FileRef string

var apiPostNew = api(func(this *ApiCtx[Post, Return[yodb.I64]]) {
	this.Ret.Result = postNew(this.Ctx, this.Args, true)
})

func postNew(ctx *Ctx, post *Post, byCurUserInCtx bool) (ret yodb.I64) {
	ctx.DbTx()

	var user *User
	post_by_user_id := post.by.Id()
	if byCurUserInCtx || (post_by_user_id <= 0) {
		user = userCur(ctx)
		post_by_user_id = user.Id
		post.by.SetId(user.Id)
	}
	if post_by_user_id <= 0 {
		panic(ErrUnauthorized)
	}
	if user == nil {
		if user = userById(ctx, post_by_user_id); user == nil {
			panic(ErrUnauthorized)
		}
	}

	if in_reply_to := post.Repl.Id(); (in_reply_to > 0) && !yodb.Exists[Post](ctx, PostId.Equal(in_reply_to)) {
		panic(ErrPostNew_RepliedToPostDoesNotExist)
	}

	if len(post.To) > 0 {
		if sl.Any(post.To, func(it yodb.I64) bool { return !sl.Has(user.Buddies, it) }) {
			panic(ErrPostNew_ExpectedOnlyBuddyRecipients)
		}
	}

	if ret = yodb.CreateOne(ctx, post); ret <= 0 {
		panic(ErrDbNotStored)
	}
	return
}
