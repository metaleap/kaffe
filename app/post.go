package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	. "yo/util"
)

type Post struct {
	Id      yodb.I64
	Created *yodb.DateTime

	By    yodb.Ref[User, yodb.RefOnDelCascade]
	To    yodb.Arr[yodb.I64]
	Md    yodb.Text
	Files yodb.Arr[struct {
		Id   string
		Name string
	}]
	Repl yodb.Ref[Post, yodb.RefOnDelCascade]
}

func UserPost(ctx *Ctx, user *User, text string, inReplyTo yodb.I64, files []Pair[string, string], to []yodb.I64) {
	post := &Post{To: nil, Md: yodb.Text(mockGetFortune(0, false)), Files: nil}
	post.By.SetId(user.Id)
	_ = yodb.CreateOne(ctx, post)
}
