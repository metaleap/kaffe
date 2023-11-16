package kaffe

import (
	"time"
	"yo"

	yoauth "yo/auth"
	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	yojobs "yo/jobs"
	. "yo/srv"
	. "yo/util"
	"yo/util/str"
)

var devModeInitMockUsers func()

func init() {
	yoauth.AutoLoginAfterSuccessfullyFinalizedSignUpOrPwdResetReq = true
	yoauth.EnforceGenericizedErrors = false
	yo.AppPkgPath = kaffePkg.PkgPath()
	AppSideStaticRePathFor = func(reqUrlPath string) string {
		return If(str.Begins(reqUrlPath, "_/"), "", "__static/kaffe.html")
	}
	StaticFileFilters["picRounded"] = imageRoundedSvgOfImage
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
		yodb.ReadOnly[UserField]{UserAccount},
		yodb.Index[UserField]{UserLastSeen},
		yodb.Unique[UserField]{UserAccount, UserNick},
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

	{ // ensure app-defined job-defs before starting jobs engine
		ctx := NewCtxNonHttp(yojobs.Timeout1Min, false, "")
		yodb.Upsert[yojobs.JobDef](ctx, &cleanUpJobDef)
		yodb.Upsert[yojobs.JobDef](ctx, &gravatarJobDef)
		ctx.OnDone(nil)
	}

	elizaEnsureUser()

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
