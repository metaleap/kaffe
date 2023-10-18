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
	Files yodb.Arr[FileRef]
	Repl  yodb.Ref[Post, yodb.RefOnDelCascade]
}

type FileRef struct {
	Id   string
	Name string
}

func UserPost(ctx *Ctx, user *User, md string, inReplyTo yodb.I64, files []FileRef, to []yodb.I64) {
	post := &Post{Md: yodb.Text(md), To: to, Files: files}
	post.By.SetId(user.Id)
	post.Repl.SetId(inReplyTo)
	_ = yodb.CreateOne(ctx, post)
}
