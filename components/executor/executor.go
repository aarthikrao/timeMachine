package executor

import (
	"errors"
	"net/http"

	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
)

var (
	ErrJobNotFound = errors.New("job not found")
	ErrNoJobId     = errors.New("no job id")
	ErrTooEarly    = errors.New("too early")
	ErrToLate      = errors.New("too late")
	ErrToRoute     = errors.New("no route")
)

// Executor queues the jobs and runs them one by one.
// It also includes methods to delete or update the job so that you can change the job state
// when it is queued in the memory.
type Executor interface {
	Run(job jm.Job) error

	Delete(jobID string) error

	Update(jobID string, newjob jm.Job) error

	SetClient(client *http.Client)
}
