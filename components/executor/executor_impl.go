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

type timeCompareFn = func(time.Time) bool

type jobEntry struct {
	// deleted tells if job is deleted or not
	deleted bool
	// version tells the version of the job, useful while updating job
	version int
	job     *jobmodels.Job
}
type executorImpl struct {
	// rw: lock for jobs map
	rw             *sync.Mutex
	jobs           map[string]jobEntry
	jobCh          chan<- *jobmodels.Job
	jobQueue       jobQueue
	isClosed       bool
	stopDispatcher context.CancelFunc
	wgDispacther   sync.WaitGroup
	nextMin        int64
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

func (e *executorImpl) SetNextMin(min int64) {
	e.nextMin = min
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

func (e *executorImpl) dispatchJob(jentry *jobEntry) {
	e.rw.Lock()
	jobId := jentry.job.ID
	entry, ok := e.jobs[jobId]
	if !ok {
		e.rw.Unlock()
		return // might be executed already
	}

	if entry.deleted {
		// skip execution, job needs to be deleted
		// delete from map
		delete(e.jobs, jobId)
		e.rw.Unlock()
	} else if entry.version == jentry.version {
		// latest job version
		delete(e.jobs, jobId)
		e.rw.Unlock()
		e.jobCh <- entry.job
	} else {
		// version mismatch
		// job was updated after being queued to Run
		e.rw.Unlock()
	}

}

func (e *executorImpl) dispatchUntil(untilFn timeCompareFn) {
	for {
		jentry := e.jobQueue.nextJob()
		if jentry == nil {
			return // We ran out of jobs
		}
		triggerTime := getTriggerTime(jentry.job)
		if untilFn(triggerTime) {
			e.dispatchJob(jentry)
			continue
		}
		// Adding job back to the scheduler queue as it's ahead of current time
		e.jobQueue.addJob(jentry)
		break // We have processed jobs till current tick
	}
}

func untilNow(t time.Time) bool {
	return time.Now().After(t)
}

// startDispatcher is starts consuming jobs as per their trigger time
// ctx is used to terminate tickering
// it doesn't stop previous tickers jobs being dispatched if there are any
func (e *executorImpl) startDispatcher(ctx context.Context) {
	e.wgDispacther.Add(1)
	go func() {
		defer e.wgDispacther.Done()

		// look for dispatchable jobs
		var ticker = time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				e.dispatchUntil(untilNow) // Dispatch jobs until current time
			case <-ctx.Done():
				ticker.Stop() // Ticker is no longer required

				lastMoment := e.getEndOfNextMinute()
				e.dispatchUntil(func(t time.Time) bool { // Dispatch jobs until lastMoment
					return lastMoment.After(t)
				})

				return
			}
		}
	}()
}

// We know about the latest minute we have jobs for,
// we are adding one more minute to it and returning the starting moment of that minute
// We can now dispatch jobs until this time and gracefully quit as we need to execute all the jobs before quitting.

func (e *executorImpl) getEndOfNextMinute() time.Time {
	return time.UnixMilli(60000 * (e.nextMin + 1))
}

func getTriggerTime(job *jobmodels.Job) time.Time {
	return time.UnixMilli(int64(job.TriggerMS))
}
