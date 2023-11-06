package haxsh

import (
	"time"
	. "yo/cfg"
	yodb "yo/db"
	yojobs "yo/jobs"
	. "yo/util"
	"yo/util/sl"
)

const cfgEnvNameDeletePostsOlderThanDays = "DEL_POSTS_OLDER_THAN_DAYS"

var cleanUpJobTypeId = yojobs.Register[cleanUpJob, cleanUpJobDetails, cleanUpJobResults, cleanUpTaskDetails, cleanUpTaskResults](func(string) cleanUpJob {
	return cleanUpJob{}
})
var cleanUpJobDef = yojobs.JobDef{
	Name:                             yodb.Text(ReflType[cleanUpJob]().String()),
	JobTypeId:                        yodb.Text(cleanUpJobTypeId),
	Schedules:                        yodb.Arr[yodb.Text]{"0 3 * * *"}, // nightly, 3am
	TimeoutSecsTaskRun:               11,
	TimeoutSecsJobRunPrepAndFinalize: 123,
	Disabled:                         false,
	DeleteAfterDays:                  1,
	MaxTaskRetries:                   123,
}

type cleanUpJob Void
type cleanUpJobDetails Void
type cleanUpJobResults Void
type cleanUpTaskDetails struct {
	User       yodb.I64
	FileDelReq yodb.I64
}
type cleanUpTaskResults struct{ NumFilesDeleted int }

type fileDelReq struct {
	Id        yodb.I64
	DtMade    *yodb.DateTime
	DtMod     *yodb.DateTime
	FileNames yodb.Arr[yodb.Text]
}

func (me cleanUpJob) JobDetails(ctx *yojobs.Context) yojobs.JobDetails {
	return nil
}

func (cleanUpJob) JobResults(_ *yojobs.Context) (func(*yojobs.JobTask, *bool), func() yojobs.JobResults) {
	return nil, nil
}

func (cleanUpJob) TaskDetails(ctx *yojobs.Context, stream func([]yojobs.TaskDetails)) {
	stream(sl.To(
		yodb.FindMany[fileDelReq](ctx.Ctx, nil, 0, nil),
		func(it *fileDelReq) yojobs.TaskDetails {
			return &cleanUpTaskDetails{FileDelReq: it.Id}
		}))

	user_ids := make([]yodb.I64, 0, 128)
	do_push := func() {
		stream(sl.To(user_ids, func(it yodb.I64) yojobs.TaskDetails {
			return &cleanUpTaskDetails{User: it}
		}))
	}
	dt_ago := time.Now().AddDate(0, 0, -CfgGet[int](cfgEnvNameDeletePostsOlderThanDays))
	yodb.Each[User](ctx.Ctx, userVip.Equal(false), 0, nil, func(rec *User, enough *bool) {
		if yodb.Exists[Post](ctx.Ctx, PostDtMade.LessThan(dt_ago)) {
			if len(user_ids) == cap(user_ids) {
				do_push()
				user_ids = user_ids[:0]
			}
		}
	})
	do_push()
}

func (me cleanUpJob) TaskResults(ctx *yojobs.Context, taskDetails yojobs.TaskDetails) yojobs.TaskResults {
	task_details, ret := taskDetails.(*cleanUpTaskDetails), &cleanUpTaskResults{}
	file_del_req := yodb.ById[fileDelReq](ctx.Ctx, task_details.FileDelReq)
	if file_del_req != nil {
		for _, file_name := range file_del_req.FileNames {
			if file_path := userUploadedFilePath(file_name.String()); IsFile(file_path) {
				DelFile(file_path)
				ret.NumFilesDeleted++
			}
		}
		yodb.Delete[fileDelReq](ctx.Ctx, fileDelReqId.Equal(file_del_req.Id))
	}
	return ret
}
