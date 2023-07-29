package executor

import (
	"testing"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

func TestAddJob(t *testing.T) {
	var exe = NewJobExecutor()
	err := exe.Run(jobmodels.Job{})
	if err != nil {
		t.Error("error occured while running the job")
	}

}
