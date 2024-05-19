package executor

import (
	"sync"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

type jobEntry struct {
	// deleted tells if job is deleted or not
	deleted bool
	// version tells the version of the job, useful while updating job
	version int
	job     *jobmodels.Job
}
type executorImpl struct {
	rw    *sync.Mutex
	jobs  map[string]jobEntry
	jobCh chan *jobmodels.Job
}

func NewExecutor() Executor {
	return &executorImpl{
		rw:    new(sync.Mutex),
		jobs:  make(map[string]jobEntry),
		jobCh: make(chan *jobmodels.Job, 100),
	}
}

func (e *executorImpl) Run(job jobmodels.Job) error {
	var triggerTime = getTriggerTime(job)
	if triggerTime.Before(time.Now()) {
		return ErrToLate
	}
	e.rw.Lock()
	defer e.rw.Unlock()

	e.jobs[job.ID] = jobEntry{job: &job}
	e.scheduleJob(job.ID, 0, triggerTime)
	return nil
}

func (e *executorImpl) Update(job jobmodels.Job) error {
	var triggerTime = getTriggerTime(job)
	if triggerTime.Before(time.Now()) {
		return ErrToLate
	}
	e.rw.Lock()
	defer e.rw.Unlock()
	entry, ok := e.jobs[job.ID]
	if ok {
		entry.version++ // increment version number
		entry.job = &job
		e.jobs[job.ID] = entry
		e.scheduleJob(job.ID, entry.version, triggerTime)
		return nil
	}
	return ErrJobNotFound
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

func (e *executorImpl) JobCh() chan *jobmodels.Job {
	return e.jobCh
}

func (e *executorImpl) Close() {
	close(e.jobCh)
}

func (e *executorImpl) scheduleJob(jobId string, version int, triggerTime time.Time) {

	time.AfterFunc(time.Until(triggerTime), func() {
		e.rw.Lock()
		defer e.rw.Unlock()
		entry, ok := e.jobs[jobId]
		if !ok {
			return // might be executed already
		}

		if entry.deleted {
			// skip execution, job needs to be deleted
			// delete from map
			delete(e.jobs, jobId)
		} else if entry.version == version {
			// latest job version
			delete(e.jobs, jobId)

			e.jobCh <- entry.job
		}
	})
}

func getTriggerTime(job jobmodels.Job) time.Time {
	return time.UnixMilli(int64(job.TriggerMS))
}
