package haxsh

import yodb "yo/db"

var devModeInitMockUsers func()

func Init() {
	yodb.Ensure[User, UserField]("", nil)
	yodb.Ensure[Post, PostField]("", nil)
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
