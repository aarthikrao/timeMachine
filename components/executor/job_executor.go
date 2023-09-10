package executor

import (
	"bytes"
	"container/list"
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"go.uber.org/zap"
)

const defaultRequestContentType = "application/json"

const defaultBatchDuration = 2 * time.Minute

// defaultSlotDuration describe the default granurality of job dispatcher, e.g. 100ms means we would dispatch all the jobs in next 100 ms at once.
const defaultSlotDuration = 100 * time.Millisecond

type executor struct {
	logger *zap.Logger

	client *http.Client

	opRegister sync.Map

	cancelFn context.CancelFunc
	jobBatch *jobBatch
}

type jobVersion struct {
	id  string
	ver int8
}

type jobOp struct {
	job *jobmodels.Job
	ver int8
}

func NewJobExecutor() Executor {
	devLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("unable to get executor", zap.Error(err))
	}
	ctx, cancelFn := context.WithCancel(context.TODO())

	exe := &executor{
		logger:   devLogger,
		client:   http.DefaultClient,
		cancelFn: cancelFn,
		jobBatch: NewJobBatch(defaultBatchDuration, defaultSlotDuration),
	}

	go exe.startCounter(ctx)
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
	jop := &jobOp{job: &job}
	exe.addJob(jop)

	return nil
}

func (exe *executor) Delete(jobId string) error {
	if len(jobId) == 0 {
		exe.logger.Debug("return: no job id")
		return ErrNoJobId
	}

	exe.opRegister.Delete(jobId)
	return nil
}

func (exe *executor) Update(jobId string, newJob jobmodels.Job) error {
	if len(newJob.Route) == 0 {
		exe.logger.Debug("return: no route")
		return ErrToRoute
	}
	ijobOp, ok := exe.opRegister.Load(newJob.ID)
	if !ok {
		exe.logger.Debug("return: job not found")
		return ErrJobNotFound
	}

	jop, ok := ijobOp.(*jobOp)
	if !ok {
		return nil
	}
	latestJop := &jobOp{job: &newJob, ver: jop.ver}
	exe.addJob(latestJop)
	return nil
}

func (exe *executor) addJob(jop *jobOp) {
	jop.ver++
	err := exe.jobBatch.add(&jobVersion{jop.job.ID, jop.ver}, jop.job.TriggerMS)
	if err != nil {
		exe.logger.Debug("err while adding the job", zap.Error(err))
	}
	exe.opRegister.Store(jop.job.ID, jop)

}

func (exe *executor) dispatchBatch(batch *list.List) error {
	for batch.Len() > 0 {
		ele := batch.Front()
		if ele == nil {
			return nil
		}
		batch.Remove(ele)
		jobVer, ok := ele.Value.(*jobVersion)
		if !ok {
			continue
		}
		id := jobVer.id
		ijob, ok := exe.opRegister.LoadAndDelete(id) // Latest job version
		if !ok {
			exe.logger.Debug("job found to be deleted", zap.String("jobId", id))
			continue
		}
		jop, ok := ijob.(*jobOp)
		if !ok {
			continue
		}
		if jop.ver != jobVer.ver {
			exe.logger.Debug("found stale version", zap.String("id", jobVer.id), zap.Int8("ver", jobVer.ver))
			continue
		}
		// Runnable job
		exe.makeHttpRequest(jop.job)
	}
	return nil
}

func (exe *executor) startCounter(ctx context.Context) {
	ticker := time.NewTicker(defaultSlotDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return // Close dispatcher
		case curTime := <-ticker.C:
			exe.jobBatch.iterateBatch(curTime, exe.dispatchBatch)
		}
	}
}

// Stop, stops the dispatcher
func (exe *executor) Stop() {
	exe.cancelFn()
}

func (exe *executor) makeHttpRequest(job *jobmodels.Job) {
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
