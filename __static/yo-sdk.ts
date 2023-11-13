// Code generated by `yo/srv/codegen_apistuff.go` DO NOT EDIT
export const Cfg_YO_API_IMPL_TIMEOUT_MS = 4000
export const Cfg_YO_AUTH_PWD_MIN_LEN = 6

// prelude-yo-sdk.ts (non-generated) below, more generated code afterwards
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


export let apiBaseUrl = ''
export let userEmailAddr = ''
export let reqTimeoutMsForJsonApis = 4321
export let reqTimeoutMsForMultipartForms = 123456
export let reqMaxReqPayloadSizeMb = 0           // declaration only, generated code sets the value
export let reqMaxReqMultipartSizeMb = 0         // declaration only, generated code sets the value
export let errMaxReqPayloadSizeExceeded = ""    // declaration only, generated code sets the value

let doFetch = fetch
export function setCustomFetch(customFetch: (reqUrl: string, reqInit?: object) => Promise<Response>) {
    doFetch = (customFetch ?? fetch) as any
}
export function setApiBaseUrl(newApiBaseUrl: string) { apiBaseUrl = newApiBaseUrl }

export async function req<TIn, TOut, TErr extends string>(methodPath: string, payload?: TIn | {}, formData?: FormData, urlQueryArgs?: { [_: string]: string }): Promise<TOut> {
    let rel_url = '/' + methodPath
    if (urlQueryArgs)
        rel_url += ('?' + new URLSearchParams(urlQueryArgs).toString())

    if (!payload)
        payload = {}
    const payload_json = JSON.stringify(payload)

    if (formData) {
        formData.set("_", payload_json)

        let req_payload_size = 0
        formData.forEach(_ => {
            const value = _.valueOf()
            const file = value as File
            if (typeof value === 'string')
                req_payload_size += value.length
            else if (file && file.name && file.size && (typeof file.size === 'number') && (file.size > 0))
                req_payload_size += file.size
        })
        if (req_payload_size > (1024 * 1024 * reqMaxReqMultipartSizeMb))
            throw new Err<TErr>(errMaxReqPayloadSizeExceeded as TErr)
    } else if (payload_json.length > (1024 * 1024 * reqMaxReqPayloadSizeMb))
        throw new Err<TErr>(errMaxReqPayloadSizeExceeded as TErr)

    const resp = await doFetch(apiBaseUrl + rel_url, {
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

reqMaxReqPayloadSizeMb = 1

reqMaxReqMultipartSizeMb = 22

const errs__postDelete = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', '__postDelete_InvalidPostId'] as const
export async function api__postDelete(payload?: __postDelete_In, formData?: FormData, query?: {[_:string]:string}): Promise<None> {
	try {
		return await req<__postDelete_In, None, __postDeleteErr>('_/postDelete', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__postDelete.indexOf(err.body_text) >= 0))
			throw(new Err<__postDeleteErr>(err.body_text as __postDeleteErr))
		throw(err)
	}
}
export type __postDeleteErr = typeof errs__postDelete[number]

const errs__postEmojiFullList = ['MissingOrExcessiveContentLength', 'TimedOut'] as const
export async function api__postEmojiFullList(payload?: None, formData?: FormData, query?: {[_:string]:string}): Promise<Return_map_string_string_> {
	try {
		return await req<None, Return_map_string_string_, __postEmojiFullListErr>('_/postEmojiFullList', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__postEmojiFullList.indexOf(err.body_text) >= 0))
			throw(new Err<__postEmojiFullListErr>(err.body_text as __postEmojiFullListErr))
		throw(err)
	}
}
export type __postEmojiFullListErr = typeof errs__postEmojiFullList[number]

const errs__postMonthsUtc = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function api__postMonthsUtc(payload?: __postMonthsUtc_In, formData?: FormData, query?: {[_:string]:string}): Promise<__postMonthsUtc_Out> {
	try {
		return await req<__postMonthsUtc_In, __postMonthsUtc_Out, __postMonthsUtcErr>('_/postMonthsUtc', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__postMonthsUtc.indexOf(err.body_text) >= 0))
			throw(new Err<__postMonthsUtcErr>(err.body_text as __postMonthsUtcErr))
		throw(err)
	}
}
export type __postMonthsUtcErr = typeof errs__postMonthsUtc[number]

const errs__postNew = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', '__postNew_ExpectedEmptyFilesFieldWithUploadedFilesInMultipartForm', '__postNew_ExpectedNonEmptyPost', '__postNew_ExpectedOnlyBuddyRecipients'] as const
export async function api__postNew(payload?: Post, formData?: FormData, query?: {[_:string]:string}): Promise<Return_yo_db_I64_> {
	try {
		return await req<Post, Return_yo_db_I64_, __postNewErr>('_/postNew', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__postNew.indexOf(err.body_text) >= 0))
			throw(new Err<__postNewErr>(err.body_text as __postNewErr))
		throw(err)
	}
}
export type __postNewErr = typeof errs__postNew[number]

const errs__postsDeleted = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function api__postsDeleted(payload?: __postsDeleted_In, formData?: FormData, query?: {[_:string]:string}): Promise<__postsDeleted_Out> {
	try {
		return await req<__postsDeleted_In, __postsDeleted_Out, __postsDeletedErr>('_/postsDeleted', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__postsDeleted.indexOf(err.body_text) >= 0))
			throw(new Err<__postsDeletedErr>(err.body_text as __postsDeletedErr))
		throw(err)
	}
}
export type __postsDeletedErr = typeof errs__postsDeleted[number]

const errs__postsForMonthUtc = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function api__postsForMonthUtc(payload?: ApiArgPeriod, formData?: FormData, query?: {[_:string]:string}): Promise<PostsListResult> {
	try {
		return await req<ApiArgPeriod, PostsListResult, __postsForMonthUtcErr>('_/postsForMonthUtc', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__postsForMonthUtc.indexOf(err.body_text) >= 0))
			throw(new Err<__postsForMonthUtcErr>(err.body_text as __postsForMonthUtcErr))
		throw(err)
	}
}
export type __postsForMonthUtcErr = typeof errs__postsForMonthUtc[number]

const errs__postsRecent = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function api__postsRecent(payload?: __postsRecent_In, formData?: FormData, query?: {[_:string]:string}): Promise<PostsListResult> {
	try {
		return await req<__postsRecent_In, PostsListResult, __postsRecentErr>('_/postsRecent', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__postsRecent.indexOf(err.body_text) >= 0))
			throw(new Err<__postsRecentErr>(err.body_text as __postsRecentErr))
		throw(err)
	}
}
export type __postsRecentErr = typeof errs__postsRecent[number]

const errs__userBuddies = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized'] as const
export async function api__userBuddies(payload?: None, formData?: FormData, query?: {[_:string]:string}): Promise<__userBuddies_Out> {
	try {
		return await req<None, __userBuddies_Out, __userBuddiesErr>('_/userBuddies', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__userBuddies.indexOf(err.body_text) >= 0))
			throw(new Err<__userBuddiesErr>(err.body_text as __userBuddiesErr))
		throw(err)
	}
}
export type __userBuddiesErr = typeof errs__userBuddies[number]

const errs__userBuddiesAdd = ['DbUpdate_ExpectedChangesForUpdate', 'DbUpdate_ExpectedQueryForUpdate', 'MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', '__userBuddiesAdd_ExpectedEitherNickNameOrEmailAddr'] as const
export async function api__userBuddiesAdd(payload?: __userBuddiesAdd_In, formData?: FormData, query?: {[_:string]:string}): Promise<__userBuddiesAdd_Out> {
	try {
		return await req<__userBuddiesAdd_In, __userBuddiesAdd_Out, __userBuddiesAddErr>('_/userBuddiesAdd', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__userBuddiesAdd.indexOf(err.body_text) >= 0))
			throw(new Err<__userBuddiesAddErr>(err.body_text as __userBuddiesAddErr))
		throw(err)
	}
}
export type __userBuddiesAddErr = typeof errs__userBuddiesAdd[number]

const errs__userBy = ['MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', '__userBy_ExpectedEitherNickNameOrEmailAddr'] as const
export async function api__userBy(payload?: __userBy_In, formData?: FormData, query?: {[_:string]:string}): Promise<User> {
	try {
		return await req<__userBy_In, User, __userByErr>('_/userBy', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__userBy.indexOf(err.body_text) >= 0))
			throw(new Err<__userByErr>(err.body_text as __userByErr))
		throw(err)
	}
}
export type __userByErr = typeof errs__userBy[number]

const errs__userSignInOrReset = ['DbUpdate_ExpectedChangesForUpdate', 'DbUpdate_ExpectedQueryForUpdate', 'MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', '___yo_authLoginOrFinalizePwdReset_AccountDoesNotExist', '___yo_authLoginOrFinalizePwdReset_EmailInvalid', '___yo_authLoginOrFinalizePwdReset_EmailRequiredButMissing', '___yo_authLoginOrFinalizePwdReset_NewPasswordExpectedToDiffer', '___yo_authLoginOrFinalizePwdReset_NewPasswordInvalid', '___yo_authLoginOrFinalizePwdReset_NewPasswordTooLong', '___yo_authLoginOrFinalizePwdReset_NewPasswordTooShort', '___yo_authLoginOrFinalizePwdReset_OkButFailedToCreateSignedToken', '___yo_authLoginOrFinalizePwdReset_PwdReqExpired', '___yo_authLoginOrFinalizePwdReset_WrongPassword', '__userSignInOrReset_ExpectedPasswordAndNickOrEmailAddr', '__userSignInOrReset_WrongPassword'] as const
export async function api__userSignInOrReset(payload?: ApiUserSignInOrReset, formData?: FormData, query?: {[_:string]:string}): Promise<None> {
	try {
		return await req<ApiUserSignInOrReset, None, __userSignInOrResetErr>('_/userSignInOrReset', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__userSignInOrReset.indexOf(err.body_text) >= 0))
			throw(new Err<__userSignInOrResetErr>(err.body_text as __userSignInOrResetErr))
		throw(err)
	}
}
export type __userSignInOrResetErr = typeof errs__userSignInOrReset[number]

const errs__userSignOut = ['MissingOrExcessiveContentLength', 'TimedOut'] as const
export async function api__userSignOut(payload?: None, formData?: FormData, query?: {[_:string]:string}): Promise<None> {
	try {
		return await req<None, None, __userSignOutErr>('_/userSignOut', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__userSignOut.indexOf(err.body_text) >= 0))
			throw(new Err<__userSignOutErr>(err.body_text as __userSignOutErr))
		throw(err)
	}
}
export type __userSignOutErr = typeof errs__userSignOut[number]

const errs__userSignUpOrForgotPassword = ['MissingOrExcessiveContentLength', 'TimedOut', '___yo_authRegister_EmailAddrAlreadyExists', '___yo_authRegister_EmailInvalid', '___yo_authRegister_EmailRequiredButMissing', '___yo_authRegister_PasswordInvalid', '___yo_authRegister_PasswordTooLong', '___yo_authRegister_PasswordTooShort', '__userSignUpOrForgotPassword_EmailInvalid', '__userSignUpOrForgotPassword_EmailRequiredButMissing'] as const
export async function api__userSignUpOrForgotPassword(payload?: ApiNickOrEmailAddr, formData?: FormData, query?: {[_:string]:string}): Promise<None> {
	try {
		return await req<ApiNickOrEmailAddr, None, __userSignUpOrForgotPasswordErr>('_/userSignUpOrForgotPassword', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__userSignUpOrForgotPassword.indexOf(err.body_text) >= 0))
			throw(new Err<__userSignUpOrForgotPasswordErr>(err.body_text as __userSignUpOrForgotPasswordErr))
		throw(err)
	}
}
export type __userSignUpOrForgotPasswordErr = typeof errs__userSignUpOrForgotPassword[number]

const errs__userUpdate = ['DbUpdExpectedIdGt0', 'DbUpdate_ExpectedChangesForUpdate', 'DbUpdate_ExpectedQueryForUpdate', 'MissingOrExcessiveContentLength', 'TimedOut', 'Unauthorized', '__userUpdate_ExpectedNonEmptyNickname', '__userUpdate_NicknameAlreadyExists'] as const
export async function api__userUpdate(payload?: ApiUpdateArgs_kaffe_app_User_kaffe_app_UserField_, formData?: FormData, query?: {[_:string]:string}): Promise<None> {
	try {
		return await req<ApiUpdateArgs_kaffe_app_User_kaffe_app_UserField_, None, __userUpdateErr>('_/userUpdate', payload, formData, query)
	} catch(err: any) {
		if (err && err['body_text'] && (errs__userUpdate.indexOf(err.body_text) >= 0))
			throw(new Err<__userUpdateErr>(err.body_text as __userUpdateErr))
		throw(err)
	}
}
export type __userUpdateErr = typeof errs__userUpdate[number]

export type PostField = 'Id' | 'DtMade' | 'DtMod' | 'By' | 'To' | 'Htm' | 'Files'

export type UserField = 'Id' | 'DtMade' | 'DtMod' | 'LastSeen' | 'Auth' | 'PicFileId' | 'Nick' | 'Btw' | 'Buddies'

export type fileDelReqField = 'Id' | 'DtMade' | 'DtMod' | 'FileNames'

export type ErrEntryField = 'Id' | 'DtMade' | 'DtMod' | 'Err' | 'StackTrace' | 'CtxVals' | 'HttpUrlPath' | 'HttpFullUri' | 'NumCaught' | 'JobRunId' | 'JobTaskId' | 'DbTx'

export type UserAuthField = 'Id' | 'DtMade' | 'DtMod' | 'EmailAddr' | 'FailedLoginAttempts'

export type UserPwdReqField = 'Id' | 'DtMade' | 'DtMod' | 'EmailAddr' | 'DoneMailReqId'

export type JobDefField = 'Id' | 'DtMade' | 'DtMod' | 'Name' | 'JobTypeId' | 'Disabled' | 'AllowManualJobRuns' | 'Schedules' | 'TimeoutSecsTaskRun' | 'TimeoutSecsJobRunPrepAndFinalize' | 'MaxTaskRetries' | 'DeleteAfterDays' | 'RunTasklessJobs'

export type JobRunField = 'Id' | 'DtMade' | 'DtMod' | 'Version' | 'JobTypeId' | 'JobDef' | 'CancelReason' | 'DueTime' | 'StartTime' | 'FinishTime' | 'AutoScheduled' | 'ScheduledNextAfter' | 'DurationPrepSecs' | 'DurationFinalizeSecs'

export type JobTaskField = 'Id' | 'DtMade' | 'DtMod' | 'Version' | 'JobTypeId' | 'JobRun' | 'StartTime' | 'FinishTime' | 'Attempts'

export type MailReqField = 'Id' | 'DtMade' | 'DtMod' | 'TmplId' | 'TmplArgs' | 'MailTo'

export type __postDelete_In = {
	Id?: I64
}

export type __postMonthsUtc_In = {
	WithUserIds?: I64[]
}

export type __postMonthsUtc_Out = {
	Periods: YearAndMonth[]
}

export type __postsDeleted_In = {
	OutOfPostIds?: I64[]
}

export type __postsDeleted_Out = {
	DeletedPostIds: I64[]
}

export type __postsRecent_In = {
	OnlyBy?: I64[]
	Since?: DateTime
}

export type __userBuddiesAdd_In = {
	NickOrEmailAddr?: string
}

export type __userBuddiesAdd_Out = {
	Done: boolean
}

export type __userBuddies_Out = {
	Buddies: User[]
	BuddyRequestsBy: User[]
}

export type __userBy_In = {
	EmailAddr?: string
	NickName?: string
}

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

export type ApiUpdateArgs_kaffe_app_User_kaffe_app_UserField_ = {
	ChangedFields?: UserField[]
	Changes?: User
	Id?: I64
}

export type DateTime = string
export type None = {
}

export type Return_map_string_string_ = {
	Result: { [_:string]: string }
}

export type Return_yo_db_I64_ = {
	Result: I64
}
