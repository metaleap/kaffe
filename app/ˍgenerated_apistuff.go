// Code generated by `yo/srv/codegen_apistuff.go` DO NOT EDIT
package haxsh

import reflect "reflect"
import yosrv "yo/srv"
import util "yo/util"
import q "yo/db/query"

type _ = q.F // just in case of no other generated import users
type apiPkgInfo util.Void

func (apiPkgInfo) PkgName() string    { return "haxsh" }
func (me apiPkgInfo) PkgPath() string { return reflect.TypeOf(me).PkgPath() }

var haxshPkg = apiPkgInfo{}

func api[TIn any, TOut any](f func(*yosrv.ApiCtx[TIn, TOut]), failIfs ...yosrv.Fails) yosrv.ApiMethod {
	return yosrv.Api[TIn, TOut](f, failIfs...).From(haxshPkg)
}

const ErrPostDelete_InvalidPostId util.Err = "PostDelete_InvalidPostId"
const ErrPostNew_ExpectedEmptyFilesFieldWithUploadedFilesInMultipartForm util.Err = "PostNew_ExpectedEmptyFilesFieldWithUploadedFilesInMultipartForm"
const ErrPostNew_ExpectedNonEmptyPost util.Err = "PostNew_ExpectedNonEmptyPost"
const ErrPostNew_ExpectedOnlyBuddyRecipients util.Err = "PostNew_ExpectedOnlyBuddyRecipients"
const ErrDbUpdate_ExpectedChangesForUpdate util.Err = "DbUpdate_ExpectedChangesForUpdate"
const ErrDbUpdate_ExpectedQueryForUpdate util.Err = "DbUpdate_ExpectedQueryForUpdate"
const ErrUserBuddiesAdd_ExpectedEitherNickNameOrEmailAddr util.Err = "UserBuddiesAdd_ExpectedEitherNickNameOrEmailAddr"
const ErrUserBy_ExpectedEitherNickNameOrEmailAddr util.Err = "UserBy_ExpectedEitherNickNameOrEmailAddr"
const ErrUserSignIn_ExpectedPasswordAndNickOrEmailAddr util.Err = "UserSignIn_ExpectedPasswordAndNickOrEmailAddr"
const ErrUserSignIn_WrongPassword util.Err = "UserSignIn_WrongPassword"
const Err___yo_authLogin_AccountDoesNotExist util.Err = "___yo_authLogin_AccountDoesNotExist"
const Err___yo_authLogin_EmailInvalid util.Err = "___yo_authLogin_EmailInvalid"
const Err___yo_authLogin_EmailRequiredButMissing util.Err = "___yo_authLogin_EmailRequiredButMissing"
const Err___yo_authLogin_OkButFailedToCreateSignedToken util.Err = "___yo_authLogin_OkButFailedToCreateSignedToken"
const Err___yo_authLogin_WrongPassword util.Err = "___yo_authLogin_WrongPassword"
const ErrUserSignUpOrForgotPassword_EmailInvalid util.Err = "UserSignUpOrForgotPassword_EmailInvalid"
const ErrUserSignUpOrForgotPassword_EmailRequiredButMissing util.Err = "UserSignUpOrForgotPassword_EmailRequiredButMissing"
const Err___yo_authRegister_EmailAddrAlreadyExists util.Err = "___yo_authRegister_EmailAddrAlreadyExists"
const Err___yo_authRegister_EmailInvalid util.Err = "___yo_authRegister_EmailInvalid"
const Err___yo_authRegister_EmailRequiredButMissing util.Err = "___yo_authRegister_EmailRequiredButMissing"
const Err___yo_authRegister_PasswordInvalid util.Err = "___yo_authRegister_PasswordInvalid"
const Err___yo_authRegister_PasswordTooLong util.Err = "___yo_authRegister_PasswordTooLong"
const Err___yo_authRegister_PasswordTooShort util.Err = "___yo_authRegister_PasswordTooShort"
const ErrUserUpdate_ExpectedNonEmptyNickname util.Err = "UserUpdate_ExpectedNonEmptyNickname"
const ErrUserUpdate_NicknameAlreadyExists util.Err = "UserUpdate_NicknameAlreadyExists"
const PostDeleteId = q.F("Id")
const PostMonthsUtcWithUserIds = q.F("WithUserIds")
const PostNewBy = q.F("By")
const PostNewDtMade = q.F("DtMade")
const PostNewDtMod = q.F("DtMod")
const PostNewFileContentTypes = q.F("FileContentTypes")
const PostNewFiles = q.F("Files")
const PostNewHtm = q.F("Htm")
const PostNewId = q.F("Id")
const PostNewTo = q.F("To")
const PostsDeletedOutOfPostIds = q.F("OutOfPostIds")
const PostsForMonthUtcOnlyBy = q.F("OnlyBy")
const PostsForMonthUtcPeriod = q.F("Period")
const PostsRecentOnlyBy = q.F("OnlyBy")
const PostsRecentSince = q.F("Since")
const UserBuddiesAddNickOrEmailAddr = q.F("NickOrEmailAddr")
const UserByEmailAddr = q.F("EmailAddr")
const UserByNickName = q.F("NickName")
const UserSignInNickOrEmailAddr = q.F("NickOrEmailAddr")
const UserSignInPasswordPlain = q.F("PasswordPlain")
const UserSignUpOrForgotPasswordNickOrEmailAddr = q.F("NickOrEmailAddr")
const UserUpdateChangedFields = q.F("ChangedFields")
const UserUpdateChanges = q.F("Changes")
const UserUpdateId = q.F("Id")
