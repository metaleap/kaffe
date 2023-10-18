package haxsh

import yodb "yo/db"

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
