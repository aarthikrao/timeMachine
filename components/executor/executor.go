package executor

import (
	"errors"

	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
)

var (
	ErrJobNotFound = errors.New("job not found")
)

// Executor queues the jobs and runs them one by one.
// It also includes methods to delete or update the job so that you can change the job state
// when it is queued in the memory.
type Executor interface {

	// Run adds the job to the execution queue.
	// Internally it is added in a time.After function
	Run(job jm.Job) error

	// // Updates the job details after the job is queued.
	// // If the job is not queued, it will return ErrJobNotFound
	// Update(job jm.Job) error

	// // Deletes the queued job.
	// // If the job is not queued, it will return ErrJobNotFound
	// Delete(jobID string) error
}
