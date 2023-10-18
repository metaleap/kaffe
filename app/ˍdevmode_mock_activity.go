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

const mockUsersNumTotal = 2342
const mockUsersNumActive = 123
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
		for i := 0; i < mockUsersNumTotal; i++ {
			mockEnsureUser(i)
		}
	}
}

func mockEnsureUser(i int) {
	user_email_addr := str.Fmt("foo%d@bar.baz", i)
	ctx := NewDebugNoCatch(time.Minute, user_email_addr)
	defer ctx.OnDone(nil)
	ctx.DbTx()

	user := UserByEmailAddr(ctx, user_email_addr)
	if user == nil { // not yet exists: create
		auth_id := yoauth.UserRegister(ctx, user_email_addr, "foobar")
		user = &User{}
		user.Auth.SetId(auth_id)
		// give new User a nickname and some btw
		for col, fld := range map[UserCol]*yodb.Text{UserColNickName: &user.NickName, UserColBtw: &user.Btw} {
			for *fld == "" {
				if *fld = yodb.Text(mockGetFortune(If(col == UserColBtw, 123, 23), true)); yodb.FindOne[User](ctx, col.Equal(*fld)) != nil {
					*fld = ""
				}
			}
		}
		// give new User some buddies
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

		if user.Id = yodb.CreateOne[User](ctx, user); user.Id <= 0 {
			panic(ErrDbNotStored)
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
