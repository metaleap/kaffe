package kaffe

import (
	"io"
	"math"
	"math/rand"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

const apiMethodNameUserUpdate = "userUpdate"
const apiMethodNameUserBuddiesAdd = "userBuddiesAdd"

func init() {
	Apis(ApiMethods{
		"userSignOut": apiUserSignOut.
			CouldFailWith(":" + yoauth.MethodPathLogout),

		"userSignUpOrForgotPassword": apiUserSignUpOrForgotPassword.
			CouldFailWith(":"+yoauth.MethodPathRegister).
			Checks(
				Fails{Err: "EmailRequiredButMissing", If: UserSignUpOrForgotPasswordNickOrEmailAddr.Equal("")},
				Fails{Err: "EmailInvalid", If: yoauth.IsEmailishEnough(UserSignUpOrForgotPasswordNickOrEmailAddr).Not()},
			),

		"userSignInOrReset": apiUserSignInOrReset.
			CouldFailWith(":"+yoauth.MethodPathLoginOrFinalizePwdReset).
			Checks(
				Fails{Err: "ExpectedPasswordAndNickOrEmailAddr", If: UserSignInOrResetNickOrEmailAddr.Equal("").Or(UserSignInOrResetPasswordPlain.Equal(""))},
				Fails{Err: "WrongPassword",
					If: UserSignInOrResetPasswordPlain.StrLen().LessThan(Cfg.YO_AUTH_PWD_MIN_LEN).Or(
						UserSignInOrResetPasswordPlain.StrLen().GreaterThan(Cfg.YO_AUTH_PWD_MAX_LEN)).Or(
						UserSignInOrResetPassword2Plain.StrLen().GreaterThan(0).And(
							UserSignInOrResetPassword2Plain.StrLen().LessThan(Cfg.YO_AUTH_PWD_MIN_LEN).Or(
								UserSignInOrResetPassword2Plain.StrLen().GreaterThan(Cfg.YO_AUTH_PWD_MAX_LEN)))),
				},
			),

		"userBy": apiUserBy.Checks(
			Fails{Err: "ExpectedEitherNickNameOrEmailAddr", If: UserByEmailAddr.Equal("").And(UserByNickName.Equal(""))},
		).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		apiMethodNameUserUpdate: apiUserUpdate.IsMultipartForm().
			CouldFailWith(":"+yodb.ErrSetDbUpdate, "NicknameAlreadyExists", "ExpectedNonEmptyNickname").
			Checks(
				Fails{Err: ErrDbUpdExpectedIdGt0, If: UserUpdateId.LessOrEqual(0)},
			).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"userBuddies": apiUserBuddies.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		apiMethodNameUserBuddiesAdd: apiUserBuddiesAdd.
			CouldFailWith(":"+yodb.ErrSetDbUpdate).
			Checks(
				Fails{Err: "ExpectedEitherNickNameOrEmailAddr", If: UserBuddiesAddNickOrEmailAddr.Equal("")},
			).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"postsRecent": apiPostsRecent.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"postsForMonthUtc": apiPostsForMonthUtc.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"postMonthsUtc": apiPostMonthsUtc.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"postsDeleted": apiPostsDeleted.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"postNew": apiPostNew.IsMultipartForm().
			CouldFailWith("ExpectedNonEmptyPost").
			Checks(
				Fails{Err: "ExpectedOnlyBuddyRecipients", If: q.ArrAreAnyIn(PostTo, q.OpLeq, 0)},
				Fails{Err: "ExpectedEmptyFilesFieldWithUploadedFilesInMultipartForm", If: PostFiles.ArrLen().NotEqual(0)},
			).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"postDelete": apiPostDelete.Checks(
			Fails{Err: "InvalidPostId", If: PostDeleteId.LessOrEqual(0)},
		).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"postEmojiFullList": apiPostEmojiFullList,
	})
}

var apiUserSignOut = api(func(this *ApiCtx[Void, Void]) {
	Do(yoauth.ApiUserLogout, this.Ctx, this.Args)
})

var apiUserSignInOrReset = api(func(this *ApiCtx[ApiUserSignInOrReset, Void]) {
	this.Ctx.DbTx()
	this.Args.ensureEmailAddr(this.Ctx, Err___yo_authLoginOrFinalizePwdReset_AccountDoesNotExist, Err___yo_authLoginOrFinalizePwdReset_EmailInvalid)
	user_auth := Do(yoauth.ApiUserLoginOrFinalizePwdReset, this.Ctx, &yoauth.ApiAccountPayload{EmailAddr: this.Args.NickOrEmailAddr, PasswordPlain: this.Args.PasswordPlain, Password2Plain: this.Args.Password2Plain})
	user := userCur(this.Ctx)
	if user == nil { // this was a new-user-sign-up rather than an existing-user-pwd-reset
		user_nick := user_auth.EmailAddr[:str.Idx(user_auth.EmailAddr.String(), '@')]
		user = &User{LastSeen: yodb.DtNow(), Nick: user_nick}
		for n := 1; yodb.Exists[User](this.Ctx, UserNick.Equal(user.Nick)); n++ {
			user.Nick = user_nick + yodb.Text(str.FromInt(n))
		}
		user.Nick = user_nick
		user.Auth.SetId(user_auth.Id)
		user.LastSeen = yodb.DtNow()
		_ = yodb.CreateOne[User](this.Ctx, user)
	}
})

var apiUserSignUpOrForgotPassword = api(func(this *ApiCtx[ApiNickOrEmailAddr, Void]) {
	this.Ctx.DbTx()
	this.Args.ensureEmailAddr(this.Ctx, Err___yo_authLoginOrFinalizePwdReset_AccountDoesNotExist, ErrUserSignUpOrForgotPassword_EmailInvalid)
	yoauth.UserPregisterOrForgotPassword(this.Ctx, this.Args.NickOrEmailAddr)
})

var apiUserBy = api(func(this *ApiCtx[struct {
	EmailAddr string
	NickName  string
}, User]) {
	if this.Args.NickName != "" {
		this.Ret = userByNickName(this.Ctx, this.Args.NickName)
	} else if this.Args.EmailAddr != "" {
		this.Ret = userByEmailAddr(this.Ctx, this.Args.EmailAddr)
	} else {
		panic(ErrUserBy_ExpectedEitherNickNameOrEmailAddr)
	}
})

var apiUserUpdate = api(func(this *ApiCtx[yodb.ApiUpdateArgs[User, UserField], Void]) {
	user_cur := userCur(this.Ctx)
	if user_cur == nil {
		panic(ErrUnauthorized)
	}
	this.Args.Changes.Id = user_cur.Id
	this.Args.Changes.Auth.SetId(user_cur.Auth.Id())

	uploaded_file_names, _ := apiHandleUploadedFiles(this.Ctx, "picfile", 1, imageSquared)
	for _, file_name := range uploaded_file_names {
		if (user_cur.PicFileId != "") && (user_cur.PicFileId != this.Args.Changes.PicFileId) {
			yodb.CreateOne[fileDelReq](this.Ctx, &fileDelReq{FileNames: yodb.Arr[yodb.Text]{user_cur.PicFileId}})
		}
		this.Args.Changes.PicFileId = file_name
		if len(this.Args.ChangedFields) > 0 {
			this.Args.ChangedFields = sl.With(this.Args.ChangedFields, UserPicFileId)
		}
		break
	}
	userUpdate(this.Ctx, &this.Args.Changes, (len(this.Args.ChangedFields) > 0), this.Args.ChangedFields...)
})

var apiUserBuddies = api(func(this *ApiCtx[Void, struct {
	Buddies         []*User
	BuddyRequestsBy []*User
}]) {
	var buddy_requests_made []*User
	this.Ret.Buddies, buddy_requests_made, this.Ret.BuddyRequestsBy = userBuddies(this.Ctx, userCur(this.Ctx), true)
	this.Ret.Buddies = append(this.Ret.Buddies, buddy_requests_made...)
})

var apiUserBuddiesAdd = api(func(this *ApiCtx[struct {
	NickOrEmailAddr string
}, struct {
	Done bool
}]) {
	this.Ret.Done = (nil != userAddBuddy(this.Ctx, userCur(this.Ctx), this.Args.NickOrEmailAddr))
})

type ApiArgPeriod struct {
	Period YearAndMonth
	OnlyBy []yodb.I64
}

var apiPostsForMonthUtc = api(func(this *ApiCtx[ApiArgPeriod, PostsListResult]) {
	this.Ret.Posts = postsForMonthUtc(this.Ctx, userCur(this.Ctx), this.Args.Period, this.Args.OnlyBy)
	this.Ret.augmentWithFileContentTypes()
	this.Ret.NextSince = nil
})

var apiPostsRecent = api(func(this *ApiCtx[struct {
	Since  *yodb.DateTime
	OnlyBy []yodb.I64
}, PostsListResult]) {
	user_cur := userCur(this.Ctx)
	if user_cur == nil {
		panic(ErrUnauthorized)
	}
	this.Ret = postsRecent(this.Ctx, user_cur, this.Args.Since, this.Args.OnlyBy)
	this.Ret.augmentWithFileContentTypes()
})

var apiPostMonthsUtc = api(func(this *ApiCtx[struct {
	WithUserIds []yodb.I64
}, struct {
	Periods []YearAndMonth
}]) {
	this.Ret.Periods = postMonthsUtc(this.Ctx, userCur(this.Ctx), this.Args.WithUserIds)
})

var apiPostsDeleted = api(func(this *ApiCtx[struct {
	OutOfPostIds []yodb.I64
}, struct {
	DeletedPostIds []yodb.I64
}]) {
	this.Ret.DeletedPostIds = postsDeleted(this.Ctx, this.Args.OutOfPostIds)
})

var apiPostNew = api(func(this *ApiCtx[Post, Return[yodb.I64]]) {
	this.Args.Files, _ = apiHandleUploadedFiles(this.Ctx, "files", 0, nil)

	{
		uris, toks := str.Dict{}, str.Split(string(this.Args.Htm), " ")
		for _, tok := range toks {
			const needle = "data:"
			for _, quot := range []string{"\"", "'"} {
				if idx1 := str.IdxSub(tok, quot+needle); idx1 >= 0 {
					if idx2 := str.IdxSub(tok[idx1+1:], quot) + idx1 + 1; idx2 > idx1 {
						uris[tok] = tok[:idx1+len(quot)] + "data:noThxPlzKbai" + tok[idx2:]
					}
				}
			}
			if !str.Has(tok, "://") {
				continue
			}
			if uri, err := url.Parse(tok); err == nil {
				uri_str := uri.String()
				uris[" "+uri_str+" "] = " <a target='_blank' href='" + uri_str + "'>" + uri_str + "</a> "
			}
		}
		this.Args.Htm = yodb.Text(str.Trim(str.Replace(string(" "+this.Args.Htm+" "), uris)))
		if idx1 := str.Idx(string(this.Args.Htm), ':'); idx1 >= 0 {
			if idx2 := str.IdxLast(string(this.Args.Htm), ':'); idx2 > idx1 {
				this.Args.Htm = yodb.Text(str.Replace(string(this.Args.Htm), emojiKnown))
			}
		}
	}

	this.Ret.Result = postNew(this.Ctx, this.Args, 0)
})

var apiPostDelete = api(func(this *ApiCtx[struct {
	Id yodb.I64
}, Void]) {
	post := yodb.ById[Post](this.Ctx, this.Args.Id)
	if post == nil {
		return
	}
	user_cur := userCur(this.Ctx)
	if (user_cur == nil) || (post.By.Id() != user_cur.Id) {
		panic(ErrUnauthorized)
	}
	_ = postDelete(this.Ctx, post)
})

var apiPostEmojiFullList = api(func(this *ApiCtx[Void, Return[map[string]string]]) {
	this.Ret.Result = make(map[string]string, len(emojiKnown))
	for emoji_code := range emojiKnown {
		this.Ret.Result[emoji_code] = emojiUnescaped(emoji_code)
	}
})

func (me *PostsListResult) augmentWithFileContentTypes() {
	for _, post := range me.Posts {
		post.FileContentTypes = make([]string, len(post.Files))
		for i, file_id := range post.Files {
			post.FileContentTypes[i] = mime.TypeByExtension(filepath.Ext(string(file_id)))
		}
	}
}

func userUploadedFilePath(fileId string) string {
	return filepath.Join(Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"], fileId)
}

func apiHandleUploadedFiles(ctx *Ctx, fieldName string, maxNumFiles int, transform func([]byte) []byte) (fileNames []yodb.Text, filePaths []string) {
	if files := ctx.Http.Req.MultipartForm.File[fieldName]; len(files) > 0 {
		for i, file := range files {
			if (maxNumFiles > 0) && (i == maxNumFiles) {
				break
			}
			multipart_file, err := file.Open()
			if multipart_file != nil {
				defer multipart_file.Close()
			}
			if err != nil {
				panic(err)
			}
			data, err := io.ReadAll(multipart_file)
			if err != nil {
				panic(err)
			}
			if transform != nil {
				data = transform(data)
			}
			dst_file_name := str.FromI64(rand.Int63n(math.MaxInt64), 36) + "_" + str.FromI64(time.Now().UnixNano(), 36) + "_" + str.FromI64(file.Size, 36) + "__yo__" + file.Filename
			dst_file_path := userUploadedFilePath(dst_file_name)
			FileWrite(dst_file_path, data)
			fileNames, filePaths = append(fileNames, yodb.Text(dst_file_name)), append(filePaths, dst_file_path)
		}
	}
	return
}

type ApiUserSignInOrReset struct {
	ApiNickOrEmailAddr
	PasswordPlain  string
	Password2Plain string
}

type ApiNickOrEmailAddr struct {
	NickOrEmailAddr string
}

func (me *ApiNickOrEmailAddr) ensureEmailAddr(ctx *Ctx, errNoSuchAccount Err, errBadEmail Err) {
	if str.Has(me.NickOrEmailAddr, "@") {
		if !str.IsEmailishEnough(me.NickOrEmailAddr) {
			panic(errBadEmail)
		}
	} else {
		existing_user := yodb.FindOne[User](ctx, UserNick.Equal(me.NickOrEmailAddr))
		if existing_user == nil {
			panic(errNoSuchAccount)
		}
		me.NickOrEmailAddr = string(yoauth.ById(ctx, existing_user.Auth.Id()).EmailAddr)
	}
}
