// Code generated by `yo/db/codegen_dbstuff.go`. DO NOT EDIT
package haxsh

import q "yo/db/query"

type UserField q.F

const (
	UserId             UserField = "Id"
	UserCreated        UserField = "Created"
	UserLastSeen       UserField = "LastSeen"
	UserAuth           UserField = "Auth"
	UserPicFileId      UserField = "PicFileId"
	UserNick           UserField = "Nick"
	UserBtw            UserField = "Btw"
	UserBtwDt          UserField = "BtwDt"
	UserBuddies        UserField = "Buddies"
	UserAuth_Id        UserField = "Auth.Id"
	UserAuth_Created   UserField = "Auth.Created"
	UserAuth_EmailAddr UserField = "Auth.EmailAddr"
	userAuth_pwdHashed UserField = "Auth.pwdHashed"
)

func (me UserField) Asc() q.OrderBy               { return ((q.F)(me)).Asc() }
func (me UserField) Desc() q.OrderBy              { return ((q.F)(me)).Desc() }
func (me UserField) Equal(a1 interface{}) q.Query { return ((q.F)(me)).Equal(a1) }
func (me UserField) Eval(a1 interface{}, a2 func(q.C) q.F) interface{} {
	return ((q.F)(me)).Eval(a1, a2)
}
func (me UserField) F() q.F                                { return ((q.F)(me)).F() }
func (me UserField) GreaterOrEqual(a1 interface{}) q.Query { return ((q.F)(me)).GreaterOrEqual(a1) }
func (me UserField) GreaterThan(a1 interface{}) q.Query    { return ((q.F)(me)).GreaterThan(a1) }
func (me UserField) In(a1 ...interface{}) q.Query          { return ((q.F)(me)).In(a1...) }
func (me UserField) LessOrEqual(a1 interface{}) q.Query    { return ((q.F)(me)).LessOrEqual(a1) }
func (me UserField) LessThan(a1 interface{}) q.Query       { return ((q.F)(me)).LessThan(a1) }
func (me UserField) Not() q.Query                          { return ((q.F)(me)).Not() }
func (me UserField) NotEqual(a1 interface{}) q.Query       { return ((q.F)(me)).NotEqual(a1) }
func (me UserField) NotIn(a1 ...interface{}) q.Query       { return ((q.F)(me)).NotIn(a1...) }
func (me UserField) StrLen(a1 ...interface{}) q.Operand    { return ((q.F)(me)).StrLen(a1...) }

type PostField q.F

const (
	PostId           PostField = "Id"
	PostCreated      PostField = "Created"
	postBy           PostField = "by"
	PostTo           PostField = "To"
	PostMd           PostField = "Md"
	PostFiles        PostField = "Files"
	PostRepl         PostField = "Repl"
	PostBy_Id        PostField = "by.Id"
	PostBy_Created   PostField = "by.Created"
	PostBy_LastSeen  PostField = "by.LastSeen"
	PostBy_Auth      PostField = "by.Auth"
	PostBy_PicFileId PostField = "by.PicFileId"
	PostBy_Nick      PostField = "by.Nick"
	PostBy_Btw       PostField = "by.Btw"
	PostBy_BtwDt     PostField = "by.BtwDt"
	PostBy_Buddies   PostField = "by.Buddies"
	PostRepl_Id      PostField = "Repl.Id"
	PostRepl_Created PostField = "Repl.Created"
	postRepl_by      PostField = "Repl.by"
	PostRepl_To      PostField = "Repl.To"
	PostRepl_Md      PostField = "Repl.Md"
	PostRepl_Files   PostField = "Repl.Files"
	PostRepl_Repl    PostField = "Repl.Repl"
)

func (me PostField) Asc() q.OrderBy               { return ((q.F)(me)).Asc() }
func (me PostField) Desc() q.OrderBy              { return ((q.F)(me)).Desc() }
func (me PostField) Equal(a1 interface{}) q.Query { return ((q.F)(me)).Equal(a1) }
func (me PostField) Eval(a1 interface{}, a2 func(q.C) q.F) interface{} {
	return ((q.F)(me)).Eval(a1, a2)
}
func (me PostField) F() q.F                                { return ((q.F)(me)).F() }
func (me PostField) GreaterOrEqual(a1 interface{}) q.Query { return ((q.F)(me)).GreaterOrEqual(a1) }
func (me PostField) GreaterThan(a1 interface{}) q.Query    { return ((q.F)(me)).GreaterThan(a1) }
func (me PostField) In(a1 ...interface{}) q.Query          { return ((q.F)(me)).In(a1...) }
func (me PostField) LessOrEqual(a1 interface{}) q.Query    { return ((q.F)(me)).LessOrEqual(a1) }
func (me PostField) LessThan(a1 interface{}) q.Query       { return ((q.F)(me)).LessThan(a1) }
func (me PostField) Not() q.Query                          { return ((q.F)(me)).Not() }
func (me PostField) NotEqual(a1 interface{}) q.Query       { return ((q.F)(me)).NotEqual(a1) }
func (me PostField) NotIn(a1 ...interface{}) q.Query       { return ((q.F)(me)).NotIn(a1...) }
func (me PostField) StrLen(a1 ...interface{}) q.Operand    { return ((q.F)(me)).StrLen(a1...) }
