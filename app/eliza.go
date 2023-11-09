package kaffe

import (
	"path/filepath"

	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	yoauth "yo/feat_auth"
	yojobs "yo/jobs"
	. "yo/util"

	_ "github.com/NortySpock/eliza-go/eliza"
)

var elizaAvatarImageFileName = "eliza.png"
var elizaUserId yodb.I64

func elizaEnsuerUser() {
	if avatar_file_path := filepath.Join(Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"], elizaAvatarImageFileName); !IsFile(avatar_file_path) {
		FileCopy(If(IsDevMode, elizaAvatarImageFileName, "/"+elizaAvatarImageFileName), avatar_file_path)
	}

	ctx, email_addr := NewCtxNonHttp(yojobs.Timeout1Min, false, ""), "eliza@metaleap.net"
	defer ctx.OnDone(nil)
	if user_auth := yoauth.ByEmailAddr(ctx, email_addr); user_auth != nil {
		elizaUserId = yodb.FindOne[User](ctx, UserAuth.Equal(user_auth.Id)).Id
	} else {
		ctx.DbTx()
		auth_id := yodb.CreateOne[yoauth.UserAuth](ctx, &yoauth.UserAuth{
			EmailAddr: yodb.Text(email_addr),
		})
		user := &User{
			LastSeen:  yodb.DtNow(),
			PicFileId: yodb.Text(elizaAvatarImageFileName),
			Nick:      "eliza",
			Btw:       "I'm Weizenbaum's O.G. chatbot (Go impl by mattshiel & NortySpock)",
		}
		user.Auth.SetId(auth_id)
		_ = yodb.CreateOne[User](ctx, user)
	}
}
