package kaffe

import (
	"math/rand"
	"path/filepath"
	"time"

	yoauth "yo/auth"
	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	yojobs "yo/jobs"
	. "yo/util"
	"yo/util/sl"

	"github.com/NortySpock/eliza-go/eliza"
)

var elizaUser = struct {
	id          yodb.I64
	nick        string
	emailAddr   string
	picFileName string
	btw         string
}{0, "eliza", "eliza@metaleap.net", "eliza.png", "I'm Weizenbaum's O.G. chatbot (Go impl by mattshiel & NortySpock & kennysong)"}

func elizaAddBuddyBack(userNick string) {
	DoAfter(time.Second*time.Duration(3+rand.Intn(4)), func() {
		ctx := NewCtxNonHttp(time.Minute, false, "")
		defer ctx.OnDone(nil)
		user_eliza := userById(ctx, elizaUser.id)
		userAddBuddy(ctx, user_eliza, userNick)
	})
}

func elizaReplyShortlyTo(postId yodb.I64) {
	DoAfter(time.Second*time.Duration(1+rand.Intn(1)), func() {
		ctx := NewCtxNonHttp(time.Minute, false, "")
		defer ctx.OnDone(nil)
		post := yodb.ById[Post](ctx, postId)
		if post == nil { // already deleted
			return
		}

		last_reply := yodb.FindOne[Post](ctx, PostBy.Equal(elizaUser.id).And(q.ArrAreAny(PostTo, q.OpEq, post.By.Id())), PostDtMade.Desc())
		reply_text := eliza.ReplyTo(string(post.Htm))
		for (last_reply != nil) && (reply_text == last_reply.Htm.String()) {
			reply_text = eliza.ReplyTo(string(post.Htm))
		}
		postNew(ctx, &Post{
			To:  yodb.Arr[yodb.I64]{post.By.Id()},
			Htm: yodb.Text(reply_text),
		}, elizaUser.id)

		eliza_user := yodb.ById[User](ctx, elizaUser.id)
		eliza_user.LastSeen = yodb.DtNow()
		yodb.Update[User](ctx, eliza_user, nil, false, UserFields(UserLastSeen)...)
	})
}

func elizaEnsureUser() {
	if avatar_file_path := filepath.Join(Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"], elizaUser.picFileName); !FsIsFile(avatar_file_path) {
		FsCopy(If(IsDevMode, elizaUser.picFileName, "/"+elizaUser.picFileName), avatar_file_path)
	}

	ctx := NewCtxNonHttp(yojobs.Timeout1Min, false, "")
	defer ctx.OnDone(nil)
	if user_account := yoauth.ByEmailAddr(ctx, elizaUser.emailAddr); user_account != nil {
		elizaUser.id = yodb.FindOne[User](ctx, UserAccount.Equal(user_account.Id)).Id
	} else {
		ctx.DbTx(true)
		account_id := yodb.CreateOne[yoauth.UserAccount](ctx, &yoauth.UserAccount{
			EmailAddr: yodb.Text(elizaUser.emailAddr),
		})
		user := &User{
			LastSeen:        yodb.DtNow(),
			PicFileId:       yodb.Text(elizaUser.picFileName),
			Nick:            yodb.Text(elizaUser.nick),
			Btw:             yodb.Text(elizaUser.btw),
			gravatarChecked: true,
		}
		user.Account.SetId(account_id)
		elizaUser.id = yodb.CreateOne[User](ctx, user)
	}

	time.AfterFunc(time.Minute, elizaEnsureBuddies)
}

func elizaEnsureBuddies() {
	// usually, anyone adding eliza as buddy gets added back in `elizaAddBuddyBack`, but with a few secs artificial delay.
	// if server restarted before that delay elapsed, this infrequent worker will remedy such rare fluke situations.
	defer time.AfterFunc(time.Hour, elizaEnsureBuddies)

	ctx := NewCtxNonHttp(time.Minute, false, "")
	defer ctx.OnDone(nil)

	eliza_user := yodb.ById[User](ctx, elizaUser.id)
	user_query := q.ArrAreAny(UserBuddies, q.OpEq, elizaUser.id)
	if len(eliza_user.Buddies) > 0 {
		user_query = user_query.And(UserId.NotIn(eliza_user.Buddies.ToAnys()...))
	}
	buddy_requests := yodb.FindMany[User](ctx, user_query, 0, UserFields(UserId))
	for _, user := range buddy_requests {
		eliza_user.Buddies = sl.With(eliza_user.Buddies, user.Id)
	}
	if len(buddy_requests) > 0 {
		eliza_user.LastSeen = yodb.DtNow()
		yodb.Update[User](ctx, eliza_user, nil, false, UserFields(UserBuddies, UserLastSeen)...)
	}
}
