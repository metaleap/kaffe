package haxsh

import (
	"os"

	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	yojobs "yo/jobs"
	. "yo/srv"
)

var devModeInitMockUsers func()
var jobs = yojobs.NewEngine(yojobs.Options{})

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
		yodb.NoUpdTrigger[UserField]{UserLastSeen, userByBuddyDtLastMsgCheck},
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

	// ensure app-defined job-defs before starting jobs engine
	ctx := NewCtxNonHttp(yojobs.TimeoutLong, false, "")
	defer ctx.OnDone(nil)
	yodb.Upsert[yojobs.JobDef](ctx, &yojobs.JobDef{
		Name:                             "exampleJob",
		JobTypeId:                        "yojobs.ExampleJobType",
		MaxTaskRetries:                   1,
		DeleteAfterDays:                  1,
		TimeoutSecsTaskRun:               2,
		TimeoutSecsJobRunPrepAndFinalize: 3,
		Schedules:                        yodb.Arr[yodb.Text]{"* * * * *"},
	})
	go jobs.Resume()
}
