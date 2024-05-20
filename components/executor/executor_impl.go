package executor

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

var (
	ErrExecutorIsClosed = errors.New("executor is closed")
)

type jobEntry struct {
	// deleted tells if job is deleted or not
	deleted bool
	// version tells the version of the job, useful while updating job
	version int
	job     *jobmodels.Job
}
type executorImpl struct {
	rw             *sync.Mutex
	jobs           map[string]jobEntry
	jobCh          chan<- *jobmodels.Job
	jobQueue       jobQueue
	isClosed       bool
	stopDispatcher context.CancelFunc
	wgDispacther   sync.WaitGroup
}

func NewExecutor(jobCh chan<- *jobmodels.Job) Executor {

	impl := &executorImpl{
		rw:       new(sync.Mutex),
		jobs:     make(map[string]jobEntry),
		jobQueue: &jobHeap{},
		jobCh:    jobCh,
	}
	ctx, cancelFn := context.WithCancel(context.TODO())
	impl.stopDispatcher = cancelFn
	go impl.startDispatcher(ctx)
	return impl
}

func (e *executorImpl) Run(job jobmodels.Job) error {
	if e.isClosed {
		return ErrExecutorIsClosed
	}

	var triggerTime = getTriggerTime(&job)
	if triggerTime.Before(time.Now()) {
		return ErrToLate
	}
	var entry = jobEntry{job: &job}
	e.rw.Lock()
	e.jobs[job.ID] = entry
	e.rw.Unlock()
	e.jobQueue.addJob(&entry)
	return nil
}

func (e *executorImpl) Update(job jobmodels.Job) error {
	if e.isClosed {
		return ErrExecutorIsClosed
	}

	var triggerTime = getTriggerTime(&job)
	if triggerTime.Before(time.Now()) {
		return ErrToLate
	}
	e.rw.Lock()
	entry, ok := e.jobs[job.ID]
	if !ok {
		e.rw.Unlock()
		return ErrJobNotFound
	}
	entry.version++ // increment version number
	entry.job = &job
	e.jobs[job.ID] = entry
	e.rw.Unlock()
	e.jobQueue.addJob(&entry)
	return nil
}

func (e *executorImpl) Delete(jobId string) error {
	e.rw.Lock()
	defer e.rw.Unlock()
	entry, ok := e.jobs[jobId]
	if ok {
		entry.deleted = true
		e.jobs[jobId] = entry
		return nil
	}
	return ErrJobNotFound
}

// Close closes the executor and waits for all the jobs to finish executing.
func (e *executorImpl) Close() {
	e.rw.Lock()
	defer e.rw.Unlock()

	e.isClosed = true

	// Stop dispatcher and wait for it to stop
	e.stopDispatcher()
	e.wgDispacther.Wait()

	// close dispatcher channel
	// so that executor go-routines can terminate
	close(e.jobCh)
}

func (e *executorImpl) nextTick() {
	var now = time.Now()
	for {
		jentry := e.jobQueue.nextJob()
		if jentry == nil {
			return // We ran out of jobs
		}
		triggerTime := getTriggerTime(jentry.job)
		if now.After(triggerTime) {
			// Adding job back to the scheduler queue as it's ahead of current time
			e.jobQueue.addJob(jentry)
			return // We have processed jobs till current tick
		}
		e.rw.Lock()
		jobId := jentry.job.ID
		entry, ok := e.jobs[jobId]
		if !ok {
			e.rw.Unlock()
			continue // might be executed already
		}

		if entry.deleted {
			// skip execution, job needs to be deleted
			// delete from map
			delete(e.jobs, jobId)
			e.rw.Unlock()
			continue
		} else if entry.version == jentry.version {
			// latest job version
			delete(e.jobs, jobId)
			e.rw.Unlock()
			e.jobCh <- entry.job
		} else {
			e.rw.Unlock()
			continue
		}
	}
}

// startDispatcher is starts consuming jobs as per their trigger time
// ctx is used to terminate tickering
// it doesn't stop previous tickers jobs being dispatched if there are any
func (e *executorImpl) startDispatcher(ctx context.Context) {
	e.wgDispacther.Add(1)
	go func() {
		defer e.wgDispacther.Done()
		var ticker = time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				e.nextTick()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func getTriggerTime(job *jobmodels.Job) time.Time {
	return time.UnixMilli(int64(job.TriggerMS))
}
