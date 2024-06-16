package executor

import (
	"testing"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

func TestQueueDeleteAndClose(t *testing.T) {
	jobCh := make(chan *jobmodels.Job, 1)
	gracePeriod := 15 * time.Second // Tests time out after 30 seconds
	accuracy := 50 * time.Millisecond

	executor := NewExecutor(jobCh, gracePeriod, accuracy)

	// Test executor Queue operation
	j := jobmodels.Job{
		ID:        "job1",
		TriggerMS: int(time.Now().Add(10 * time.Second).UnixMilli()),
		Meta:      nil,
		Route:     "route2",
	}
	err := executor.Queue(j)
	if err != nil {
		t.Errorf("Failed to queue job: %v", err)
	}

	job, version, deleted, err := executor.GetJob("job1")
	if err != nil {
		t.Errorf("Failed to get job: %v", err)
	} else if job.ID != j.ID {
		t.Errorf("Unexpected job ID: got %s, want %s", job.ID, j.ID)
	}

	if version != 0 {
		t.Errorf("Unexpected job version: got %d, want %d", version, 0)
	}

	if deleted {
		t.Errorf("Unexpected job deleted status: got %v, want %v", deleted, false)
	}

	// Test executor Delete operation
	err = executor.Delete("job1")
	if err != nil {
		t.Errorf("Failed to delete job: %v", err)
	}

	// Close the executor
	executor.Close()

	// Assert that the executor is closed
	if !executor.isClosed {
		t.Errorf("Expected executor to be closed, but it is not")
	}

	// Test executor Queue operation after closing
	err = executor.Queue(j)
	if err != ErrExecutorIsClosed {
		t.Errorf("Expected error: %v, got: %v", ErrExecutorIsClosed, err)
	}

}

func TestQueueAndWait(t *testing.T) {
	jobCh := make(chan *jobmodels.Job)
	gracePeriod := 15 * time.Second // Tests time out after 30 seconds
	accuracy := 50 * time.Millisecond

	executor := NewExecutor(jobCh, gracePeriod, accuracy)

	// Test executor Queue operation
	j := jobmodels.Job{
		ID:        "job1",
		TriggerMS: int(time.Now().Add(2 * time.Second).UnixMilli()),
		Meta:      nil,
		Route:     "route2",
	}
	err := executor.Queue(j)
	if err != nil {
		t.Errorf("Failed to queue job: %v", err)
	}

	job, version, deleted, err := executor.GetJob("job1")
	if err != nil {
		t.Errorf("Failed to get job: %v", err)
	} else if job.ID != j.ID {
		t.Errorf("Unexpected job ID: got %s, want %s", job.ID, j.ID)
	}

	if version != 0 {
		t.Errorf("Unexpected job version: got %d, want %d", version, 0)
	}

	if deleted {
		t.Errorf("Unexpected job deleted status: got %v, want %v", deleted, false)
	}

	// Test executor Delete operation
	recievedJob := <-jobCh
	if recievedJob.ID != j.ID {
		t.Errorf("Unexpected job ID: got %s, want %s", recievedJob.ID, j.ID)
	}

	if recievedJob.TriggerMS != j.TriggerMS {
		t.Errorf("Unexpected job TriggerMS: got %d, want %d", recievedJob.TriggerMS, j.TriggerMS)
	}

	currentMillisPlus2 := int(time.Now().Add(2 * time.Second).UnixMilli())
	if int64(currentMillisPlus2) < int64(recievedJob.TriggerMS) {
		t.Errorf("Delayed job TriggerMS: got %d, want %d", time.Now().UnixMilli(), j.TriggerMS)
	}

	// Close the executor
	executor.Close()

	// Assert that the executor is closed
	if !executor.isClosed {
		t.Errorf("Expected executor to be closed, but it is not")
	}

	// Test executor Queue operation after closing
	err = executor.Queue(j)
	if err != ErrExecutorIsClosed {
		t.Errorf("Expected error: %v, got: %v", ErrExecutorIsClosed, err)
	}

}
