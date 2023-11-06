package haxsh

import (
	yodb "yo/db"
	yojobs "yo/jobs"
	. "yo/util"
)

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
type cleanUpTaskDetails struct{ FileNames []string }
type cleanUpTaskResults Void

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
	// reqs := yodb.FindMany[MailReq](ctx.Ctx, mailReqDtDone.Equal(nil), 0, nil)
	// stream(sl.To(reqs,
	// 	func(it *MailReq) yojobs.TaskDetails { return &mailReqTaskDetails{ReqId: it.Id} }))
}

func (me cleanUpJob) TaskResults(ctx *yojobs.Context, task yojobs.TaskDetails) yojobs.TaskResults {
	return nil
}
