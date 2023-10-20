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
	q "yo/db/query"
	yoauth "yo/feat_auth"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

var mockLiveActivity = true

const mockUsersNumTotal = 1444 // don't go higher than that due to limited number of `fortune`s (at nickname short length) for unique-nickname generation
const mockFilesDirPath = "__static/mockfiles"

var mockUsersNumMaxBuddies = 11 + rand.Intn(22)
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
			for i, num_parallel := 0, 11+rand.Intn(11); i < num_parallel; i++ {
				time.AfterFunc(time.Second*time.Duration(i), mockSomeActivity)
			}
		}
	}
}

var mockLock sync.Mutex
var mockActions = []string{ // don't reorder items with consulting/adapting the below `mockSomeActivity` func
	"postSomething",
	"changeNick",
	"changeBtw",
	"changePic",
	"changeBuddy",
}
var busy = map[string]bool{}

func mockSomeActivity() {
	if !mockLiveActivity { // for turning off live during debugging
		return
	}
	defer time.AfterFunc(time.Millisecond*time.Duration(111+rand.Intn(1111)), mockSomeActivity)
	// we do about 1-3 dozen reqs per sec with the above and the `rand`ed goroutining of this func set up in `init`

	action := mockActions[0]                // default to the much-more-frequent-than-the-others-by-design action...
	if rand.Intn(2*len(mockActions)) <= 1 { // ...except there's still a (just much-lower) chance for another action
		action = mockActions[rand.Intn(len(mockActions))]
	}
	var user_email_addr string
	var user_client *http.Client
	var must_log_in_first bool
	{
		mockLock.Lock()
		for (user_email_addr == "") || busy[user_email_addr] {
			user_email_addr = str.Fmt("foo%d@bar.baz", rand.Intn(mockUsersNumTotal))
		}
		user_client = mockUsersLoggedIn[user_email_addr]
		busy[user_email_addr] = true
		if must_log_in_first = (user_client == nil); must_log_in_first {
			user_client = NewClient()
			mockUsersLoggedIn[user_email_addr] = user_client
		}
		defer func() { mockLock.Lock(); busy[user_email_addr] = false; mockLock.Unlock() }()
		mockLock.Unlock()
	}

	ctx := NewCtxNonHttp(time.Minute, user_email_addr+" "+action)
	defer ctx.OnDone(nil)
	ctx.DbTx()
	ctx.TimingsNoPrintInDevMode = true

	if must_log_in_first {
		ViaHttp[yoauth.ApiAccountPayload, Void](apiUserSignIn, ctx, &yoauth.ApiAccountPayload{
			EmailAddr: user_email_addr, PasswordPlain: "foobar",
		}, user_client)
	}

	do_update := func(upd *User, changedFields ...UserField) {
		userUpdate(ctx, upd, true, changedFields...)
		ViaHttp[yodb.ApiUpdateArgs[User, UserField], Void](apiUserUpdate, ctx, &yodb.ApiUpdateArgs[User, UserField]{
			Changes:       *upd,
			Id:            upd.Id,
			ChangedFields: changedFields,
		}, user_client)
	}

	user := userByEmailAddr(ctx, user_email_addr)
	switch _ = user; action {
	case "changeBtw":
		mockUpdEnsureChange(&user.Btw, func() yodb.Text { return yodb.Text(mockGetFortune(123, false)) }, nil)
		if rand.Intn(22) == 0 {
			user.Btw = ""
		}
		do_update(&User{Id: user.Id, Btw: user.Btw}, UserBtw)
	case "changeNick":
		mockUpdEnsureChange(&user.Nick, func() yodb.Text {
			var one, two string
			for (one == "") || (two == "") || (one == two) {
				one, two = mockGetFortune(11+rand.Intn(11), true), mockGetFortune(11+rand.Intn(11), true)
			}
			return yodb.Text(If(rand.Intn(2) == 0, one+two, two+one))
		}, func(it yodb.Text) bool {
			return !yodb.Exists[User](ctx, UserNick.Equal(it))
		})
		do_update(&User{Id: user.Id, Nick: user.Nick}, UserNick)
	case "changePic":
		mockUpdEnsureChange(&user.PicFileId, func() yodb.Text { return yodb.Text(mockUserPicFiles[rand.Intn(len(mockUserPicFiles))]) }, nil)
		if rand.Intn(22) == 0 {
			user.PicFileId = ""
		}
		do_update(&User{Id: user.Id, PicFileId: user.PicFileId}, UserPicFileId)
	case "changeBuddy":
		mockSomeActivityChangeBuddy(ctx, user, user_email_addr)
		do_update(&User{Id: user.Id, Buddies: user.Buddies}, UserBuddies)
	case "postSomething":
		mockSomeActivityPostSomething(ctx, user, user_client)
	default:
		panic(action)
	}
}

func mockSomeActivityChangeBuddy(ctx *Ctx, user *User, userEmailAddr string) {
	if add_or_remove := rand.Intn(2); ((add_or_remove == 0) || (len(user.Buddies) > mockUsersNumMaxBuddies)) && (len(user.Buddies) > 0) {
		user.Buddies = sl.WithoutIdx(user.Buddies, rand.Intn(len(user.Buddies)), true) // remove a buddy
	} else { // add a buddy
		var buddy_email_addr string
		var buddy_id yodb.I64
		for (buddy_id == 0) || (buddy_id == user.Id) || sl.Has(user.Buddies, buddy_id) || (buddy_email_addr == "") || (buddy_email_addr == userEmailAddr) {
			if buddy_email_addr = str.Fmt("foo%d@bar.baz", rand.Intn(mockUsersNumTotal)); buddy_email_addr != userEmailAddr {
				buddy_id = userByEmailAddr(ctx, buddy_email_addr).Id
			}
		}
		user.Buddies = append(user.Buddies, buddy_id)
	}
}

func mockSomeActivityPostSomething(ctx *Ctx, user *User, client *http.Client) {
	var files []FileRef
	var to []yodb.I64
	var in_reply_to yodb.I64
	md := mockGetFortune(0, false)

	// add one or more files?
	if rand.Intn(11) <= 2 {
		md = ""
	} // separate rands because can have file-only posts as well as text+file/s posts
	if (md == "") || (rand.Intn(11) <= 2) {
		num_files := If(rand.Intn(2) == 0, 1, 1+rand.Intn(11))
		for i := 0; i < num_files; i++ {
			var file_name string
			for (file_name == "") || sl.Has(files, FileRef(file_name)) {
				file_name = mockPostFiles[rand.Intn(len(mockPostFiles))]
			}
			files = append(files, FileRef(file_name))
		}
	}

	// addressing only some not all?
	if max := len(user.Buddies) - 1; (rand.Intn(11) <= 4) && (max > 1) {
		for to = make([]yodb.I64, 0, 1+rand.Intn(max-1)); len(to) < cap(to); {
			if buddy_id := user.Buddies[rand.Intn(len(user.Buddies))]; !sl.Has(to, buddy_id) {
				to = append(to, buddy_id)
			}
		}
	}

	// in reply to some other post? (if so, changes `to` to NULL but apis/ux make it then effectively that post's `to`)
	if (rand.Intn(11) <= 3) && len(user.Buddies) > 0 {
		if post := yodb.FindOne[Post](ctx,
			PostRepl.Equal(nil).And(postBy.In(user.Buddies.Anys()...).And(
				PostTo.Equal(nil).Or(q.JsonHas(PostTo, q.That(user.Id))),
			)),
		); post != nil {
			to, in_reply_to = nil, post.Id
		}
	}

	new_post := &Post{Md: yodb.Text(md), Files: files, To: to}
	new_post.by.SetId(user.Id)
	new_post.Repl.SetId(in_reply_to)
	ViaHttp[Post, Return[yodb.I64]](apiPostNew, ctx, new_post, client)
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
	ctx.TimingsNoPrintInDevMode = true

	ctx.Timings.Step("check exists")
	user := userByEmailAddr(ctx, user_email_addr)
	if user == nil { // not yet exists: create
		ctx.TimingsNoPrintInDevMode = false
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
			ret = str.Replace(ret, str.Dict{"'": "", "`": "", "´": ""})
			ret = str.Up0(ToIdentWith(ret, 0))
		}
	}
	return
}
