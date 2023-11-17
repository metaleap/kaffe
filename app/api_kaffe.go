package kaffe

import (
	"encoding/base64"
	"io"
	"math"
	"math/rand"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	yoauth "yo/auth"
	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	"yo/misc/emoji"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

const apiMethodPathUserUpdate = "_/userUpdate"
const apiMethodPathUserBuddiesAdd = "_/userBuddiesAdd"

func init() {
	Apis(ApiMethods{
		"_/userSignOut": apiUserSignOut.
			CouldFailWith(":" + yoauth.MethodPathLogout),

		"_/userSignUpOrForgotPassword": apiUserSignUpOrForgotPassword.
			CouldFailWith(":"+yoauth.MethodPathRegister).
			Checks(
				Fails{Err: "EmailRequiredButMissing", If: __userSignUpOrForgotPasswordNickOrEmailAddr.Equal("")},
				Fails{Err: "EmailInvalid", If: yoauth.IsEmailishEnough(__userSignUpOrForgotPasswordNickOrEmailAddr).Not()},
			),

		"_/userSignInOrReset": apiUserSignInOrReset.
			CouldFailWith(":"+yoauth.MethodPathLoginOrFinalizePwdReset).
			Checks(
				Fails{Err: "ExpectedPasswordAndNickOrEmailAddr", If: __userSignInOrResetNickOrEmailAddr.Equal("").Or(__userSignInOrResetPasswordPlain.Equal(""))},
				Fails{Err: "WrongPassword",
					If: __userSignInOrResetPasswordPlain.StrLen().LessThan(Cfg.YO_AUTH_PWD_MIN_LEN).Or(
						__userSignInOrResetPasswordPlain.StrLen().GreaterThan(Cfg.YO_AUTH_PWD_MAX_LEN)).Or(
						__userSignInOrResetPassword2Plain.StrLen().GreaterThan(0).And(
							__userSignInOrResetPassword2Plain.StrLen().LessThan(Cfg.YO_AUTH_PWD_MIN_LEN).Or(
								__userSignInOrResetPassword2Plain.StrLen().GreaterThan(Cfg.YO_AUTH_PWD_MAX_LEN)))),
				},
			),

		"_/userBy": apiUserBy.Checks(
			Fails{Err: "ExpectedEitherNickNameOrEmailAddr", If: __userByEmailAddr.Equal("").And(__userByNickName.Equal(""))},
		).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		apiMethodPathUserUpdate: apiUserUpdate.IsMultipartForm().
			CouldFailWith(":"+yodb.ErrSetDbUpdate, "NicknameAlreadyExists", "ExpectedNonEmptyNickname").
			Checks(
				Fails{Err: ErrDbUpdExpectedIdGt0, If: __userUpdateId.LessOrEqual(0)},
			).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/userBuddies": apiUserBuddies.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		apiMethodPathUserBuddiesAdd: apiUserBuddiesAdd.
			CouldFailWith(":"+yodb.ErrSetDbUpdate).
			Checks(
				Fails{Err: "ExpectedEitherNickNameOrEmailAddr", If: __userBuddiesAddNickOrEmailAddr.Equal("")},
			).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/postsRecent": apiPostsRecent.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/postsForMonthUtc": apiPostsForMonthUtc.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/postMonthsUtc": apiPostMonthsUtc.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/postsDeleted": apiPostsDeleted.
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/postNew": apiPostNew.IsMultipartForm().
			Checks(
				Fails{Err: "ExpectedNonEmptyPost", If: __postNewNewPost.Equal(nil)},
				Fails{Err: "ExpectedOnlyBuddyRecipients", If: q.Dot(__postNewNewPost, q.ArrAreAny(PostTo, q.OpLeq, 0))},
				Fails{Err: "ExpectedEmptyFilesField", If: q.Dot(__postNewNewPost, PostFiles.ArrLen().NotEqual(0))},
			).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/postDelete": apiPostDelete.Checks(
			Fails{Err: "InvalidPostId", If: __postDeleteId.LessOrEqual(0)},
		).
			FailIf(yoauth.IsNotCurrentlyLoggedIn, ErrUnauthorized),

		"_/postEmojiFullList": apiPostEmojiFullList,
	})
}

type ApiUserSignInOrReset struct {
	ApiNickOrEmailAddr
	PasswordPlain  string
	Password2Plain string
}

type ApiNickOrEmailAddr struct {
	NickOrEmailAddr string
}

type PostNew struct {
	NewPost *Post
}

var apiUserSignOut = api(func(this *ApiCtx[None, None]) {
	Do(yoauth.ApiUserLogout, this.Ctx, this.Args)
})

var apiUserSignInOrReset = api(func(this *ApiCtx[ApiUserSignInOrReset, None]) {
	this.Ctx.DbTx(true)
	this.Args.ensureEmailAddr(this.Ctx, Err___yo_authLoginOrFinalizePwdReset_AccountDoesNotExist, Err___yo_authLoginOrFinalizePwdReset_EmailInvalid)
	user_account := Do(yoauth.ApiUserLoginOrFinalizePwdReset, this.Ctx, &yoauth.ApiAccountPayload{EmailAddr: this.Args.NickOrEmailAddr, PasswordPlain: this.Args.PasswordPlain, Password2Plain: this.Args.Password2Plain})
	user := userCur(this.Ctx)
	if user == nil { // this was a new-user-sign-up rather than an existing-user-pwd-reset
		user_nick := user_account.EmailAddr[:str.Idx(user_account.EmailAddr.String(), '@')]
		user = &User{LastSeen: yodb.DtNow(), Nick: user_nick}
		for n := 1; yodb.Exists[User](this.Ctx, UserNick.Equal(user.Nick)); n++ {
			user.Nick = user_nick + yodb.Text(str.FromInt(n))
		}
		user.Nick = user_nick
		user.Account.SetId(user_account.Id)
		user.LastSeen = yodb.DtNow()
		_ = yodb.CreateOne[User](this.Ctx, user)
	}
})

var apiUserSignUpOrForgotPassword = api(func(this *ApiCtx[ApiNickOrEmailAddr, None]) {
	this.Ctx.DbTx(true)
	this.Args.ensureEmailAddr(this.Ctx, Err___yo_authLoginOrFinalizePwdReset_AccountDoesNotExist, Err__userSignUpOrForgotPassword_EmailInvalid)
	yoauth.UserPregisterOrForgotPassword(this.Ctx, this.Args.NickOrEmailAddr)
})

var apiUserBy = api(func(this *ApiCtx[struct {
	EmailAddr string
	NickName  string
}, Return[*User]]) {
	if this.Args.NickName != "" {
		this.Ret.Result = userByNickName(this.Ctx, this.Args.NickName)
	} else if this.Args.EmailAddr != "" {
		this.Ret.Result = userByEmailAddr(this.Ctx, this.Args.EmailAddr)
	} else {
		panic(Err__userBy_ExpectedEitherNickNameOrEmailAddr)
	}
})

var apiUserUpdate = api(func(this *ApiCtx[yodb.ApiUpdateArgs[User, UserField], None]) {
	user_cur := userCur(this.Ctx)
	if user_cur == nil {
		panic(ErrUnauthorized)
	}
	this.Args.Changes.Id = user_cur.Id
	this.Args.Changes.Account.SetId(user_cur.Account.Id())

	uploaded_file_names, _ := apiHandleUploadedFiles(this.Ctx, "picfile", 1, imageSquared, nil)
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

var apiUserBuddies = api(func(this *ApiCtx[None, struct {
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

var apiPostNew = api(func(this *ApiCtx[PostNew, Return[yodb.I64]]) {
	data_files := map[string][]byte{}
	{ // html processing
		uris, toks := str.Dict{}, str.Split(string(this.Args.NewPost.Htm), " ")
		for _, tok := range toks {
			const needle = "data:"
			for _, quot := range []string{"\"", "'"} {
				if idx1 := str.IdxSub(tok, quot+needle); idx1 >= 0 {
					if idx2 := str.IdxSub(tok[idx1+1:], quot) + idx1 + 1; idx2 > idx1 {
						/*<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAABR8AAAX4CAYAAAA3gL7m...*/
						proto_and_data := tok[idx1+len(quot)+len(needle) : idx2]
						if idx3 := str.IdxSub(proto_and_data, ";"); idx3 < 0 {
							uris[tok] = tok[:idx1+len(quot)] + "data:nope" + tok[idx2:]
						} else {
							data_payload := proto_and_data[idx3+1:]
							data_bytes := []byte(data_payload)
							if needle2 := "base64,"; str.Begins(data_payload, needle2) {
								data_bytes, _ = base64.StdEncoding.DecodeString(data_payload[len(needle2):])
							}
							up_file_name := userUploadedFileNameNew("__yodata__"+str.FromInt(len(data_files)), int64(len(data_bytes)))
							data_files[up_file_name] = data_bytes
							uris[tok] = tok[:idx1+len(quot)] + "/_postfiles/" + up_file_name + tok[idx2:]
						}
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
		this.Args.NewPost.Htm = yodb.Text(str.Trim(str.Replace(string(" "+this.Args.NewPost.Htm+" "), uris)))
		if idx1 := str.Idx(string(this.Args.NewPost.Htm), ':'); idx1 >= 0 {
			if idx2 := str.IdxLast(string(this.Args.NewPost.Htm), ':'); idx2 > idx1 {
				this.Args.NewPost.Htm = yodb.Text(str.Replace(string(this.Args.NewPost.Htm), emoji.GithubLikeAsHtml))
			}
		}
	}

	this.Args.NewPost.Files, _ = apiHandleUploadedFiles(this.Ctx, "files", 0, nil, data_files)
	this.Ret.Result = postNew(this.Ctx, this.Args.NewPost, 0)
})

var apiPostDelete = api(func(this *ApiCtx[struct {
	Id yodb.I64
}, None]) {
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

var apiPostEmojiFullList = api(func(this *ApiCtx[None, Return[map[string]string]]) {
	this.Ret.Result = make(map[string]string, len(emoji.GithubLikeAsHtml))
	for emoji_code := range emoji.GithubLikeAsHtml {
		this.Ret.Result[emoji_code] = emoji.GithubLike(emoji_code)
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

func userUploadedFileNameNew(origFileName string, origFileSize int64) string {
	return str.FromI64(rand.Int63n(math.MaxInt64), 36) + "_" + str.FromI64(time.Now().UnixNano(), 36) + "_" + str.FromI64(origFileSize, 36) + "__yo__" + origFileName
}

func apiHandleUploadedFiles(ctx *Ctx, fieldName string, maxNumFiles int, transform func([]byte) []byte, additionalFiles map[string][]byte) (fileNames []yodb.Text, filePaths []string) {
	handle_file := func(dstFileName string, fileBytes []byte) {
		dst_file_path := userUploadedFilePath(dstFileName)
		FsWrite(dst_file_path, fileBytes)
		fileNames, filePaths = append(fileNames, yodb.Text(dstFileName)), append(filePaths, dst_file_path)
	}
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
			handle_file(userUploadedFileNameNew(file.Filename, file.Size), data)
		}
	}
	for file_name, file_bytes := range additionalFiles {
		if transform != nil {
			file_bytes = transform(file_bytes)
		}
		handle_file(file_name, file_bytes)
	}
	return
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
		me.NickOrEmailAddr = string(yoauth.ById(ctx, existing_user.Account.Id()).EmailAddr)
	}
}
