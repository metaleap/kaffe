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
    let uri = '/' + methodPath
    if (urlQueryArgs)
        uri += '?' + new URLSearchParams(urlQueryArgs).toString()
    console.log('callAPI:', uri, payload)
    const resp = await fetch(uri, {
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

export const errsUserSignIn = ['TimedOut', '___yo_authLogin_AccountDoesNotExist', '___yo_authLogin_EmailInvalid', '___yo_authLogin_EmailRequiredButMissing', '___yo_authLogin_OkButFailedToCreateSignedToken', '___yo_authLogin_WrongPassword'] as const
export type UserSignInErr = typeof errsUserSignIn[number]
export async function callUserSignIn(payload: ApiAccountPayload, query?: {[_:string]:string}): Promise<Void> {
	try {
		return req<ApiAccountPayload, Void>('userSignIn', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserSignIn.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignInErr>(err.body_text as UserSignInErr))
		throw(err)
	}
}

export const errsUserSignOut = ['TimedOut'] as const
export type UserSignOutErr = typeof errsUserSignOut[number]
export async function callUserSignOut(payload: Void, query?: {[_:string]:string}): Promise<Void> {
	try {
		return req<Void, Void>('userSignOut', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserSignOut.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignOutErr>(err.body_text as UserSignOutErr))
		throw(err)
	}
}

export const errsUserSignUp = ['DbWriteRequestAcceptedWithoutErrButNotStoredEither', 'TimedOut', '___yo_authLogin_AccountDoesNotExist', '___yo_authLogin_EmailInvalid', '___yo_authLogin_EmailRequiredButMissing', '___yo_authLogin_OkButFailedToCreateSignedToken', '___yo_authLogin_WrongPassword', '___yo_authRegister_EmailAddrAlreadyExists', '___yo_authRegister_EmailInvalid', '___yo_authRegister_EmailRequiredButMissing', '___yo_authRegister_PasswordInvalid', '___yo_authRegister_PasswordTooLong', '___yo_authRegister_PasswordTooShort'] as const
export type UserSignUpErr = typeof errsUserSignUp[number]
export async function callUserSignUp(payload: ApiAccountPayload, query?: {[_:string]:string}): Promise<User> {
	try {
		return req<ApiAccountPayload, User>('userSignUp', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserSignUp.indexOf(err.body_text) >= 0))
			throw(new Err<UserSignUpErr>(err.body_text as UserSignUpErr))
		throw(err)
	}
}

export const errsUserUpdate = ['DbUpdate_ExpectedChangesForUpdate', 'DbUpdate_ExpectedQueryForUpdate', 'DbWriteRequestAcceptedWithoutErrButNotStoredEither', 'ErrDbUpdExpectedIdGt0', 'TimedOut', 'Unauthorized'] as const
export type UserUpdateErr = typeof errsUserUpdate[number]
export async function callUserUpdate(payload: ApiUpdateArgs_main_User_, query?: {[_:string]:string}): Promise<Void> {
	try {
		return req<ApiUpdateArgs_main_User_, Void>('userUpdate', payload, query)
	} catch(err) {
		if (err && err['body_text'] && (errsUserUpdate.indexOf(err.body_text) >= 0))
			throw(new Err<UserUpdateErr>(err.body_text as UserUpdateErr))
		throw(err)
	}
}

export type UserField = 'Id' | 'Created' | 'Auth' | 'NickName'

export type UserAuthField = 'Id' | 'Created' | 'EmailAddr'

export type User = {
	Auth: I64
	Created?: DateTime
	Id: I64
	NickName: string
}

export type ApiUpdateArgs_main_User_ = {
	Changes: User
	Id: I64
	IncludingEmptyOrMissingFields: boolean
}

export type DateTime = string
export type ApiAccountPayload = {
	EmailAddr: string
	PasswordPlain: string
}

export type Void = {
}
