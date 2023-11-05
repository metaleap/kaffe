package haxsh

import (
	. "yo/cfg"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	yojobs "yo/jobs"
	yomail "yo/mail"
	"yo/util/str"
)

const mailTmplVarHref = "href"
const mailTmplVarReqTime = "req_time"

var mailTmpl = str.Trim(`
Hi {` + yoauth.MailTmplVarName + `},

you (or someone trolling you) requested that you {action} at ` + appDomain + `.

Just delete this if you did not request this (at around {` + mailTmplVarReqTime + `} UTC).

Else, go to {` + mailTmplVarHref + `} and enter your email address plus the following 2 passwords:

First, this below auto-generated one-time code below, best via copy-and-paste:

{` + yoauth.MailTmplVarTmpPwd + `}

Secondly, your own chosen password for future logins (no shorter than ` + str.FromInt(Cfg.YO_AUTH_PWD_MIN_LEN) + ` characters).


Rock on!

`)

func init() {
	yomail.Templates[yoauth.MailTmplIdSignUp] = &yomail.Templ{
		Vars:    []string{mailTmplVarHref, mailTmplVarReqTime, yoauth.MailTmplVarTmpPwd, yoauth.MailTmplVarName},
		Subject: "Your sign-up request at " + appDomain,
		Body:    str.Repl(mailTmpl, str.Dict{"action": "sign up"}),
	}
	yomail.Templates[yoauth.MailTmplIdPwdForgot] = &yomail.Templ{
		Vars:    []string{mailTmplVarHref, mailTmplVarReqTime, yoauth.MailTmplVarTmpPwd, yoauth.MailTmplVarName},
		Subject: "Your reset request at " + appDomain,
		Body:    str.Repl(mailTmpl, str.Dict{"action": "reset your password"}),
	}

	yoauth.AppSideTmplPopulate = func(ctx *yojobs.Context, reqTime *yodb.DateTime, emailAddr yodb.Text, existingMaybe *yoauth.UserAuth, tmplArgsToPopulate yodb.JsonMap[string]) {
		var user *User
		if existingMaybe != nil {
			user = yodb.FindOne[User](ctx.Ctx, UserAuth.Equal(existingMaybe.Id))
		}
		if user != nil {
			tmplArgsToPopulate[yoauth.MailTmplVarName] = string(user.Nick)
		}
		tmplArgsToPopulate[mailTmplVarHref] = appHref + "?needPwd"
		tmplArgsToPopulate[mailTmplVarReqTime] = reqTime.Time().Format("15:04:05")
	}
}
