package haxsh

import yodb "yo/db"

var devModeInitMockUsers func()

func Init() {
	yodb.Ensure[User, UserField]("", nil,
		yodb.Unique[UserField]{UserAuth, UserNick}, yodb.ReadOnly[UserField]{UserAuth})
	yodb.Ensure[Post, PostField]("", nil,
		yodb.Index[PostField]{postBy, PostTo}, yodb.ReadOnly[PostField]{postBy, PostRepl})
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
