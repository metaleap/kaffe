//go:build debug

package haxsh

import (
	"math/rand"
	"os/exec"
	"time"

	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

const numMockUsers = 1234
const mockFilesDirPath = "__static/mockfiles"

var mockUserPicFiles = []string{
	"user0.png",
	"user1.jpg",
	"user2.png",
	"user3.jpg",
	"user4.png",
	"user5.jpg",
	"user6.png",
	"user7.jpg",
}

var mockPostFiles = []string{
	"vid1.webm",
	"vid2.mp4",
	"vid3.mp4",
	"post1.jpg",
	"post10.png",
	"post11.jpg",
	"post12.jpg",
	"post13.png",
	"post14.jpg",
	"post15.jpg",
	"post16.png",
	"post17.png",
	"post18.png",
	"post19.jpg",
	"post2.jpg",
	"post20.png",
	"post21.webp",
	"post22.jpg",
	"post23.png",
	"post24.jpg",
	"post25.jpg",
	"post26.png",
	"post27.jpeg",
	"post28.jpg",
	"post29.jpg",
	"post3.jpg",
	"post30.jpg",
	"post31.webp",
	"post4.jpg",
	"post5.jpg",
	"post6.jpg",
	"post7.jpg",
	"post8.jpg",
	"post9.jpg",
}

func init() {
	devModeInitMockUsers = func() {
		for i := 0; i < numMockUsers; i++ {
			mockEnsureUser(i)
		}
	}
}

func mockEnsureUser(i int) {
	user_email_addr := str.Fmt("foo%d@bar.baz", i)
	ctx := NewDebugNoCatch(time.Minute, user_email_addr)
	defer ctx.OnDone(nil)
	ctx.DbTx()

	is_every_11th, is_every_7th := ((i % 11) == 0), ((i % 7) == 0)
	user := UserByEmailAddr(ctx, user_email_addr)
	if user == nil { // not yet exists: create
		auth_id := yoauth.UserRegister(ctx, user_email_addr, "foobar")
		user = &User{NickName: If(is_every_11th, "", yodb.Text(user_email_addr[:str.Idx(user_email_addr, '@')]))}
		user.Auth.SetId(auth_id)
		if user.Id = yodb.CreateOne[User](ctx, user); user.Id <= 0 {
			panic(ErrDbNotStored)
		}
	}

	// give a new nickname
	for (!is_every_11th) && (user.NickName == "") {
		if user.NickName = yodb.Text(mockGetFortune(22, true)); yodb.FindOne[User](ctx, UserColNickName.Equal(user.NickName)) != nil {
			user.NickName = ""
		}
	}
	if is_every_11th {
		user.NickName = ""
	}

	if !is_every_7th { // give a new btw
		for old_btw := user.Btw; (user.Btw == "") || (user.Btw == old_btw); user.Btw.Do(str.Trim) {
			user.Btw = yodb.Text(mockGetFortune(88, false))
		}
	}

	if len(user.Buddies) == 0 { // give user some buddies
		num_buddies := 3 + rand.Intn(22)
		for i := 0; i < num_buddies; i++ {
			var buddy *User
			for buddy_id := yodb.I64(0); buddy == nil; buddy_id = 0 {
				for (buddy_id == 0) || (buddy_id == user.Id) || sl.Has(user.Buddies, buddy_id) {
					buddy_id = yodb.I64(4 + rand.Intn(1234))
				}
				buddy = yodb.FindOne[User](ctx, UserColId.Equal(buddy_id))
			}
			user.Buddies = append(user.Buddies, buddy.Id)
		}
	}
}

func mockGetFortune(maxLen int, ident bool) (ret string) {
	allow_multi_line := (maxLen <= 0)
	for (ret == "") || ((!allow_multi_line) && (str.Idx(ret, '\n') >= 0)) || str.IsUp(ret) {
		cmd := exec.Command("fortune", "-n", str.FromInt(maxLen), "-s")
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		ret = str.Trim(string(output))
	}
	if ident {
		ret = str.Up0(ToIdentWith(ret, 0))
	}
	return
}
