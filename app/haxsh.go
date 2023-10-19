package haxsh

import yodb "yo/db"

var devModeInitMockUsers func()

func Init() {
	yodb.Ensure[User, UserField]("", nil,
		yodb.Unique[UserField]{UserAuth, UserNick},
		nil)
	yodb.Ensure[Post, PostField]("", nil,
		nil,
		yodb.Index[PostField]{PostBy, PostTo})
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
