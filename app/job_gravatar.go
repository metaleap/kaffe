package kaffe

import (
	"path/filepath"

	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	yojobs "yo/jobs"
	. "yo/util"
	"yo/util/sl"
	"yo/web/gravatar"
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
		src_img, got_resp_body_read, _ := gravatar.ImageByEmailAddr(ctx, user.Account.Get(ctx).EmailAddr.String())
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
