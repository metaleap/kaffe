package haxsh

import (
	yodb "yo/db"
	yoauth "yo/feat_auth"
	yojobs "yo/jobs"
	yomail "yo/mail"
	"yo/util/str"
)

const mailTmplVarHref = "href"
const mailTmplVarReqTime = "req_time"

var mailTmpl = `
Hi {` + yoauth.MailTmplVarName + `},

you (or someone trolling you) requested that you {action} at {` + mailTmplVarHref + `}.

If you did not request this (at around {` + mailTmplVarReqTime + `} UTC), simply delete this email.

Otherwise, go to {` + mailTmplVarHref + `} and enter the following 2 passwords:

1. Under "old password": this unique one-time password, best via copy-and-paste:

{` + yoauth.MailTmplVarTmpPwd + `}

2. under "new password": your own chosen password for future logins.

Rock on!
`

func init() {
	yomail.Templates[yoauth.MailTmplIdSignUp] = &yomail.Templ{
		Vars:    []string{mailTmplVarHref, mailTmplVarReqTime, yoauth.MailTmplVarTmpPwd, yoauth.MailTmplVarName},
		Subject: "Your sign-up request at " + appDomain,
		Body:    str.Repl(mailTmpl, str.Dict{"action": "sign up"}),
	}
	yomail.Templates[yoauth.MailTmplIdPwdForgot] = &yomail.Templ{
		Vars:    []string{mailTmplVarHref, mailTmplVarReqTime, yoauth.MailTmplVarTmpPwd, yoauth.MailTmplVarName},
		Subject: "Your password-reset request at " + appDomain,
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
		tmplArgsToPopulate[mailTmplVarHref] = "https://sesh.cafe"
		tmplArgsToPopulate[mailTmplVarReqTime] = reqTime.Time().Format("15:04:05")
	}
}
