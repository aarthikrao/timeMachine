package executor

import (
	"errors"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

var (
	ErrJobNotFound                  = errors.New("job not found")
	ErrTooLate                      = errors.New("too late")
	ErrNotWithinExecutorGracePeriod = errors.New("job is not within executor grace period")
)

// Executor queues the jobs and runs them one by one.
// It also includes methods to delete or update the job so that you can change the job state
// when it is queued in the memory.
type Executor interface {

	// Queue adds the job to the execution queue.
	// When the job is ready to be executed, it will be sent to the job channel.
	Queue(job jobmodels.Job) error

	// Delete deletes the queued job.
	// If the job is not queued, it will return ErrJobNotFound
	Delete(jobID string) error

	// returns the job with the given jobID.
	GetJob(jobId string) (job *jobmodels.Job, version int, deleted bool, err error)

	// Close closes the executor and waits for all the jobs to finish executing.
	Close()
}
