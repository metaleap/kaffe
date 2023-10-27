package haxsh

import (
	"os"

	. "yo/cfg"
	yodb "yo/db"
	. "yo/srv"
)

var devModeInitMockUsers func()

func init() {
	AppApiUrlPrefix = "_/"
	AppSideStaticRePathFor = func(requestPath string) string {
		return "__static/haxsh.html"
	}
	StaticFileFilters["picRounded"] = imageRoundedSvgOfPng
	for dir_name, dir_path := range Cfg.STATIC_FILE_STORAGE_DIRS {
		StaticFileDirs[dir_name] = os.DirFS(dir_path)
	}
}

func Init() {
	yodb.Ensure[User, UserField]("", nil, false,
		yodb.ReadOnly[UserField]{UserAuth},
		yodb.Unique[UserField]{UserAuth, UserNick},
		yodb.NoUpdTrigger[UserField]{UserLastSeen},
	)
	yodb.Ensure[Post, PostField]("", nil, false,
		yodb.ReadOnly[PostField]{PostBy},
		yodb.Index[PostField]{PostBy, PostTo},
	)
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
