package executor

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"go.uber.org/zap"
)

const defaultRequestContentType = "application/json"

type executor struct {
	logger     *zap.Logger
	client     *http.Client
	udRegister sync.Map

	opSeq atomic.Int64
}

func NewJobExecutor() Executor {
	devLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("unable to get executor", zap.Error(err))
	}
	return &executor{

		logger: devLogger,
	}
}

func (exe *executor) SetClient(client *http.Client) {
	exe.client = client
}

func (exe *executor) Run(job jobmodels.Job) error {
	if len(job.Route) == 0 {
		exe.logger.Debug("return: no route")
		return nil
	}

	exe.scheduleJob(job, 0)
	return nil
}

func (exe *executor) scheduleJob(job jobmodels.Job, seq int64) {
	triggerTime := time.UnixMilli(job.TriggerMS)
	time.AfterFunc(time.Until(triggerTime), exe.runnable(job, seq))
}

func (exe *executor) runnable(job jobmodels.Job, seq int64) func() {
	return func() {
		opSeqi, ok := exe.udRegister.Load(job.ID)
		if ok {
			opSeq := opSeqi.(int64)
			if opSeq < 0 {
				exe.logger.Debug("job found to be deleted", zap.Any("job", job))
				return
			} else if opSeq > 0 && seq != opSeq {
				exe.logger.Debug("out-dated job found", zap.Any("job", job))
				return
			}
		}
		exe.makeHttpRequest(job)
	}
}

func (exe *executor) Delete(jobId string) error {
	delSeq := exe.opSeq.Add(1)
	x := -delSeq
	exe.udRegister.Store(jobId, x)
	return nil
}

func (exe *executor) Update(jobId string, newJob jobmodels.Job) error {
	if len(newJob.Route) == 0 {
		return nil
	}
	updateSeq := exe.opSeq.Add(1)
	exe.udRegister.Store(jobId, updateSeq)
	exe.scheduleJob(newJob, updateSeq)
	return nil
}

func (exe *executor) makeHttpRequest(job jobmodels.Job) {
	exe.logger.Debug("request recieved to execute", zap.Any("job", job))

	_, err := exe.client.Post(job.Route, defaultRequestContentType, bytes.NewReader(job.Meta))
	if err != nil {
		exe.logger.Error("error while posting job content",
			zap.Any("job", job),
			zap.Error(err))
		return
	}
	exe.logger.Debug("done with makeing post call", zap.Any("job", job))
}
