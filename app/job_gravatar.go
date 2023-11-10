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
	Schedules:                        yojobs.ScheduleOncePerHour,
	TimeoutSecsTaskRun:               22,
	TimeoutSecsJobRunPrepAndFinalize: 44,
	Disabled:                         false,
	DeleteAfterDays:                  1,
	MaxTaskRetries:                   1, // keep low, the to-dos will anyway resurface for the next job run
}

type gravatarJob Void
type gravatarJobDetails Void
type gravatarJobResults Void
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
	stream(sl.To(
		yodb.FindMany[User](ctx, userGravatarChecked.Equal(false).And(UserPicFileId.Equal("")), 11, UserFields(UserId), UserLastSeen.Desc()),
		func(it *User) yojobs.TaskDetails { return &gravatarTaskDetails{UserId: it.Id} },
	))
}

func (me gravatarJob) TaskResults(ctx *Ctx, taskDetails yojobs.TaskDetails) yojobs.TaskResults {
	task_details := taskDetails.(*gravatarTaskDetails)
	user := yodb.ById[User](ctx, task_details.UserId)
	if (user != nil) && (!user.gravatarChecked) && (user.PicFileId == "") {
		user_auth := user.Auth.Get(ctx)
		sha256 := sha256.New()
		_, _ = sha256.Write([]byte(str.Lo(user_auth.EmailAddr.String())))
		hash_of_email_addr := str.Fmt("%x", sha256.Sum(nil))
		if http_req, _ := http.NewRequestWithContext(ctx, "GET", "https://gravatar.com/avatar/"+hash_of_email_addr+"?d=404", nil); http_req != nil {
			if resp, _ := http.DefaultClient.Do(http_req); (resp != nil) && (resp.Body != nil) {
				defer resp.Body.Close()
				var buf bytes.Buffer
				if _, _ = io.Copy(&buf, resp.Body); buf.Len() > 0 {
					src_raw := buf.Bytes()
					if img, _, _ := image.Decode(&buf); img != nil {
						file_name := "gravatar_" + hash_of_email_addr
						if file_path := filepath.Join(Cfg.STATIC_FILE_STORAGE_DIRS["_postfiles"], file_name); !FsIsFile(file_path) {
							FsWrite(file_path, src_raw)
						}
						user.PicFileId = yodb.Text(file_name)
					}
				}
				user.gravatarChecked = true
			}
		}
		yodb.Update[User](ctx, user, nil, false, UserFields(userGravatarChecked, UserPicFileId)...)
	}
	return nil
}
