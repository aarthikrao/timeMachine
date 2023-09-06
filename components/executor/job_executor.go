package executor

import (
	"bytes"
	"container/list"
	"log"
	"net/http"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"go.uber.org/zap"
)

const defaultRequestContentType = "application/json"

const (
	addJob opCode = iota + 1
	removeJob
	updateJob
)

type opCode int8

type executor struct {
	logger           *zap.Logger
	client           *http.Client
	jobRegister      map[string]*jobmodels.Job
	schdulerRegister map[int64]*list.List
	jobOpCh          chan jobOp
}

type jobOp struct {
	job       *jobmodels.Job
	operation opCode
	id        string
}

func NewJobExecutor() Executor {
	devLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("unable to get executor", zap.Error(err))
	}
	exe := &executor{
		logger:           devLogger,
		jobRegister:      make(map[string]*jobmodels.Job),
		schdulerRegister: make(map[int64]*list.List),
		jobOpCh:          make(chan jobOp),
	}
	go exe.startOperator()
	return exe
}

func (exe *executor) SetClient(client *http.Client) {
	exe.client = client
}

func (exe *executor) Run(job jobmodels.Job) error {
	if len(job.Route) == 0 {
		exe.logger.Debug("return: no route")
		return ErrToRoute
	}
	exe.jobOpCh <- jobOp{job: &job, operation: addJob, id: job.ID}
	return nil
}

func (exe *executor) Delete(jobId string) error {
	if len(jobId) == 0 {
		exe.logger.Debug("return: no job id")
		return ErrNoJobId
	}
	exe.jobOpCh <- jobOp{operation: removeJob, id: jobId}
	return nil
}

func (exe *executor) Update(jobId string, newJob jobmodels.Job) error {
	if len(newJob.Route) == 0 {
		exe.logger.Debug("return: no route")
		return ErrToRoute
	}
	exe.jobOpCh <- jobOp{job: &newJob, operation: updateJob, id: jobId}
	return nil
}

func (exe *executor) startOperator() {
	for op := range exe.jobOpCh {
		switch op.operation {
		case addJob:
			exe.jobRegister[op.id] = op.job
			scheduleTime := time.UnixMilli(op.job.TriggerMS)
			milisecDiff := time.UnixMilli(op.job.TriggerMS).Sub(time.Now()).Milliseconds()
			exe.addJob(op.id, milisecDiff)
		case removeJob:
			delete(exe.jobRegister, op.id)
		case updateJob:
			exe.jobRegister[op.id] = op.job
		}
	}
}

func (exe *executor) addJob(jobId string, milisecDiff int64) {
	jobList, ok := exe.schdulerRegister[milisecDiff]
	if !ok {
		jobList = list.New()
		exe.schdulerRegister[milisecDiff] = jobList
	}
	jobList.PushBack(jobId)

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
