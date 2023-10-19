package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	. "yo/srv"
	. "yo/util"
)

func init() {
	Apis(ApiMethods{
		"postNew": apiPostNew.
			Checks(Fails{"ExpectedNonEmptyPost", PostMd.Equal("").And(PostFiles.Len().Equal(0))}).
			CouldFailWith(ErrDbNotStored, "RepliedToPostDoesNotExist", "ExpectedOnlyBuddyRecipients"),
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

type FileRef struct {
	Id   string
	Name string
}

var apiPostNew = api(func(this *ApiCtx[Post, Return[yodb.I64]]) {
	this.Ret.Result = postNew(this.Ctx, this.Args, true)
})

func postNew(ctx *Ctx, post *Post, byCurUserInCtx bool) (ret yodb.I64) {
	if byCurUserInCtx {
		user := UserCur(ctx)
		post.by.SetId(user.Id)
	}

	if ret = yodb.CreateOne(ctx, post); ret <= 0 {
		panic(ErrDbNotStored)
	}
	return
}
