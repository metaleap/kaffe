package haxsh

import (
	. "yo/ctx"
	yodb "yo/db"
	. "yo/srv"
	. "yo/util"
	"yo/util/str"
)

var blaPostNew ApiMethod

func init() {
	blaPostNew = Api(apiPostNew, PkgInfo)
	Apis(ApiMethods{
		"postNew": blaPostNew,
	})
}

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

func apiPostNew(this *ApiCtx[Post, Void]) {
	postNew(this.Ctx, this.Args)
}

func postNew(ctx *Ctx, post *Post) {
	_ = yodb.CreateOne(ctx, post)
}

func PostNew(ctx *Ctx, user *User, md string, inReplyTo yodb.I64, files []FileRef, to []yodb.I64) {
	post := &Post{Md: yodb.Text(md).But(str.Trim), To: to, Files: files}
	post.By.SetId(user.Id)
	post.Repl.SetId(inReplyTo)
	postNew(ctx, post)
}
