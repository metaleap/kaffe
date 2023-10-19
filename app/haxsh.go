package haxsh

import yodb "yo/db"

var devModeInitMockUsers func()

func Init() {
	yodb.Ensure[User, UserField]("", nil, []UserField{UserFieldAuth})
	yodb.Ensure[Post, PostField]("", nil, nil)
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
