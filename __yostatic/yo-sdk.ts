// Code generated by `yo/srv/codegen_apistuff.go` DO NOT EDIT
export const Cfg_YO_AUTH_PWD_MIN_LEN = 6
export const Cfg_YO_API_IMPL_TIMEOUT_MS = 1320000

// prelude-yo-sdk.ts below, more generated code afterwards
export type I8 = number
export type I16 = number
export type I32 = number
export type I64 = number
export type U8 = number
export type U16 = number
export type U32 = number
export type U64 = number
export type F32 = number
export type F64 = number


export let userEmailAddr = ''
export let reqTimeoutMsForJsonApis = 4321
export let reqTimeoutMsForMultipartForms = 123456
export let reqMaxReqPayloadSizeMb = 0           // declaration only, generated code sets the value
export let errMaxReqPayloadSizeExceeded = ""    // declaration only, generated code sets the value

export async function req<TIn, TOut, TErr extends string>(methodPath: string, payload?: TIn | {}, formData?: FormData, urlQueryArgs?: { [_: string]: string }): Promise<TOut> {
    let rel_url = '/' + methodPath
    if (urlQueryArgs)
        rel_url += ('?' + new URLSearchParams(urlQueryArgs).toString())

    if (!payload)
        payload = {}
    const payload_json = JSON.stringify(payload)

    if ((reqMaxReqPayloadSizeMb > 0) && errMaxReqPayloadSizeExceeded && (payload_json.length > (1024 * 1024 * reqMaxReqPayloadSizeMb)))
        throw new Err<TErr>(errMaxReqPayloadSizeExceeded as TErr)

    if (formData) {
        formData.set("_", payload_json)
        if ((reqMaxReqPayloadSizeMb > 0) && errMaxReqPayloadSizeExceeded) {
            let req_payload_size = 0
            formData.forEach(_ => {
                const value = _.valueOf()
                const file = value as File
                if (typeof value === 'string')
                    req_payload_size += value.length
                else if (file && file.name && file.size && (typeof file.size === 'number') && (file.size > 0))
                    req_payload_size += file.size
            })
            if (req_payload_size > (1024 * 1024 * reqMaxReqPayloadSizeMb))
                throw new Err<TErr>(errMaxReqPayloadSizeExceeded as TErr)
        }
    }

    const resp = await fetch(rel_url, {
        method: 'POST', headers: (formData ? undefined : ({ 'Content-Type': 'application/json' })), body: (formData ? formData : payload_json),
        cache: 'no-store', mode: 'same-origin', redirect: 'error', signal: AbortSignal.timeout(formData ? reqTimeoutMsForMultipartForms : reqTimeoutMsForJsonApis),
    })
    if (resp.status !== 200) {
        let body_text: string = '', body_err: any
        try { body_text = await resp.text() } catch (err) { body_err = err }
        throw ({ 'status_code': resp?.status, 'status_text': resp?.statusText, 'body_text': body_text.trim(), 'body_err': body_err })
    }
    userEmailAddr = resp?.headers?.get('X-Yo-User') ?? ''

    const resp_str_raw = await resp.text()
    try {
        return JSON.parse(resp_str_raw) as TOut
    } catch (err) {
        console.warn(resp_str_raw || "bug: empty non-JSON response despite 200 OK")
        throw err
    }
}

export class Err<T extends string> extends Error {
    knownErr: T
    constructor(err: T) {
        super()
        this.knownErr = err
    }
}

type QueryOperator = 'EQ' | 'NE' | 'LT' | 'LE' | 'GT' | 'GE' | 'IN' | 'AND' | 'OR' | 'NOT'

export interface QueryVal {
    __yoQLitValue?: any,
    __yoQFieldName?: any
    toApiQueryExpr: () => object | null,
}

export class QueryExpr {
    __yoQOp: QueryOperator
    __yoQConds: QueryExpr[] = []
    __yoQOperands: QueryVal[] = []
    private constructor() { }
    and(...conds: QueryExpr[]): QueryExpr { return qAll(...[this as QueryExpr].concat(conds)) }
    or(...conds: QueryExpr[]): QueryExpr { return qAny(...[this as QueryExpr].concat(conds)) }
    not(): QueryExpr { return qNot(this as QueryExpr) }
    toApiQueryExpr(): object {
        const ret = {} as any
        if (this.__yoQOp === 'NOT')
            ret['NOT'] = this.__yoQConds[0].toApiQueryExpr()
        else if ((this.__yoQOp === 'AND') || (this.__yoQOp === 'OR'))
            ret[this.__yoQOp] = this.__yoQConds.map((_) => _.toApiQueryExpr())
        else
            ret[this.__yoQOp] = this.__yoQOperands.map((_) => _.toApiQueryExpr())
        return ret
    }
}

export class QVal<T extends (string | number | boolean | null)>  {
    __yoQLitValue: T
    constructor(literalValue: T) { this.__yoQLitValue = literalValue }
    equal(other: QueryVal): QueryExpr { return qEqual(this, other) }
    notEqual(other: QueryVal): QueryExpr { return qNotEqual(this, other) }
    lessThan(other: QueryVal): QueryExpr { return qLessThan(this, other) }
    lessOrEqual(other: QueryVal): QueryExpr { return qLessOrEqual(this, other) }
    greaterThan(other: QueryVal): QueryExpr { return qGreaterThan(this, other) }
    greaterOrEqual(other: QueryVal): QueryExpr { return qGreaterOrEqual(this, other) }
    in(...set: QueryVal[]): QueryExpr { return qIn(this, ...set) }
    toApiQueryExpr(): object | null {
        if (typeof this.__yoQLitValue === 'string')
            return { 'Str': this.__yoQLitValue ?? '' }
        if (typeof this.__yoQLitValue === 'number')
            return { 'Int': this.__yoQLitValue ?? 0 }
        if (typeof this.__yoQLitValue === 'boolean')
            return { 'Bool': this.__yoQLitValue ?? '' }
        return null
    }
}

export class QFld<T extends string> {
    __yoQFieldName: T
    constructor(fieldName: T) { this.__yoQFieldName = fieldName }
    equal(other: QueryVal): QueryExpr { return qEqual(this, other) }
    notEqual(other: QueryVal): QueryExpr { return qNotEqual(this, other) }
    lessThan(other: QueryVal): QueryExpr { return qLessThan(this, other) }
    lessOrEqual(other: QueryVal): QueryExpr { return qLessOrEqual(this, other) }
    greaterThan(other: QueryVal): QueryExpr { return qGreaterThan(this, other) }
    greaterOrEqual(other: QueryVal): QueryExpr { return qGreaterOrEqual(this, other) }
    in(...set: QueryVal[]): QueryExpr { return qIn(this, ...set) }
    toApiQueryExpr(): object { return { 'Fld': this.__yoQFieldName } }
}

function qAll(...conds: QueryExpr[]): QueryExpr { return { __yoQOp: 'AND', __yoQConds: conds } as QueryExpr }
function qAny(...conds: QueryExpr[]): QueryExpr { return { __yoQOp: 'OR', __yoQConds: conds } as QueryExpr }
function qNot(cond: QueryExpr): QueryExpr { return { __yoQOp: 'NOT', __yoQConds: [cond] } as QueryExpr }

function qEqual(x: QueryVal, y: QueryVal): QueryExpr { return { __yoQOp: 'EQ', __yoQOperands: [x, y] } as QueryExpr }
function qNotEqual(x: QueryVal, y: QueryVal): QueryExpr { return { __yoQOp: 'NE', __yoQOperands: [x, y] } as QueryExpr }
function qLessThan(x: QueryVal, y: QueryVal): QueryExpr { return { __yoQOp: 'LT', __yoQOperands: [x, y] } as QueryExpr }
function qLessOrEqual(x: QueryVal, y: QueryVal): QueryExpr { return { __yoQOp: 'LE', __yoQOperands: [x, y] } as QueryExpr }
function qGreaterThan(x: QueryVal, y: QueryVal): QueryExpr { return { __yoQOp: 'GT', __yoQOperands: [x, y] } as QueryExpr }
function qGreaterOrEqual(x: QueryVal, y: QueryVal): QueryExpr { return { __yoQOp: 'GE', __yoQOperands: [x, y] } as QueryExpr }
function qIn(x: QueryVal, ...set: QueryVal[]): QueryExpr { return { __yoQOp: 'IN', __yoQOperands: [x].concat(set) } as QueryExpr }

// prelude-yo-sdk.ts ends, the rest below is fully generated code only:

reqTimeoutMsForJsonApis = Cfg_YO_API_IMPL_TIMEOUT_MS

errMaxReqPayloadSizeExceeded = 'MissingOrExcessiveContentLength'

reqMaxReqPayloadSizeMb = 22

const errsPostDelete = ['MissingOrExcessiveContentLength', 'PostDelete_InvalidPostId', 'TimedOut', 'Unauthorized'] as const
export async function apiPostDelete(payload?: postDelete_In, formData?: FormData, query?: {[_:string]:string}): Promise<Void> {
	try {
		return await req<postDelete_In, Void, PostDeleteErr>('_/postDelete', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsPostDelete.indexOf(err.body_text) >= 0))
			throw(new Err<PostDeleteErr>(err.body_text as PostDeleteErr))
		throw(err)
	}
}
export type PostDeleteErr = typeof errsPostDelete[number]

const errsPostEmojiFullList = ['MissingOrExcessiveContentLength', 'TimedOut'] as const
export async function apiPostEmojiFullList(payload?: Void, formData?: FormData, query?: {[_:string]:string}): Promise<Return_map_string_string_> {
	try {
		return await req<Void, Return_map_string_string_, PostEmojiFullListErr>('_/postEmojiFullList', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsPostEmojiFullList.indexOf(err.body_text) >= 0))
			throw(new Err<PostEmojiFullListErr>(err.body_text as PostEmojiFullListErr))
		throw(err)
	}
}
export type PostEmojiFullListErr = typeof errsPostEmojiFullList[number]

const errsPostMonthsUtc = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function apiPostMonthsUtc(payload?: postMonthsUtc_In, formData?: FormData, query?: {[_:string]:string}): Promise<postMonthsUtc_Out> {
	try {
		return await req<postMonthsUtc_In, postMonthsUtc_Out, PostMonthsUtcErr>('_/postMonthsUtc', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsPostMonthsUtc.indexOf(err.body_text) >= 0))
			throw(new Err<PostMonthsUtcErr>(err.body_text as PostMonthsUtcErr))
		throw(err)
	}
}
export type PostMonthsUtcErr = typeof errsPostMonthsUtc[number]

const errsPostNew = ['MissingOrExcessiveContentLength', 'PostNew_ExpectedEmptyFilesFieldWithUploadedFilesInMultipartForm', 'PostNew_ExpectedNonEmptyPost', 'PostNew_ExpectedOnlyBuddyRecipients', 'TimedOut', 'Unauthorized'] as const
export async function apiPostNew(payload?: Post, formData?: FormData, query?: {[_:string]:string}): Promise<Return_yo_db_I64_> {
	try {
		return await req<Post, Return_yo_db_I64_, PostNewErr>('_/postNew', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsPostNew.indexOf(err.body_text) >= 0))
			throw(new Err<PostNewErr>(err.body_text as PostNewErr))
		throw(err)
	}
}
export type PostNewErr = typeof errsPostNew[number]

const errsPostsDeleted = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function apiPostsDeleted(payload?: postsDeleted_In, formData?: FormData, query?: {[_:string]:string}): Promise<postsDeleted_Out> {
	try {
		return await req<postsDeleted_In, postsDeleted_Out, PostsDeletedErr>('_/postsDeleted', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsPostsDeleted.indexOf(err.body_text) >= 0))
			throw(new Err<PostsDeletedErr>(err.body_text as PostsDeletedErr))
		throw(err)
	}
}
export type PostsDeletedErr = typeof errsPostsDeleted[number]

const errsPostsForMonthUtc = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function apiPostsForMonthUtc(payload?: ApiArgPeriod, formData?: FormData, query?: {[_:string]:string}): Promise<PostsListResult> {
	try {
		return await req<ApiArgPeriod, PostsListResult, PostsForMonthUtcErr>('_/postsForMonthUtc', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsPostsForMonthUtc.indexOf(err.body_text) >= 0))
			throw(new Err<PostsForMonthUtcErr>(err.body_text as PostsForMonthUtcErr))
		throw(err)
	}
}
export type PostsForMonthUtcErr = typeof errsPostsForMonthUtc[number]

const errsPostsRecent = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function apiPostsRecent(payload?: postsRecent_In, formData?: FormData, query?: {[_:string]:string}): Promise<PostsListResult> {
	try {
		return await req<postsRecent_In, PostsListResult, PostsRecentErr>('_/postsRecent', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsPostsRecent.indexOf(err.body_text) >= 0))
			throw(new Err<PostsRecentErr>(err.body_text as PostsRecentErr))
		throw(err)
	}
}
export type PostsRecentErr = typeof errsPostsRecent[number]

const errsUserBuddies = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function apiUserBuddies(payload?: Void, formData?: FormData, query?: {[_:string]:string}): Promise<userBuddies_Out> {
	try {
		return await req<Void, userBuddies_Out, UserBuddiesErr>('_/userBuddies', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsUserBuddies.indexOf(err.body_text) >= 0))
			throw(new Err<UserBuddiesErr>(err.body_text as UserBuddiesErr))
		throw(err)
	}
}
export type UserBuddiesErr = typeof errsUserBuddies[number]

const errsUserBuddiesAdd = ['DbUpdate_ExpectedChangesForUpdate', 'DbUpdate_ExpectedQueryForUpdate', 'MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', 'UserBuddiesAdd_ExpectedEitherNickNameOrEmailAddr'] as const
export async function apiUserBuddiesAdd(payload?: userBuddiesAdd_In, formData?: FormData, query?: {[_:string]:string}): Promise<userBuddiesAdd_Out> {
	try {
		return await req<userBuddiesAdd_In, userBuddiesAdd_Out, UserBuddiesAddErr>('_/userBuddiesAdd', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsUserBuddiesAdd.indexOf(err.body_text) >= 0))
			throw(new Err<UserBuddiesAddErr>(err.body_text as UserBuddiesAddErr))
		throw(err)
	}
}
export type UserBuddiesAddErr = typeof errsUserBuddiesAdd[number]

const errsUserBy = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', 'UserBy_ExpectedEitherNickNameOrEmailAddr'] as const
export async function apiUserBy(payload?: userBy_In, formData?: FormData, query?: {[_:string]:string}): Promise<User> {
	try {
		return await req<userBy_In, User, UserByErr>('_/userBy', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsUserBy.indexOf(err.body_text) >= 0))
			throw(new Err<UserByErr>(err.body_text as UserByErr))
		throw(err)
	}
}
export type UserByErr = typeof errsUserBy[number]

const errsUserSignInOrReset = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', 'UserSignInOrReset_ExpectedPasswordAndNickOrEmailAddr', 'UserSignInOrReset_WrongPassword', '___yo_authLoginOrFinalizePwdReset_AccountDoesNotExist', '___yo_authLoginOrFinalizePwdReset_EmailInvalid', '___yo_authLoginOrFinalizePwdReset_EmailRequiredButMissing', '___yo_authLoginOrFinalizePwdReset_NewPasswordExpectedToDiffer', '___yo_authLoginOrFinalizePwdReset_NewPasswordInvalid', '___yo_authLoginOrFinalizePwdReset_NewPasswordTooLong', '___yo_authLoginOrFinalizePwdReset_NewPasswordTooShort', '___yo_authLoginOrFinalizePwdReset_OkButFailedToCreateSignedToken', '___yo_authLoginOrFinalizePwdReset_WrongPassword'] as const
export async function apiUserSignInOrReset(payload?: ApiUserSignInOrReset, formData?: FormData, query?: {[_:string]:string}): Promise<Void> {
	try {
		return await req<ApiUserSignInOrReset, Void, UserSignInOrResetErr>('_/userSignInOrReset', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsUserSignInOrReset.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignInOrResetErr>(err.body_text as UserSignInOrResetErr))
		throw(err)
	}
}
export type UserSignInOrResetErr = typeof errsUserSignInOrReset[number]

const errsUserSignOut = ['MissingOrExcessiveContentLength', 'TimedOut'] as const
export async function apiUserSignOut(payload?: Void, formData?: FormData, query?: {[_:string]:string}): Promise<Void> {
	try {
		return await req<Void, Void, UserSignOutErr>('_/userSignOut', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsUserSignOut.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignOutErr>(err.body_text as UserSignOutErr))
		throw(err)
	}
}
export type UserSignOutErr = typeof errsUserSignOut[number]

const errsUserSignUpOrForgotPassword = ['MissingOrExcessiveContentLength', 'TimedOut', 'UserSignUpOrForgotPassword_EmailInvalid', 'UserSignUpOrForgotPassword_EmailRequiredButMissing', '___yo_authRegister_EmailAddrAlreadyExists', '___yo_authRegister_EmailInvalid', '___yo_authRegister_EmailRequiredButMissing', '___yo_authRegister_PasswordInvalid', '___yo_authRegister_PasswordTooLong', '___yo_authRegister_PasswordTooShort'] as const
export async function apiUserSignUpOrForgotPassword(payload?: ApiNickOrEmailAddr, formData?: FormData, query?: {[_:string]:string}): Promise<Void> {
	try {
		return await req<ApiNickOrEmailAddr, Void, UserSignUpOrForgotPasswordErr>('_/userSignUpOrForgotPassword', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsUserSignUpOrForgotPassword.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignUpOrForgotPasswordErr>(err.body_text as UserSignUpOrForgotPasswordErr))
		throw(err)
	}
}
export type UserSignUpOrForgotPasswordErr = typeof errsUserSignUpOrForgotPassword[number]

const errsUserUpdate = ['DbUpdExpectedIdGt0', 'DbUpdate_ExpectedChangesForUpdate', 'DbUpdate_ExpectedQueryForUpdate', 'MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', 'UserUpdate_ExpectedNonEmptyNickname', 'UserUpdate_NicknameAlreadyExists'] as const
export async function apiUserUpdate(payload?: ApiUpdateArgs_haxsh_app_User_haxsh_app_UserField_, formData?: FormData, query?: {[_:string]:string}): Promise<Void> {
	try {
		return await req<ApiUpdateArgs_haxsh_app_User_haxsh_app_UserField_, Void, UserUpdateErr>('_/userUpdate', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errsUserUpdate.indexOf(err.body_text) >= 0))
			throw(new Err<UserUpdateErr>(err.body_text as UserUpdateErr))
		throw(err)
	}
}
export type UserUpdateErr = typeof errsUserUpdate[number]

export type PostField = 'Id' | 'DtMade' | 'DtMod' | 'By' | 'To' | 'Htm' | 'Files'

export type UserField = 'Id' | 'DtMade' | 'DtMod' | 'LastSeen' | 'Auth' | 'PicFileId' | 'Nick' | 'Btw' | 'Buddies'

export type fileDelReqField = 'Id' | 'DtMade' | 'DtMod' | 'FileNames'

export type UserAuthField = 'Id' | 'DtMade' | 'DtMod' | 'EmailAddr'

export type UserPwdReqField = 'Id' | 'DtMade' | 'DtMod' | 'EmailAddr' | 'DoneMailReqId' | 'DtFinalized'

export type JobDefField = 'Id' | 'DtMade' | 'DtMod' | 'Name' | 'JobTypeId' | 'Disabled' | 'AllowManualJobRuns' | 'Schedules' | 'TimeoutSecsTaskRun' | 'TimeoutSecsJobRunPrepAndFinalize' | 'MaxTaskRetries' | 'DeleteAfterDays' | 'StoreAndRunTasklessJobs'

export type JobRunField = 'Id' | 'DtMade' | 'DtMod' | 'Version' | 'JobTypeId' | 'JobDef' | 'DueTime' | 'StartTime' | 'FinishTime' | 'AutoScheduled' | 'ScheduledNextAfter' | 'DurationPrepSecs' | 'DurationFinalizeSecs'

export type JobTaskField = 'Id' | 'DtMade' | 'DtMod' | 'Version' | 'JobTypeId' | 'JobRun' | 'StartTime' | 'FinishTime' | 'Attempts'

export type MailReqField = 'Id' | 'DtMade' | 'DtMod' | 'TmplId' | 'TmplArgs' | 'MailTo'

export type ApiArgPeriod = {
	OnlyBy?: I64[]
	Period?: YearAndMonth
}

export type ApiNickOrEmailAddr = {
	NickOrEmailAddr?: string
}

export type ApiUserSignInOrReset = {
	NickOrEmailAddr?: string
	Password2Plain?: string
	PasswordPlain?: string
}

export type Post = {
	By?: I64
	DtMade?: DateTime
	DtMod?: DateTime
	FileContentTypes?: string[]
	Files?: string[]
	Htm?: string
	Id?: I64
	To?: I64[]
}

export type PostsListResult = {
	NextSince: DateTime
	Posts: Post[]
	UnreadCounts: { [_:string]: I64 }
}

export type User = {
	Auth?: I64
	Btw?: string
	BtwEmoji?: string
	Buddies?: I64[]
	DtMade?: DateTime
	DtMod?: DateTime
	Id?: I64
	LastSeen?: DateTime
	Nick?: string
	Offline?: boolean
	PicFileId?: string
}

export type YearAndMonth = {
	Month?: U8
	Year?: U16
}

export type postDelete_In = {
	Id?: I64
}

export type postMonthsUtc_In = {
	WithUserIds?: I64[]
}

export type postMonthsUtc_Out = {
	Periods: YearAndMonth[]
}

export type postsDeleted_In = {
	OutOfPostIds?: I64[]
}

export type postsDeleted_Out = {
	DeletedPostIds: I64[]
}

export type postsRecent_In = {
	OnlyBy?: I64[]
	Since?: DateTime
}

export type userBuddiesAdd_In = {
	NickOrEmailAddr?: string
}

export type userBuddiesAdd_Out = {
	Done: boolean
}

export type userBuddies_Out = {
	Buddies: User[]
	BuddyRequestsBy: User[]
}

export type userBy_In = {
	EmailAddr?: string
	NickName?: string
}

export type ApiUpdateArgs_haxsh_app_User_haxsh_app_UserField_ = {
	ChangedFields?: UserField[]
	Changes?: User
	Id?: I64
}

export type DateTime = string
export type Return_map_string_string_ = {
	Result: { [_:string]: string }
}

export type Return_yo_db_I64_ = {
	Result: I64
}

export type Void = {
}
