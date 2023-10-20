// Code generated by `yo/srv/codegen_apistuff.go` DO NOT EDIT
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
export let reqTimeoutMilliSec = 1234

export function setReqTimeoutMilliSec(timeout: number) {
    reqTimeoutMilliSec = timeout
}

export async function req<TIn, TOut>(methodPath: string, payload: TIn, urlQueryArgs?: { [_: string]: string }): Promise<TOut> {
    let rel_url = '/' + methodPath
    if (urlQueryArgs)
        rel_url += ('?' + new URLSearchParams(urlQueryArgs).toString())
    console.log('callAPI:', rel_url, payload)
    const resp = await fetch(rel_url, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
        cache: 'no-store', mode: 'same-origin', redirect: 'error', signal: AbortSignal.timeout(reqTimeoutMilliSec)
    })
    if (resp.status !== 200) {
        let body_text: string = '', body_err: any
        try { body_text = await resp.text() } catch (err) { if (err) body_err = err }
        throw ({ 'status_code': resp?.status, 'status_text': resp?.statusText, 'body_text': body_text.trim(), 'body_err': body_err })
    }
    userEmailAddr = resp?.headers?.get('X-Yo-User') ?? ''
    const json_resp = await resp.json()
    return json_resp as TOut
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
    toApiQueryExpr: () => object,
}

export class QueryExpr {
    __yoQOp: QueryOperator
    __yoQConds: QueryExpr[]
    __yoQOperands: QueryVal[]
    private constructor() { }
    and(...conds: QueryExpr[]): QueryExpr { return qAll(...[this as QueryExpr].concat(conds)) }
    or(...conds: QueryExpr[]): QueryExpr { return qAny(...[this as QueryExpr].concat(conds)) }
    not(): QueryExpr { return qNot(this as QueryExpr) }
    toApiQueryExpr(): object {
        const ret = {}
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
    toApiQueryExpr(): object {
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

const errsPostNew = ['PostNew_ExpectedNonEmptyPost', 'PostNew_ExpectedOnlyBuddyRecipients', 'PostNew_InvalidItemInFiles', 'PostNew_RepliedToPostDoesNotExist', 'TimedOut', 'Unauthorized'] as const
export type PostNewErr = typeof errsPostNew[number]
export async function apiPostNew(payload: Post, query?: {[_:string]:string}): Promise<Return_yo_db_I64_> {
	try {
		return req<Post, Return_yo_db_I64_>('_/postNew', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsPostNew.indexOf(err.body_text) >= 0))
			throw(new Err<PostNewErr>(err.body_text as PostNewErr))
		throw(err)
	}
}

const errsUserBy = ['TimedOut', 'Unauthorized', 'UserBy_ExpectedEitherNickNameOrEmailAddr'] as const
export type UserByErr = typeof errsUserBy[number]
export async function apiUserBy(payload: userBy_In, query?: {[_:string]:string}): Promise<User> {
	try {
		return req<userBy_In, User>('_/userBy', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserBy.indexOf(err.body_text) >= 0))
			throw(new Err<UserByErr>(err.body_text as UserByErr))
		throw(err)
	}
}

const errsUserSignIn = ['TimedOut', '___yo_authLogin_AccountDoesNotExist', '___yo_authLogin_EmailInvalid', '___yo_authLogin_EmailRequiredButMissing', '___yo_authLogin_OkButFailedToCreateSignedToken', '___yo_authLogin_WrongPassword'] as const
export type UserSignInErr = typeof errsUserSignIn[number]
export async function apiUserSignIn(payload: ApiAccountPayload, query?: {[_:string]:string}): Promise<Void> {
	try {
		return req<ApiAccountPayload, Void>('_/userSignIn', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserSignIn.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignInErr>(err.body_text as UserSignInErr))
		throw(err)
	}
}

const errsUserSignOut = ['TimedOut'] as const
export type UserSignOutErr = typeof errsUserSignOut[number]
export async function apiUserSignOut(payload: Void, query?: {[_:string]:string}): Promise<Void> {
	try {
		return req<Void, Void>('_/userSignOut', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserSignOut.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignOutErr>(err.body_text as UserSignOutErr))
		throw(err)
	}
}

const errsUserSignUp = ['TimedOut', '___yo_authLogin_AccountDoesNotExist', '___yo_authLogin_EmailInvalid', '___yo_authLogin_EmailRequiredButMissing', '___yo_authLogin_OkButFailedToCreateSignedToken', '___yo_authLogin_WrongPassword', '___yo_authRegister_EmailAddrAlreadyExists', '___yo_authRegister_EmailInvalid', '___yo_authRegister_EmailRequiredButMissing', '___yo_authRegister_PasswordInvalid', '___yo_authRegister_PasswordTooLong', '___yo_authRegister_PasswordTooShort'] as const
export type UserSignUpErr = typeof errsUserSignUp[number]
export async function apiUserSignUp(payload: ApiAccountPayload, query?: {[_:string]:string}): Promise<User> {
	try {
		return req<ApiAccountPayload, User>('_/userSignUp', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserSignUp.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignUpErr>(err.body_text as UserSignUpErr))
		throw(err)
	}
}

const errsUserUpdate = ['DbUpdExpectedIdGt0', 'DbUpdate_ExpectedChangesForUpdate', 'DbUpdate_ExpectedQueryForUpdate', 'TimedOut', 'Unauthorized', 'UserUpdate_NicknameAlreadyExists'] as const
export type UserUpdateErr = typeof errsUserUpdate[number]
export async function apiUserUpdate(payload: ApiUpdateArgs_haxsh_app_User_haxsh_app_UserField_, query?: {[_:string]:string}): Promise<Void> {
	try {
		return req<ApiUpdateArgs_haxsh_app_User_haxsh_app_UserField_, Void>('_/userUpdate', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserUpdate.indexOf(err.body_text) >= 0))
			throw(new Err<UserUpdateErr>(err.body_text as UserUpdateErr))
		throw(err)
	}
}

export type PostField = 'Id' | 'DtMade' | 'DtMod' | 'To' | 'Md' | 'Files' | 'Repl' | 'by.Id' | 'by.DtMade' | 'by.DtMod' | 'by.LastSeen' | 'by.Auth' | 'by.PicFileId' | 'by.Nick' | 'by.Btw' | 'by.Buddies' | 'Repl.Id' | 'Repl.DtMade' | 'Repl.DtMod' | 'Repl.To' | 'Repl.Md' | 'Repl.Files' | 'Repl.Repl'

export type UserField = 'Id' | 'DtMade' | 'DtMod' | 'LastSeen' | 'Auth' | 'PicFileId' | 'Nick' | 'Btw' | 'Buddies' | 'Auth.Id' | 'Auth.DtMade' | 'Auth.DtMod' | 'Auth.EmailAddr'

export type UserAuthField = 'Id' | 'DtMade' | 'DtMod' | 'EmailAddr'

export type Post = {
	DtMade?: DateTime
	DtMod?: DateTime
	Files: string[]
	Id: I64
	Md: string
	Repl: I64
	To: I64[]
}

export type User = {
	Auth: I64
	Btw: string
	Buddies: I64[]
	DtMade?: DateTime
	DtMod?: DateTime
	Id: I64
	LastSeen?: DateTime
	Nick: string
	PicFileId: string
}

export type userBy_In = {
	EmailAddr: string
	NickName: string
}

export type ApiUpdateArgs_haxsh_app_User_haxsh_app_UserField_ = {
	ChangedFields: UserField[]
	Changes: User
	Id: I64
}

export type DateTime = string
export type ApiAccountPayload = {
	EmailAddr: string
	PasswordPlain: string
}

export type Return_yo_db_I64_ = {
	Result: I64
}

export type Void = {
}
