// Code generated by `yo/db/codegen_dbstuff.go`. DO NOT EDIT
package haxsh

import q "yo/db/query"

type UserField q.F

const (
	UserId             UserField = "Id"
	UserDtMade         UserField = "DtMade"
	UserDtMod          UserField = "DtMod"
	UserLastSeen       UserField = "LastSeen"
	UserAuth           UserField = "Auth"
	UserPicFileId      UserField = "PicFileId"
	UserNick           UserField = "Nick"
	UserBtw            UserField = "Btw"
	UserBuddies        UserField = "Buddies"
	UserAuth_Id        UserField = "Auth.Id"
	UserAuth_DtMade    UserField = "Auth.DtMade"
	UserAuth_DtMod     UserField = "Auth.DtMod"
	UserAuth_EmailAddr UserField = "Auth.EmailAddr"
	userAuth_pwdHashed UserField = "Auth.pwdHashed"
)

func (me UserField) ArrLen(a1 ...interface{}) q.Operand { return ((q.F)(me)).ArrLen(a1...) }
func (me UserField) Asc() q.OrderBy                     { return ((q.F)(me)).Asc() }
func (me UserField) Desc() q.OrderBy                    { return ((q.F)(me)).Desc() }
func (me UserField) Equal(a1 interface{}) q.Query       { return ((q.F)(me)).Equal(a1) }
func (me UserField) Eval(a1 interface{}, a2 func(q.C) q.F) interface{} {
	return ((q.F)(me)).Eval(a1, a2)
}
func (me UserField) F() q.F                                { return ((q.F)(me)).F() }
func (me UserField) GreaterOrEqual(a1 interface{}) q.Query { return ((q.F)(me)).GreaterOrEqual(a1) }
func (me UserField) GreaterThan(a1 interface{}) q.Query    { return ((q.F)(me)).GreaterThan(a1) }
func (me UserField) In(a1 ...interface{}) q.Query          { return ((q.F)(me)).In(a1...) }
func (me UserField) InArr(a1 interface{}) q.Query          { return ((q.F)(me)).InArr(a1) }
func (me UserField) LessOrEqual(a1 interface{}) q.Query    { return ((q.F)(me)).LessOrEqual(a1) }
func (me UserField) LessThan(a1 interface{}) q.Query       { return ((q.F)(me)).LessThan(a1) }
func (me UserField) Not() q.Query                          { return ((q.F)(me)).Not() }
func (me UserField) NotEqual(a1 interface{}) q.Query       { return ((q.F)(me)).NotEqual(a1) }
func (me UserField) NotIn(a1 ...interface{}) q.Query       { return ((q.F)(me)).NotIn(a1...) }
func (me UserField) NotInArr(a1 interface{}) q.Query       { return ((q.F)(me)).NotInArr(a1) }
func (me UserField) StrLen(a1 ...interface{}) q.Operand    { return ((q.F)(me)).StrLen(a1...) }

type PostField q.F

const (
	PostId           PostField = "Id"
	PostDtMade       PostField = "DtMade"
	PostDtMod        PostField = "DtMod"
	PostBy           PostField = "By"
	PostTo           PostField = "To"
	PostHtm          PostField = "Htm"
	PostFiles        PostField = "Files"
	PostBy_Id        PostField = "By.Id"
	PostBy_DtMade    PostField = "By.DtMade"
	PostBy_DtMod     PostField = "By.DtMod"
	PostBy_LastSeen  PostField = "By.LastSeen"
	PostBy_Auth      PostField = "By.Auth"
	PostBy_PicFileId PostField = "By.PicFileId"
	PostBy_Nick      PostField = "By.Nick"
	PostBy_Btw       PostField = "By.Btw"
	PostBy_Buddies   PostField = "By.Buddies"
)

func (me PostField) ArrLen(a1 ...interface{}) q.Operand { return ((q.F)(me)).ArrLen(a1...) }
func (me PostField) Asc() q.OrderBy                     { return ((q.F)(me)).Asc() }
func (me PostField) Desc() q.OrderBy                    { return ((q.F)(me)).Desc() }
func (me PostField) Equal(a1 interface{}) q.Query       { return ((q.F)(me)).Equal(a1) }
func (me PostField) Eval(a1 interface{}, a2 func(q.C) q.F) interface{} {
	return ((q.F)(me)).Eval(a1, a2)
}
func (me PostField) F() q.F                                { return ((q.F)(me)).F() }
func (me PostField) GreaterOrEqual(a1 interface{}) q.Query { return ((q.F)(me)).GreaterOrEqual(a1) }
func (me PostField) GreaterThan(a1 interface{}) q.Query    { return ((q.F)(me)).GreaterThan(a1) }
func (me PostField) In(a1 ...interface{}) q.Query          { return ((q.F)(me)).In(a1...) }
func (me PostField) InArr(a1 interface{}) q.Query          { return ((q.F)(me)).InArr(a1) }
func (me PostField) LessOrEqual(a1 interface{}) q.Query    { return ((q.F)(me)).LessOrEqual(a1) }
func (me PostField) LessThan(a1 interface{}) q.Query       { return ((q.F)(me)).LessThan(a1) }
func (me PostField) Not() q.Query                          { return ((q.F)(me)).Not() }
func (me PostField) NotEqual(a1 interface{}) q.Query       { return ((q.F)(me)).NotEqual(a1) }
func (me PostField) NotIn(a1 ...interface{}) q.Query       { return ((q.F)(me)).NotIn(a1...) }
func (me PostField) NotInArr(a1 interface{}) q.Query       { return ((q.F)(me)).NotInArr(a1) }
func (me PostField) StrLen(a1 ...interface{}) q.Operand    { return ((q.F)(me)).StrLen(a1...) }
