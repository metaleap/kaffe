package haxsh

import (
	yodb "yo/db"
	yosrv "yo/srv"
	. "yo/util"
)

var devModeInitMockUsers func()

func init() {
	yosrv.AppApiUrlPrefix = "_/"
	yosrv.AppSideStaticRePathFor = func(requestPath string) string {
		is_home := (requestPath == "") // `requestPath` never has a leading slash
		return If(is_home, "__static/home.html", "__static/haxsh.html")
	}
}

func Init() {
	yodb.Ensure[User, UserField]("", nil, false,
		yodb.ReadOnly[UserField]{UserAuth},
		yodb.Unique[UserField]{UserAuth, UserNick},
		yodb.NoUpdTrigger[UserField]{UserLastSeen},
	)
	yodb.Ensure[Post, PostField]("", nil, false,
		yodb.ReadOnly[PostField]{PostBy, PostRepl},
		yodb.Index[PostField]{PostBy, PostTo},
	)
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
