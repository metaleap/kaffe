package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	. "yo/srv"
	. "yo/util"
)

func init() {
	Apis(ApiMethods{
		"postNew": apiPostNew,
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

var apiPostNew = api(func(this *ApiCtx[Post, Void]) {
	postNew(this.Ctx, this.Args, true)
})

func postNew(ctx *Ctx, post *Post, byCurUserInCtx bool) {
	if byCurUserInCtx {
		user := UserCur(ctx)
		post.by.SetId(user.Id)
	}
	_ = yodb.CreateOne(ctx, post)
}
