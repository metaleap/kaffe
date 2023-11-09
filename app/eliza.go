package kaffe

import (
	"math/rand"
	"path/filepath"
	"time"

	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
	yoauth "yo/feat_auth"
	yojobs "yo/jobs"
	. "yo/util"

	"github.com/NortySpock/eliza-go/eliza"
)

var elizaUser = struct {
	picFileName string
	id          yodb.I64
	nick        string
	emailAddr   string
	btw         string
}{"eliza.png", 0, "eliza", "eliza@metaleap.net", "I'm Weizenbaum's O.G. chatbot (Go impl by mattshiel & NortySpock & kennysong)"}

func elizaAddBuddy(userNick string) {
	DoAfter(time.Second*time.Duration(4+rand.Intn(7)), func() {
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

		last_reply := yodb.FindOne[Post](ctx, PostBy.Equal(elizaUser.id).And(q.ArrAreAnyIn(PostTo, q.OpEq, post.By.Id())), PostDtMade.Desc())
		reply_text := eliza.ReplyTo(string(post.Htm))
		for (last_reply != nil) && (reply_text == last_reply.Htm.String()) {
			reply_text = eliza.ReplyTo(string(post.Htm))
		}
		postNew(ctx, &Post{
			To:  yodb.Arr[yodb.I64]{post.By.Id()},
			Htm: yodb.Text(reply_text),
		}, elizaUser.id)
	})
}

func elizaEnsuerUser() {
	if avatar_file_path := filepath.Join(Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"], elizaUser.picFileName); !IsFile(avatar_file_path) {
		FileCopy(If(IsDevMode, elizaUser.picFileName, "/"+elizaUser.picFileName), avatar_file_path)
	}

	ctx := NewCtxNonHttp(yojobs.Timeout1Min, false, "")
	defer ctx.OnDone(nil)
	if user_auth := yoauth.ByEmailAddr(ctx, elizaUser.emailAddr); user_auth != nil {
		elizaUser.id = yodb.FindOne[User](ctx, UserAuth.Equal(user_auth.Id)).Id
	} else {
		ctx.DbTx()
		auth_id := yodb.CreateOne[yoauth.UserAuth](ctx, &yoauth.UserAuth{
			EmailAddr: yodb.Text(elizaUser.emailAddr),
		})
		user := &User{
			LastSeen:  yodb.DtNow(),
			PicFileId: yodb.Text(elizaUser.picFileName),
			Nick:      yodb.Text(elizaUser.nick),
			Btw:       yodb.Text(elizaUser.btw),
		}
		user.Auth.SetId(auth_id)
		_ = yodb.CreateOne[User](ctx, user)
	}
}
