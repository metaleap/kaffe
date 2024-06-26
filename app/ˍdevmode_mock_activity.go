//go:build debug

package kaffe

import (
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
	"yo"

	yoauth "yo/auth"
	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	. "yo/srv"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

var mockLiveActivity = true

const mockNumReqsPerSecApprox = 11
const mockUsersNumTotal = 12345

var mockUsersNumMaxBuddies = 22 + rand.Intn(22)
var mockUserPicFiles = []string{"user0.png", "user1.jpg", "user2.png", "user3.jpg", "user4.png", "user5.jpg", "user6.png", "user7.jpg"}
var mockPostFiles = []string{"vid1.webm", "vid2.mp4", "vid3.mp4", "post1.jpg", "post10.png", "post11.jpg", "post12.jpg", "post13.png", "post14.jpg", "post15.jpg", "post16.png", "post17.png", "post18.png", "post19.jpg", "post2.jpg", "post20.png", "post21.webp", "post22.jpg", "post23.png", "post24.jpg", "post25.jpg", "post26.png", "post27.jpeg", "post28.jpg", "post29.jpg", "post3.jpg", "post30.jpg", "post31.webp", "post4.jpg", "post5.jpg", "post6.jpg", "post7.jpg", "post8.jpg", "post9.jpg"}
var mockUsersAllById = map[yodb.I64]string{}
var mockUsersAllByEmail = map[string]yodb.I64{}
var mockUsersLoggedIn = map[string]*http.Client{}
var mockUsersNever = map[string]bool{
	"foo1@bar.baz":      true,
	"foo123@bar.baz":    true,
	"foo234@bar.baz":    true,
	"foo321@bar.baz":    true,
	elizaUser.emailAddr: true,
}

func mockFilesDirPath() string { return Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"] }

func init() {
	yo.AppSideBuildTimeContainerFileNames = append(yo.AppSideBuildTimeContainerFileNames, elizaUser.picFileName)

	if !mockLiveActivity {
		return
	}
	devModeInitMockUsers = func() {
		// ensure all users exist
		for i := 2; /*mock user eliza already exists, necessarily always under id 1*/ i <= mockUsersNumTotal; i++ {
			mockEnsureUser(i)
		}

		// initiate some goroutines that regularly fake some action or other
		if mockLiveActivity {
			for i := 0; i < mockNumReqsPerSecApprox; i++ {
				time.AfterFunc(time.Second+(time.Duration(i)*(time.Second/mockNumReqsPerSecApprox)), mockSomeActivity)
			}
		}
	}
}

var mockLock sync.Mutex
var mockActions = []string{ // don't reorder items with consulting/adapting the below `mockSomeActivity` func
	"postSomething",
	"changeBuddy",
	"changeNick",
	"changeBtw",
	"changePic",
}
var busy = map[string]bool{}

func mockSomeActivity() {
	if !mockLiveActivity { // for turning off on-the-fly during debugging
		return
	}

	const sec_half = time.Second / 2
	defer time.AfterFunc(sec_half+time.Duration(rand.Intn(int(2*sec_half))), mockSomeActivity)

	action := mockActions[0]              // default to the much-more-frequent-than-the-others-by-design action...
	if rand.Intn(len(mockActions)) == 0 { // ...except there's still a (just much-lower) chance for another action
		action = mockActions[rand.Intn(len(mockActions))]
	}
	var user_email_addr string
	var user_client *http.Client
	var must_log_in_first bool
	{
		mockLock.Lock()
		for (user_email_addr == "") || busy[user_email_addr] || mockUsersNever[user_email_addr] {
			user_email_addr = str.Fmt("foo%d@bar.baz", 1+rand.Intn(mockUsersNumTotal))
		}
		user_client = mockUsersLoggedIn[user_email_addr]
		busy[user_email_addr] = true
		if must_log_in_first = (user_client == nil) || (0 == len(user_client.Jar.Cookies(nil))); must_log_in_first {
			user_client = NewClient()
			mockUsersLoggedIn[user_email_addr] = user_client
		}
		defer func() { mockLock.Lock(); busy[user_email_addr] = false; mockLock.Unlock() }()
		mockLock.Unlock()
	}

	ctx := NewCtxNonHttp(time.Minute, false, user_email_addr+" "+action)
	defer ctx.OnDone(nil)
	ctx.TimingsNoPrintInDevMode = true

	if must_log_in_first {
		ViaHttp[ApiUserSignInOrReset, None](apiUserSignInOrReset, ctx, &ApiUserSignInOrReset{
			ApiNickOrEmailAddr: ApiNickOrEmailAddr{NickOrEmailAddr: user_email_addr}, PasswordPlain: "foobar",
		}, user_client)
	}

	do_update := func(curUser *User, upd *User, changedFields ...UserField) {
		upd.Id = curUser.Id
		upd.Account.SetId(curUser.Account.Id())
		ViaHttp[yodb.ApiUpdateArgs[User, UserField], None](apiUserUpdate, ctx, &yodb.ApiUpdateArgs[User, UserField]{
			Changes:       *upd,
			Id:            curUser.Id,
			ChangedFields: changedFields,
		}, user_client)
	}

	ctx.DbTx(false)
	user := userByEmailAddr(ctx, user_email_addr)
	if user == nil {
		panic("how come user email addr '" + user_email_addr + "' gone?!")
	}
	switch _ = user; action {
	case "changeBtw":
		orig := user.Btw
		mockUpdEnsureChange(&user.Btw, func() yodb.Text { return yodb.Text(mockGetFortune(123, false)) }, nil)
		if (orig != "") && (rand.Intn(22) == 0) {
			user.Btw = ""
		}
		do_update(user, &User{Btw: user.Btw}, UserBtw)
	case "changePic":
		orig := user.PicFileId
		mockUpdEnsureChange(&user.PicFileId, func() yodb.Text { return yodb.Text(mockUserPicFiles[rand.Intn(len(mockUserPicFiles))]) }, nil)
		if (orig != "") && (rand.Intn(11) == 0) {
			user.PicFileId = ""
		}
		do_update(user, &User{PicFileId: user.PicFileId}, UserPicFileId)
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
		do_update(user, &User{Nick: user.Nick}, UserNick)
	case "changeBuddy":
		mockSomeActivityChangeBuddy(ctx, user, user_email_addr)
		do_update(user, &User{Buddies: user.Buddies}, UserBuddies)
	case "postSomething":
		mockSomeActivityPostSomething(ctx, user, user_client)
	default:
		panic(action)
	}
}

func mockSomeActivityChangeBuddy(ctx *Ctx, user *User, userEmailAddr string) {
	if add_or_remove := rand.Intn(11); ((add_or_remove == 0) || (len(user.Buddies) > mockUsersNumMaxBuddies)) && (len(user.Buddies) > 0) {
		user.Buddies = sl.WithoutIdx(user.Buddies, rand.Intn(len(user.Buddies)), true) // remove a buddy
	} else { // add a buddy
		var buddy_id yodb.I64
		for (buddy_id == 0) || (buddy_id == user.Id) || sl.Has(user.Buddies, buddy_id) {
			buddy_id = yodb.I64(1 + rand.Intn(mockUsersNumTotal))
		}
		user.Buddies = append(user.Buddies, buddy_id)
	}
}

func mockSomeActivityPostSomething(ctx *Ctx, user *User, client *http.Client) {
	md := mockGetFortune(0, false)
	new_post := &Post{Htm: yodb.Text(md)}
	new_post.By.SetId(user.Id)
	ViaHttp[PostNew, Return[yodb.I64]](apiPostNew, ctx, &PostNew{NewPost: new_post}, client)
}

func mockUpdEnsureChange[T comparable](at *T, getAnother func() T, ok func(T) bool) {
	orig := *at
	for ((*at) == orig) || ((ok != nil) && !ok(*at)) {
		*at = getAnother()
	}
}

func mockEnsureUser(i int) yodb.I64 {
	user_email_addr := str.Fmt("foo%d@bar.baz", i)
	ctx := NewCtxNonHttp(time.Minute, false, user_email_addr)
	defer ctx.OnDone(nil)
	ctx.DbTx(false)
	ctx.TimingsNoPrintInDevMode = true

	ctx.Timings.Step("check exists")
	user := userByEmailAddr(ctx, user_email_addr)
	if user == nil { // not yet exists: create
		ctx.TimingsNoPrintInDevMode = false
		ctx.Timings.Step("register new auth")
		account_id := yoauth.UserRegister(ctx, user_email_addr, "foobar")
		user = &User{Nick: yodb.Text(user_email_addr[:str.Idx(string(user_email_addr), '@')]), byBuddyDtLastMsgCheck: yodb.JsonMap[*yodb.DateTime]{}}
		switch i {
		case 123:
			user.Buddies = yodb.Arr[yodb.I64]{234, 321}
		case 234:
			user.Buddies = yodb.Arr[yodb.I64]{123, 321}
		case 321:
			user.Buddies = yodb.Arr[yodb.I64]{123, 234}
		}
		user.Account.SetId(account_id)
		user.LastSeen = yodb.DtNow()

		ctx.Timings.Step("insert new user")
		user.gravatarChecked = true
		user.Id = yodb.CreateOne[User](ctx, user)
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
			args = append(args, filepath.Join(mockFilesDirPath(), "fortune_showerthoughts.txt"))
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
