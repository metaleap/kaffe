//go:build debug

package haxsh

import (
	"math/rand"
	"net/http"
	"os/exec"
	"time"

	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

const mockLiveActivity = true
const mockUsersNumTotal = 1448 // don't go higher than that due to limited number of `fortune`s (at nickname short length) for unique-nickname generation
const mockUsersNumActiveMin = mockUsersNumTotal / 2
const mockFilesDirPath = "__static/mockfiles"

var mockUserPicFiles = []string{"user0.png", "user1.jpg", "user2.png", "user3.jpg", "user4.png", "user5.jpg", "user6.png", "user7.jpg"}
var mockPostFiles = []string{"vid1.webm", "vid2.mp4", "vid3.mp4", "post1.jpg", "post10.png", "post11.jpg", "post12.jpg", "post13.png", "post14.jpg", "post15.jpg", "post16.png", "post17.png", "post18.png", "post19.jpg", "post2.jpg", "post20.png", "post21.webp", "post22.jpg", "post23.png", "post24.jpg", "post25.jpg", "post26.png", "post27.jpeg", "post28.jpg", "post29.jpg", "post3.jpg", "post30.jpg", "post31.webp", "post4.jpg", "post5.jpg", "post6.jpg", "post7.jpg", "post8.jpg", "post9.jpg"}
var mockUsersAllById = map[yodb.I64]string{}
var mockUsersAllByEmail = map[string]yodb.I64{}
var mockUsersLoggedIn = map[string]*http.Client{}

func init() {
	devModeInitMockUsers = func() {
		for i := 0; i < mockUsersNumTotal; i++ {
			mockEnsureUser(i)
		}
		if mockLiveActivity {
			next_email_addr := func() string { return str.Fmt("foo%d@bar.baz", rand.Intn(mockUsersNumTotal)) }
			// mock-login at least the min number of some random users
			for len(mockUsersLoggedIn) < mockUsersNumActiveMin {
				user_email_addr := next_email_addr()
				for mockUsersLoggedIn[user_email_addr] != nil {
					user_email_addr = next_email_addr()
				}
				mockUsersLoggedIn[user_email_addr] = newClient()
			}
			for i, num_parallel := 0, 11+rand.Intn(11); i < num_parallel; i++ {
				time.AfterFunc(time.Second*time.Duration(i), mockSomeActivity)
			}
		}
	}
}

func newClient() *http.Client { return &http.Client{Timeout: time.Second} }

var mockActions = []Pair[string, func(*Ctx, *User)]{
	{"logInOrOut", nil}, // this must be at index 0, see `mockSomeActivity`
	{"changeBtw", nil},
	{"changeNick", nil},
	{"changePic", nil},
	{"postSomething", nil},
	{"changeBuddy", nil},
}

func mockSomeActivity() {
	defer time.AfterFunc(time.Millisecond*time.Duration(111+rand.Intn(1111)), mockSomeActivity)

	user_email_addr := str.Fmt("foo%d@bar.baz", rand.Intn(mockUsersNumTotal))
	do := mockActions[rand.Intn(len(mockActions))]
	is_logged_in := mockUsersLoggedIn[user_email_addr]
	if (len(mockUsersLoggedIn) < mockUsersNumActiveMin) && (is_logged_in == nil) {
		do = mockActions[0]
	}

	ctx := NewCtxNonHttp(time.Minute, user_email_addr+" "+do.Key)
	defer ctx.OnDone(nil)
	ctx.DbTx()

	user := UserByEmailAddr(ctx, user_email_addr)
	switch _ = user; do.Key {
	case "logInOrOut":
		if is_logged_in == nil {
			mockUsersLoggedIn[user_email_addr] = newClient()
		} else {
			delete(mockUsersLoggedIn, user_email_addr)
		}
	case "changeBtw":
	case "changeNick":
	case "changePic":
	case "postSomething":
	case "changeBuddy":
	default:
		panic(do.Key)
	}
}

func mockEnsureUser(i int) {
	user_email_addr := str.Fmt("foo%d@bar.baz", i)
	ctx := NewCtxNonHttp(time.Minute, user_email_addr)
	defer ctx.OnDone(nil)
	ctx.DbTx()

	ctx.Timings.Step("check exists")
	user := UserByEmailAddr(ctx, user_email_addr)
	if user == nil { // not yet exists: create
		ctx.Timings.Step("init new user")
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
		num_buddies := rand.Intn(23)
		for len(user.Buddies) < num_buddies {
			var buddy *User
			for buddy_id := yodb.I64(0); buddy == nil; buddy_id = 0 {
				for (buddy_id == 0) || (buddy_id == user.Id) || sl.Has(user.Buddies, buddy_id) {
					buddy_id = yodb.I64(4 + rand.Intn(1234))
				}
				buddy = yodb.FindOne[User](ctx, UserColId.Equal(buddy_id))
			}
			user.Buddies = append(user.Buddies, buddy.Id)
		}

		ctx.Timings.Step("insert new user")
		if user.Id = yodb.CreateOne[User](ctx, user); user.Id <= 0 {
			panic(ErrDbNotStored)
		}
	}
	mockUsersAllByEmail[user_email_addr] = user.Id
	mockUsersAllById[user.Id] = user_email_addr
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
