package kaffe

import (
	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	yomail "yo/mail"
	"yo/util/str"
)

const mailTmplVarHref = "href"
const mailTmplVarReqTime = "req_time"

var mailTmpl = str.Trim(`
Hi {` + yoauth.MailTmplVarName + `},

you (or someone trolling you) requested that you {action} at ` + Cfg.YO_APP_DOMAIN + `.

Just delete this if you did not request this (at around {` + mailTmplVarReqTime + `} UTC).

Else, go to {` + mailTmplVarHref + `} and enter your email address plus the following 2 passwords:

First, this below one-time code, best via copy-and-paste:

{` + yoauth.MailTmplVarTmpPwd + `}

Secondly, your own chosen password for future logins (min. length ` + str.FromInt(Cfg.YO_AUTH_PWD_MIN_LEN) + ` characters).


Rock on!

`)

func init() {
	yomail.Templates[yoauth.MailTmplIdSignUp] = &yomail.Templ{
		Subject: "Your sign-up request at " + Cfg.YO_APP_DOMAIN,
		Body:    str.Repl(mailTmpl, str.Dict{"action": "sign up"}),
	}
	yomail.Templates[yoauth.MailTmplIdPwdForgot] = &yomail.Templ{
		Subject: "Your reset request at " + Cfg.YO_APP_DOMAIN,
		Body:    str.Repl(mailTmpl, str.Dict{"action": "reset your password"}),
	}

	yoauth.AppSideTmplPopulate = func(ctx *Ctx, reqTime *yodb.DateTime, emailAddr yodb.Text, existingMaybe *yoauth.UserAuth, tmplArgsToPopulate yodb.JsonMap[string]) {
		var user *User
		if existingMaybe != nil {
			user = yodb.FindOne[User](ctx, UserAuth.Equal(existingMaybe.Id))
		}
		if user != nil {
			tmplArgsToPopulate[yoauth.MailTmplVarName] = string(user.Nick)
		}
		tmplArgsToPopulate[mailTmplVarHref] = "https://" + Cfg.YO_APP_DOMAIN + "?needPwd"
		tmplArgsToPopulate[mailTmplVarReqTime] = reqTime.Time().Format("15:04:05")
	}
}
