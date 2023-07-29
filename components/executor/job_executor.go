package executor

import (
	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

type executor struct {
}

func NewJobExecutor() Executor {
	return &executor{}
}

func (exe *executor) Run(job jobmodels.Job) error {

	return nil

}
