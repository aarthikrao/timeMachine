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
	mu   sync.Mutex
	jobs map[string]jobEntry

	outboundJobs   chan<- *jobmodels.Job
	jobQueue       JobQueue
	isClosed       bool
	stopDispatcher context.CancelFunc
	wgDispacther   sync.WaitGroup

	// gracePeriod is the time duration for which the executor will wait for the jobs to be executed during shutdown
	// before forcefully closing the executor.
	gracePeriod time.Duration

	// accuracy is the time duration after which the executor will check for the jobs to be executed.
	// the smaller the accuracy, the more accurate the executor will be in executing the jobs.
	accuracy time.Duration
}

var _ Executor = (*executorImpl)(nil)

// NewExecutor creates a new executor that will start dispatching jobs
// when the trigger time of the job is reached.
//
// Refer executorImpl for more details.
func NewExecutor(jobCh chan<- *jobmodels.Job, gracePeriod time.Duration, accuracy time.Duration) *executorImpl {
	impl := &executorImpl{
		jobs:         make(map[string]jobEntry),
		jobQueue:     NewJobHeap(),
		outboundJobs: jobCh,
		gracePeriod:  gracePeriod,
		accuracy:     accuracy,
	}
	ctx, cancelFn := context.WithCancel(context.TODO())
	impl.stopDispatcher = cancelFn
	go impl.startDispatcher(ctx)
	return impl
}

func (e *executorImpl) GetJob(jobId string) (job *jobmodels.Job, version int, deleted bool, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, ok := e.jobs[jobId]
	if ok {
		return entry.job, entry.version, entry.deleted, nil
	}
	return nil, 0, false, ErrJobNotFound
}

func (e *executorImpl) Queue(job jobmodels.Job) error {
	if e.isClosed {
		return ErrExecutorIsClosed
	}

	if job.GetTriggerTime().Before(time.Now()) {
		return ErrTooLate
	}

	inGracePeriod := e.jobLiesWithinGracePeriod(&job)

	e.mu.Lock()
	defer e.mu.Unlock()

	// check if the job exists in the executor
	entry, exists := e.jobs[job.ID]
	if !exists { // its a new job
		if !inGracePeriod {
			return ErrNotWithinExecutorGracePeriod
		}

		entry = jobEntry{
			job: &job,
		}
		e.jobs[job.ID] = entry
		e.jobQueue.AddJob(&entry)

	} else if exists && !inGracePeriod {
		// This means the updated trigger time of the job doesnt lie within the graceperiod
		// hence we can delete the job. This job will be added again to the queue when the time comes
		entry.deleted = true
		e.jobs[job.ID] = entry

	} else {
		// update the job, increment version number to keep track of the latest job
		entry.version++
		entry.job = &job
		e.jobs[job.ID] = entry
		e.jobQueue.AddJob(&entry)
	}

	return nil
}

func (e *executorImpl) Delete(jobId string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

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
	e.isClosed = true

	time.AfterFunc(e.gracePeriod, func() {
		e.stopDispatcher()
	})

	e.wgDispacther.Wait()

	// close dispatcher channel
	// so that executor go-routines can terminate
	close(e.outboundJobs)
}

func (e *executorImpl) dispatchJob(jentry *jobEntry) {
	e.mu.Lock()
	jobId := jentry.job.ID
	entry, ok := e.jobs[jobId]
	if !ok {
		e.mu.Unlock()
		return // might be executed already
	}

	if entry.deleted {
		// skip execution, job needs to be deleted
		delete(e.jobs, jobId)
		e.mu.Unlock()

	} else if entry.version == jentry.version {
		// latest job version
		delete(e.jobs, jobId)
		e.mu.Unlock()
		e.outboundJobs <- entry.job

	} else {
		// version mismatch
		// job was updated after being queued to Run
		e.mu.Unlock()
	}

}

// fetchAndDispatch fetches a job from the scheduler queue and dispatches it for execution.
// If the job's trigger time is in the future, the job is dispatched and the function continues.
// If the job's trigger time is ahead of the current time, the job is added back to the scheduler queue.
// The function breaks out of the loop to wait for the next tick to process more jobs.
func (e *executorImpl) fetchAndDispatch() {
	for {
		// Fetch the job from the scheduler queue
		jentry := e.jobQueue.NextJob()
		if jentry == nil {
			return // We ran out of jobs
		}

		if jentry.job.GetTriggerTime().Before(time.Now()) {
			e.dispatchJob(jentry)
			continue
		}

		// Adding job back to the scheduler queue as it's ahead of current time
		// this happens when the job trigger time lies in the next tick
		e.jobQueue.AddJob(jentry)

		// We have processed jobs till current tick
		// We will wait for next tick to process more jobs
		break
	}
}

// startDispatcher is starts consuming jobs as per their trigger time
// ctx is used to terminate tickering
// it doesn't stop previous tickers jobs being dispatched if there are any
func (e *executorImpl) startDispatcher(ctx context.Context) {
	e.wgDispacther.Add(1)
	go func() {
		defer e.wgDispacther.Done()

		var ticker = time.NewTicker(e.accuracy)
		for {
			select {
			case <-ticker.C:
				e.fetchAndDispatch() // Dispatch jobs until current time
				if e.isClosed && e.jobQueue.Len() == 0 {
					// No more jobs to dispatch
					return
				}

			case <-ctx.Done():
				ticker.Stop() // Ticker is no longer required
				return
			}
		}
	}()
}

func (e *executorImpl) jobLiesWithinGracePeriod(job *jobmodels.Job) bool {
	return job.GetTriggerTime().Before(time.Now().Add(e.gracePeriod))
}
