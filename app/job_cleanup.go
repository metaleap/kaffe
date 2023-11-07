package haxsh

import (
	"time"
	. "yo/cfg"
	. "yo/ctx"
	yodb "yo/db"
	q "yo/db/query"
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
	Schedules:                        yojobs.ScheduleOncePerDay,
	TimeoutSecsTaskRun:               123,
	TimeoutSecsJobRunPrepAndFinalize: 123,
	Disabled:                         false,
	DeleteAfterDays:                  1,
	MaxTaskRetries:                   1, // keep low, the to-dos will anyway resurface for the next job run
}

type cleanUpJob Void
type cleanUpJobDetails Void
type cleanUpJobResults Void
type cleanUpTaskDetails struct {
	User       yodb.I64
	FileDelReq yodb.I64
}
type cleanUpTaskResults struct {
	NumPostsDeleted int
	NumFilesDeleted int
}

type fileDelReq struct {
	Id        yodb.I64
	DtMade    *yodb.DateTime
	DtMod     *yodb.DateTime
	FileNames yodb.Arr[yodb.Text]
}

func (me cleanUpJob) JobDetails(ctx *Ctx) yojobs.JobDetails {
	return nil
}

func (cleanUpJob) JobResults(_ *Ctx) (func(func() *Ctx, *yojobs.JobTask, *bool), func() yojobs.JobResults) {
	return nil, nil
}

func (cleanUpJob) dtCutOff() time.Time {
	return time.Now().AddDate(0, 0, -CfgGet[int](cfgEnvNameDeletePostsOlderThanDays))
}

func (me cleanUpJob) TaskDetails(ctx *Ctx, stream func([]yojobs.TaskDetails)) {
	// file-deletion job tasks from pending file-deletion reqs
	stream(sl.To(
		yodb.Ids[fileDelReq](ctx, nil),
		func(id yodb.I64) yojobs.TaskDetails {
			return &cleanUpTaskDetails{FileDelReq: id}
		}))

	// post-deletion job tasks from non-vip users that have old posts
	user_ids := make(sl.Of[yodb.I64], 0, 128)
	do_push := func(users []yodb.I64) {
		stream(sl.To(users, func(it yodb.I64) yojobs.TaskDetails {
			return &cleanUpTaskDetails{User: it}
		}))
	}
	dt_ago := me.dtCutOff()
	for _, user_id := range yodb.Ids[User](ctx, userVip.Equal(false).And(UserLastSeen.GreaterThan(UserDtMod))) {
		if yodb.Exists[Post](ctx, PostBy.Equal(user_id).And(PostDtMade.LessThan(dt_ago))) {
			user_ids.BufNext(user_id, do_push)
		}
	}
	user_ids.BufDone(do_push)
}

func (me cleanUpJob) TaskResults(ctx *Ctx, taskDetails yojobs.TaskDetails) yojobs.TaskResults {
	task_details, ret := taskDetails.(*cleanUpTaskDetails), &cleanUpTaskResults{}
	// file deletions
	file_del_req := yodb.ById[fileDelReq](ctx, task_details.FileDelReq)
	if file_del_req != nil {
		for _, file_name := range file_del_req.FileNames {
			if file_path := userUploadedFilePath(file_name.String()); IsFile(file_path) {
				DelFile(file_path)
				ret.NumFilesDeleted++
			}
		}
		yodb.Delete[fileDelReq](ctx, fileDelReqId.Equal(file_del_req.Id))
	}

	// post deletions
	if task_details.User != 0 {
		query := PostBy.Equal(task_details.User).And(PostDtMade.LessThan(me.dtCutOff()))
		num_rows_affected := yodb.Delete[Post](ctx, query.And(q.ArrIsEmpty(PostTo)))
		ret.NumPostsDeleted += int(num_rows_affected)

		var del_post_ids sl.Of[yodb.I64]
		for _, post := range yodb.FindMany[Post](ctx, query, 0, PostFields(PostId, PostTo)) {
			any_vip := yodb.Exists[User](ctx, UserId.In(post.To.ToAnys()...).And(userVip.Equal(true)))
			if !any_vip {
				del_post_ids = append(del_post_ids, post.Id)
			}
		}
		if len(del_post_ids) > 0 {
			num_rows_affected := yodb.Delete[Post](ctx, query.And(PostId.In(del_post_ids.ToAnys()...)))
			ret.NumPostsDeleted += int(num_rows_affected)
		}
	}
	return ret
}
