package haxsh

import (
	"os"
	"time"
	"yo"

	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	yojobs "yo/jobs"
	. "yo/srv"
	. "yo/util"
	"yo/util/str"
)

const appDomain = "sesh.cafe"
const appHref = "https://" + appDomain

var devModeInitMockUsers func()

func init() {
	yo.AppPkgPath = haxshPkg.PkgPath()
	AppApiUrlPrefix = "_/"
	AppSideStaticRePathFor = func(requestPath string) string {
		return "__static/haxsh.html"
	}
	StaticFileFilters["picRounded"] = imageRoundedSvgOfPng
	for dir_name, dir_path := range Cfg.STATIC_FILE_STORAGE_DIRS {
		StaticFileDirs[dir_name] = os.DirFS(dir_path)
	}
	OnBeforeServingStaticFile = func(ctx *Ctx) {
		var is_anon *bool
		for static_dir_name := range StaticFileDirs {
			if str.Begins(ctx.Http.UrlPath, static_dir_name) {
				if is_anon == nil {
					is_anon = ToPtr(yoauth.IsNotCurrentlyLoggedIn(ctx))
				}
				if *is_anon {
					panic(ErrUnauthorized)
				}
			}
		}
	}
}

func Init() {
	yodb.Ensure[User, UserField]("", nil, false,
		yodb.ReadOnly[UserField]{UserAuth},
		yodb.Index[UserField]{UserLastSeen},
		yodb.Unique[UserField]{UserAuth, UserNick},
		yodb.NoUpdTrigger[UserField]{UserLastSeen, userByBuddyDtLastMsgCheck},
	)
	yodb.Ensure[Post, PostField]("", nil, false,
		yodb.ReadOnly[PostField]{PostBy},
		yodb.Index[PostField]{PostBy, PostTo},
	)
	yodb.Ensure[fileDelReq, fileDelReqField]("", nil, false)
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}

	// ensure app-defined job-defs before starting jobs engine
	{
		ctx := NewCtxNonHttp(yojobs.Timeout1Min, false, "")
		defer ctx.OnDone(nil)

		yodb.Upsert[yojobs.JobDef](ctx, &yojobs.ExampleJobDef)
		yodb.Upsert[yojobs.JobDef](ctx, &cleanUpJobDef)
	}

	// ensure configured vip users are so in db
	user_email_addrs_vip := CfgGet[[]string]("VIP_USER_EMAIL_ADDRS")
	for _, email_addr := range user_email_addrs_vip {
		ctx := NewCtxNonHttp(time.Minute, false, "")
		defer ctx.OnDone(nil)
		if user := userByEmailAddr(ctx, email_addr); (user != nil) && (!user.vip) {
			user.vip = true
			yodb.Update[User](ctx, user, nil, false, UserFields(userVip)...)
		}
	}
}
