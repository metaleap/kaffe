package kaffe

import (
	"bytes"
	"crypto/sha256"
	"image"
	"io"
	"net/http"
	"path/filepath"

	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	yojobs "yo/jobs"
	. "yo/util"
	"yo/util/sl"
	"yo/util/str"
)

var gravatarJobTypeId = yojobs.Register[gravatarJob, gravatarJobDetails, gravatarJobResults, gravatarTaskDetails, gravatarTaskResults](func(string) gravatarJob {
	return gravatarJob{}
})
var gravatarJobDef = yojobs.JobDef{
	Name:                             yodb.Text(ReflType[gravatarJob]().String()),
	JobTypeId:                        yodb.Text(gravatarJobTypeId),
	Schedules:                        If(IsDevMode, yojobs.ScheduleOncePerMinute, yojobs.ScheduleOncePerHour),
	TimeoutSecsTaskRun:               22,
	TimeoutSecsJobRunPrepAndFinalize: 44,
	Disabled:                         false,
	DeleteAfterDays:                  1,
	MaxTaskRetries:                   1, // keep low, the to-dos will anyway resurface for the next job run
}

type gravatarJob None
type gravatarJobDetails None
type gravatarJobResults None
type gravatarTaskDetails struct {
	UserId yodb.I64
}
type gravatarTaskResults struct {
}

func (gravatarJob) JobDetails(ctx *Ctx) yojobs.JobDetails {
	return nil
}

func (gravatarJob) JobResults(_ *Ctx) (func(func() *Ctx, *yojobs.JobTask, *bool), func() yojobs.JobResults) {
	return nil, nil
}

func (me gravatarJob) TaskDetails(ctx *Ctx, stream func([]yojobs.TaskDetails)) {
	stream(sl.As(
		yodb.FindMany[User](ctx, userGravatarChecked.Equal(false).And(UserPicFileId.Equal("")), 44, UserFields(UserId), UserLastSeen.Desc()),
		func(it *User) yojobs.TaskDetails { return &gravatarTaskDetails{UserId: it.Id} },
	))
}

func (me gravatarJob) TaskResults(ctx *Ctx, taskDetails yojobs.TaskDetails) yojobs.TaskResults {
	task_details := taskDetails.(*gravatarTaskDetails)
	user := yodb.ById[User](ctx, task_details.UserId)
	if (user != nil) && (!user.gravatarChecked) && (user.PicFileId == "") {
		src_img, got_resp_body_read := imageGravatarByEmailAddr(ctx, user.Account.Get(ctx).EmailAddr.String())
		if len(src_img) > 0 {
			file_name := userUploadedFileNameNew("gravatar", int64(len(src_img)))
			if file_path := filepath.Join(Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"], file_name); !FsIsFile(file_path) {
				FsWrite(file_path, src_img)
			}
			user.PicFileId = yodb.Text(file_name)
		}
		user.gravatarChecked = yodb.Bool(got_resp_body_read)
		yodb.Update[User](ctx, user, nil, false, UserFields(userGravatarChecked, UserPicFileId)...)
	}
	return nil
}

func sha256HexOf(s string) string {
	sha256 := sha256.New()
	_, _ = sha256.Write([]byte(str.Lo(s)))
	return str.Fmt("%x", sha256.Sum(nil))
}

func imageUrlGravatarByEmailAddr(emailAddr string) string {
	return "https://gravatar.com/avatar/" + sha256HexOf(emailAddr) + "?d=404"
}
func imageGravatarByEmailAddr(ctx *Ctx, emailAddr string) (srcImg []byte, gotToRespBodyRead bool) {
	url_gravatar := imageUrlGravatarByEmailAddr(emailAddr)
	if http_req, _ := http.NewRequestWithContext(ctx, "GET", url_gravatar, nil); http_req != nil {
		if resp, _ := http.DefaultClient.Do(http_req); (resp != nil) && (resp.Body != nil) {
			defer resp.Body.Close()
			var buf bytes.Buffer
			if _, _ = io.Copy(&buf, resp.Body); buf.Len() > 0 {
				gotToRespBodyRead, srcImg = true, buf.Bytes() // must capture before Decode
				if img, _, _ := image.Decode(&buf); img == nil {
					srcImg = nil
				}
			}
		}
	}
	return
}
