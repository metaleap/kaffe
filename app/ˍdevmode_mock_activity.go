//go:build debug

package haxsh

import (
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

const mockLiveActivity = true
const mockUsersNumTotal = 1444 // don't go higher than that due to limited number of `fortune`s (at nickname short length) for unique-nickname generation
const mockUsersNumActiveMin = mockUsersNumTotal / 2
const mockFilesDirPath = "__static/mockfiles"

var mockUsersNumMaxBuddies = 22 + rand.Intn(44)
var mockUserPicFiles = []string{"user0.png", "user1.jpg", "user2.png", "user3.jpg", "user4.png", "user5.jpg", "user6.png", "user7.jpg"}
var mockPostFiles = []string{"vid1.webm", "vid2.mp4", "vid3.mp4", "post1.jpg", "post10.png", "post11.jpg", "post12.jpg", "post13.png", "post14.jpg", "post15.jpg", "post16.png", "post17.png", "post18.png", "post19.jpg", "post2.jpg", "post20.png", "post21.webp", "post22.jpg", "post23.png", "post24.jpg", "post25.jpg", "post26.png", "post27.jpeg", "post28.jpg", "post29.jpg", "post3.jpg", "post30.jpg", "post31.webp", "post4.jpg", "post5.jpg", "post6.jpg", "post7.jpg", "post8.jpg", "post9.jpg"}
var mockUsersAllById = map[yodb.I64]string{}
var mockUsersAllByEmail = map[string]yodb.I64{}
var mockUsersLoggedIn = map[string]*http.Client{}

func init() {
	devModeInitMockUsers = func() {
		// ensure all users exist
		ids_so_far := make([]yodb.I64, 0, mockUsersNumTotal)
		for i := 0; i < mockUsersNumTotal; i++ {
			ids_so_far = append(ids_so_far, mockEnsureUser(i, ids_so_far))
		}

		// initiate some goroutines that regularly fake some action or other
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

var mockLock sync.Mutex
var mockActions = []string{ // don't reorder items with consulting/adapting the below `mockSomeActivity` func
	"logInOrOut",
	"changeNick",
	"changeBtw",
	"changePic",
	"changeBuddy",
	"postSomething",
}

func mockSomeActivity() {
	defer time.AfterFunc(time.Millisecond*time.Duration(111+rand.Intn(1111)), mockSomeActivity)
	// we do about 1-3 dozen reqs per sec with the above and the `rand`ed goroutining of this func set up in `init`

	action := mockActions[len(mockActions)-1] // default to the much-more-frequent-than-the-others-by-design action...
	if rand.Intn(len(mockActions)) == 0 {     // ...except there's a 1-in-n chance for another action
		action = mockActions[rand.Intn(len(mockActions))]
	}

	user_email_addr := str.Fmt("foo%d@bar.baz", rand.Intn(mockUsersNumTotal))
	mockLock.Lock()
	if (len(mockUsersLoggedIn) < mockUsersNumActiveMin) && (mockUsersLoggedIn[user_email_addr] == nil) {
		action = mockActions[0]
	}
	mockLock.Unlock()

	ctx := NewCtxNonHttp(time.Minute, user_email_addr+" "+action)
	defer ctx.OnDone(nil)
	ctx.DbTx()
	ctx.TimingsNoPrintInDevMode = true

	user := UserByEmailAddr(ctx, user_email_addr)
	if user.Nick == "" {
		action = mockActions[1]
	}
	switch _ = user; action {
	case "logInOrOut":
		mockLock.Lock()
		if mockUsersLoggedIn[user_email_addr] == nil {
			mockUsersLoggedIn[user_email_addr] = newClient()
		} else {
			delete(mockUsersLoggedIn, user_email_addr)
		}
		mockLock.Unlock()
	case "changeBtw":
		mockUpdEnsureChange(&user.Btw, func() yodb.Text { return yodb.Text(mockGetFortune(123, false)) }, nil)
		if rand.Intn(22) == 0 {
			user.Btw = ""
		}
		if upd := (&User{Id: user.Id, Auth: user.Auth, Btw: user.Btw}); !UserUpdate(ctx, upd, false) {
			panic(str.From(upd))
		}
	case "changeNick":
		mockUpdEnsureChange(&user.Nick, func() yodb.Text {
			var one, two string
			for one == "" || two == "" || one == two {
				one, two = mockGetFortune(11+rand.Intn(11), true), mockGetFortune(11+rand.Intn(11), true)
			}
			return yodb.Text(If(rand.Intn(2) == 0, one+two, two+one))
		}, func(it yodb.Text) bool {
			return !yodb.Exists[User](ctx, UserColNick.Equal(it))
		})
		_ = UserUpdate(ctx, &User{Id: user.Id, Auth: user.Auth, Nick: user.Nick}, false)
	case "changePic":
		mockUpdEnsureChange(&user.PicFileId, func() yodb.Text { return yodb.Text(mockUserPicFiles[rand.Intn(len(mockUserPicFiles))]) }, nil)
		if upd := (&User{Id: user.Id, Auth: user.Auth, PicFileId: user.PicFileId}); !UserUpdate(ctx, upd, false) {
			panic(str.From(upd))
		}
	case "changeBuddy":
		mockSomeActivityChangeBuddy(ctx, user, user_email_addr)
		_ = UserUpdate(ctx, &User{Id: user.Id, Auth: user.Auth, Buddies: user.Buddies}, false)
	case "postSomething":
		mockSomeActivityPostSomething(ctx, user)
	default:
		panic(action)
	}
}

func mockSomeActivityChangeBuddy(ctx *Ctx, user *User, userEmailAddr string) {
	if add_or_remove := rand.Intn(3); ((add_or_remove == 0) || (len(user.Buddies) > mockUsersNumMaxBuddies)) && (len(user.Buddies) > 0) {
		user.Buddies = sl.WithoutIdx(user.Buddies, rand.Intn(len(user.Buddies)), true) // remove a buddy
	} else { // add a buddy
		var buddy_email_addr string
		var buddy_id yodb.I64
		for (buddy_id == 0) || (buddy_id == user.Id) || sl.Has(user.Buddies, buddy_id) || (buddy_email_addr == "") || (buddy_email_addr == userEmailAddr) {
			if buddy_email_addr = str.Fmt("foo%d@bar.baz", rand.Intn(mockUsersNumTotal)); buddy_email_addr != userEmailAddr {
				buddy_id = UserByEmailAddr(ctx, buddy_email_addr).Id
			}
		}
		user.Buddies = append(user.Buddies, buddy_id)
	}
}

func mockSomeActivityPostSomething(ctx *Ctx, user *User) {
	var files []FileRef
	var to []yodb.I64
	var in_reply_to yodb.I64
	md := mockGetFortune(0, false)

	// add one or more files?
	if rand.Intn(11) <= 2 {
		md = ""
	} // separate rands because can have file-only posts as well as text+file/s posts
	if (md == "") || (rand.Intn(11) <= 2) {
		for i := 0; i < rand.Intn(11); i++ {
			file_name := mockPostFiles[rand.Intn(len(mockPostFiles))]
			files = append(files, FileRef{Id: file_name, Name: file_name})
		}
	}

	// in reply to some other post?
	if rand.Intn(11) <= 2 {
		if post := yodb.FindOne[Post](ctx, PostColBy.In(user.Buddies.ToAnys()...).And(PostColRepl.Equal(""))); post != nil {
			in_reply_to = post.Id
		}
	}

	if rand.Intn(11) <= 3 {
		for to = make([]yodb.I64, 0, 1+rand.Intn(len(user.Buddies)-2)); len(to) < cap(to); {
			if buddy_id := user.Buddies[rand.Intn(len(user.Buddies))]; !sl.Has(to, buddy_id) {
				to = append(to, buddy_id)
			}
		}
	}

	UserPost(ctx, user, md, in_reply_to, files, to)
}

func mockUpdEnsureChange[T comparable](at *T, getAnother func() T, ok func(T) bool) {
	orig := *at
	for (*at) == orig || ((ok != nil) && !ok(*at)) {
		*at = getAnother()
	}
}

func mockEnsureUser(i int, idsSoFar []yodb.I64) yodb.I64 {
	user_email_addr := str.Fmt("foo%d@bar.baz", i)
	ctx := NewCtxNonHttp(time.Minute, user_email_addr)
	defer ctx.OnDone(nil)
	ctx.DbTx()

	ctx.Timings.Step("check exists")
	user := UserByEmailAddr(ctx, user_email_addr)
	if user == nil { // not yet exists: create
		ctx.Timings.Step("register new auth")
		auth_id := yoauth.UserRegister(ctx, user_email_addr, "foobar")
		user = &User{Nick: yodb.Text(user_email_addr[:str.Idx(string(user_email_addr), '@')])}
		user.Auth.SetId(auth_id)

		ctx.Timings.Step("insert new user")
		if user.Id = yodb.CreateOne[User](ctx, user); user.Id <= 0 {
			panic(ErrDbNotStored)
		}
	}
	mockUsersAllByEmail[user_email_addr] = user.Id
	mockUsersAllById[user.Id] = user_email_addr
	return user.Id
}

func mockGetFortune(maxLen int, ident bool) (ret string) {
	allow_multi_line, did_alt := (maxLen <= 0), false
	for (ret == "") || ((!allow_multi_line) && (str.Idx(ret, '\n') >= 0)) || str.IsUp(ret) {
		var args []string
		if maxLen > 0 {
			args = append(args, "-n", str.FromInt(maxLen), "-s")
		}
		if did_alt = ((maxLen >= 77) || (maxLen <= 0)) && (rand.Intn(If(maxLen <= 0, 3, 2)) != 0); did_alt {
			args = append(args, filepath.Join(mockFilesDirPath, "fortune_showerthoughts.txt"))
		}
		cmd := exec.Command("fortune", args...)
		output, err := cmd.CombinedOutput()
		if ret = string(output); err != nil {
			panic(err)
		}
		if idx := str.IdxRune(ret, '―'); idx >= 0 {
			ret = ret[:idx]
		}
		if ret = str.Trim(ret); ident {
			ret = str.Up0(ToIdentWith(ret, 0))
		}
	}
	return
}
