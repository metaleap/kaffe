package haxsh

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

func init() {
	Apis(ApiMethods{
		"userSignOut": apiUserSignOut.
			CouldFailWith(":" + yoauth.MethodPathLogout),

		"userSignUp": apiUserSignUp.
			CouldFailWith(":"+yoauth.MethodPathRegister, ":userSignIn"),

		"userSignIn": apiUserSignIn.
			CouldFailWith(":" + yoauth.MethodPathLogin),

		"userBy": apiUserBy.Checks(
			Fails{Err: "ExpectedEitherNickNameOrEmailAddr", If: UserByEmailAddr.Equal("").And(UserByNickName.Equal(""))},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),

		"userUpdate": apiUserUpdate.IsMultipartForm().Checks(
			Fails{Err: ErrDbUpdExpectedIdGt0, If: UserUpdateId.LessOrEqual(0)},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized).
			CouldFailWith(":"+yodb.ErrSetDbUpdate, "NicknameAlreadyExists"),

		"userBuddies": apiUserBuddies.
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),

		"postsRecent": apiPostsRecent.
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),

		"postsForPeriod": apiPostsForPeriod.Checks(
			Fails{Err: "ExpectedPeriodGreater0AndLess33Days", If: PostsForPeriodFrom.Equal(nil).Or(PostsForPeriodUntil.Equal(nil)).Or(PostsForPeriodUntil.LessOrEqual(PostsForPeriodFrom))},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),

		"postsDeleted": apiPostsDeleted.
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),

		"postNew": apiPostNew.IsMultipartForm().Checks(
			Fails{Err: "ExpectedOnlyBuddyRecipients", If: q.ArrAreAnyIn(PostTo, q.OpLeq, 0)},
			Fails{Err: "ExpectedEmptyFilesFieldWithUploadedFilesInMultipartForm", If: PostFiles.ArrLen().NotEqual(0)},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized).
			CouldFailWith("ExpectedNonEmptyPost"),

		"postDelete": apiPostDelete.Checks(
			Fails{Err: "InvalidPostId", If: PostDeleteId.LessOrEqual(0)},
		).
			FailIf(yoauth.CurrentlyNotLoggedIn, ErrUnauthorized),
	})
}

var apiUserSignIn = api(func(this *ApiCtx[yoauth.ApiAccountPayload, Void]) {
	Do(yoauth.ApiUserLogin, this.Ctx, this.Args)
})

var apiUserSignUp = api(func(this *ApiCtx[yoauth.ApiAccountPayload, User]) {
	this.Ctx.DbTx()

	auth := Do(yoauth.ApiUserRegister, this.Ctx, this.Args)
	user := User{LastSeen: yodb.DtNow(), byBuddyDtLastMsgCheck: yodb.JsonMap[*yodb.DateTime]{}}
	user.Auth.SetId(auth.Id)
	user.Id = yodb.CreateOne(this.Ctx, &user)
	// _ = Do(apiUserSignIn, this.Ctx, this.Args)
	this.Ret = &user
})

var apiUserSignOut = api(func(this *ApiCtx[Void, Void]) {
	_ = Do(yoauth.ApiUserLogout, this.Ctx, this.Args)
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
	_, user_auth_id := yoauth.CurrentlyLoggedInUser(this.Ctx)
	this.Args.Changes.Id = this.Args.Id
	this.Args.Changes.Auth.SetId(user_auth_id)

	uploaded_file_names, uploaded_file_paths := apiHandleUploadedFiles(this.Ctx, "picfile", 1, imageSquared)
	for i, file_name := range uploaded_file_names {
		old_file_path := filepath.Join(filepath.Dir(uploaded_file_paths[i]), string(userCur(this.Ctx).PicFileId))
		DelFile(old_file_path)
		this.Args.Changes.PicFileId = file_name
		if len(this.Args.ChangedFields) > 0 {
			this.Args.ChangedFields = sl.With(this.Args.ChangedFields, UserPicFileId)
		}
		break
	}
	userUpdate(this.Ctx, &this.Args.Changes, true, (len(this.Args.ChangedFields) > 0), this.Args.ChangedFields...)
})

var apiUserBuddies = api(func(this *ApiCtx[Void, Return[[]*User]]) {
	this.Ret.Result = userBuddies(this.Ctx, userCur(this.Ctx), true)
})

type ApiArgPeriod struct {
	From   *time.Time
	Until  *time.Time
	OnlyBy []yodb.I64
}

var apiPostsForPeriod = api(func(this *ApiCtx[ApiArgPeriod, PostsListResult]) {
	if (this.Args.From == nil) || (this.Args.Until == nil) {
		panic(ErrPostsForPeriod_ExpectedPeriodGreater0AndLess33Days)
	}
	this.Ret.Posts = postsFor(this.Ctx, userCur(this.Ctx), *this.Args.From, *this.Args.Until, this.Args.OnlyBy)
	this.Ret.augmentWithFileContentTypes()
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
				this.Args.Htm = yodb.Text(str.Replace(string(this.Args.Htm), postEmoji))
			}
		}
	}

	this.Ret.Result = postNew(this.Ctx, this.Args, true)
})

var apiPostDelete = api(func(this *ApiCtx[struct {
	Id yodb.I64
}, Void]) {
	_ = postDelete(this.Ctx, this.Args.Id)
})

func (me *PostsListResult) augmentWithFileContentTypes() {
	for _, post := range me.Posts {
		post.FileContentTypes = make([]string, len(post.Files))
		for i, file_id := range post.Files {
			post.FileContentTypes[i] = mime.TypeByExtension(filepath.Ext(string(file_id)))
		}
	}
}

func apiHandleUploadedFiles(ctx *Ctx, fieldName string, maxNumFiles int, transform func([]byte) []byte) (fileNames []yodb.Text, filePaths []string) {
	if files := ctx.Http.Req.MultipartForm.File[fieldName]; len(files) > 0 {
		dst_dir_path := Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"]
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
			dst_file_path := filepath.Join(dst_dir_path, dst_file_name)
			WriteFile(dst_file_path, data)
			fileNames, filePaths = append(fileNames, yodb.Text(dst_file_name)), append(filePaths, dst_file_path)
		}
	}
	return
}
