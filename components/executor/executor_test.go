package executor

import (
	"testing"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

func TestRun(t *testing.T) {
	exe := NewExecutor()
	job := jobmodels.Job{
		ID:        "1",
		TriggerMS: int(time.Now().Add(time.Millisecond).UnixMilli()),
	}
	err := exe.Run(job)
	if err != nil {
		t.Error("error while adding job in executor")
		return
	}
	time.Sleep(time.Millisecond)
	var resultJob jobmodels.Job
	if exe.Next(&resultJob) && resultJob.ID != job.ID {
		t.Error("job didn't scheduled on time")
		return
	}

	job = jobmodels.Job{
		ID:        "2",
		TriggerMS: int(time.Now().Add(-time.Millisecond).UnixMilli()),
	}
	err = exe.Run(job)
	if err != ErrToLate {
		t.Error("job should return an error")
	}
	if exe.Next(&resultJob) {
		t.Error("there shouldn't be any scheduled job")
	}
}

func TestUpdate(t *testing.T) {
	exe := NewExecutor()
	job := jobmodels.Job{
		ID:        "1",
		TriggerMS: int(time.Now().Add(2 * time.Millisecond).UnixMilli()),
	}
	err := exe.Run(job)
	if err != nil {
		t.Error("error while adding job in executor")
		return
	}
	updatedjob := jobmodels.Job{
		ID:        "1",
		TriggerMS: int(time.Now().Add(5 * time.Millisecond).UnixMilli()),
	}

	err = exe.Update(updatedjob)
	if err != nil {
		t.Error("error while updating job")
	}

	time.Sleep(2 * time.Millisecond)
	var resultJob jobmodels.Job
	if exe.Next(&resultJob) && resultJob.ID == job.ID && resultJob.TriggerMS == updatedjob.TriggerMS {
		t.Error("out dated job found")
		return
	}
	time.Sleep(3 * time.Millisecond) // We have already slept 2 ms

	if !exe.Next(&resultJob) && resultJob.ID == updatedjob.ID && resultJob.TriggerMS == updatedjob.TriggerMS {
		t.Error("updated job didn't ran on time")
	}

}

func TestDelete(t *testing.T) {
	exe := NewExecutor()
	job := jobmodels.Job{
		ID:        "1",
		TriggerMS: int(time.Now().Add(time.Millisecond).UnixMilli()),
	}
	err := exe.Run(job)
	if err != nil {
		t.Error("error while adding job in executor")
		return
	}
	err = exe.Delete(job.ID)
	if err != nil {
		t.Error("error while deleting job")
		return
	}
	time.Sleep(2 * time.Millisecond) // waiting for 1ms extra, so that it can settle
	var resultJob jobmodels.Job
	if exe.Next(&resultJob) {
		t.Error("deleted job got scheduled")
		return
	}

}
