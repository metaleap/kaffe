package haxsh

import yodb "yo/db"

var devModeInitMockUsers func()

func Init() {
	yodb.Ensure[User, UserField]("", nil,
		yodb.ReadOnly[UserField]{UserAuth},
		yodb.Unique[UserField]{UserAuth, UserNick},
	)
	yodb.Ensure[Post, PostField]("", nil,
		yodb.ReadOnly[PostField]{PostBy, PostRepl},
		yodb.Index[PostField]{PostBy, PostTo},
	)
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
