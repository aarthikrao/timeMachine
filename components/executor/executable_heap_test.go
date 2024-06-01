package executor

import (
	"encoding/json"
	"testing"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

func TestJobHeap(t *testing.T) {
	// Create a new job heap
	jh := NewJobHeap()

	// Add some job entries
	entry1 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job1",
			TriggerMS: 100,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	entry2 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job2",
			TriggerMS: 200,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	entry3 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job3",
			TriggerMS: 50,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	jh.AddJob(entry1)
	jh.AddJob(entry2)
	jh.AddJob(entry3)

	// Check the length of the job heap
	if jh.Len() != 3 {
		t.Errorf("Expected job heap length to be 3, got %d", jh.Len())
	}

	// Get the next job entry
	nextJob := jh.NextJob()

	// Check if the next job entry is correct
	if nextJob.job.ID != entry3.job.ID {
		t.Errorf("Expected next job entry to be %v, got %v", entry3.job.ID, nextJob.job.ID)
	}

	// Check the length of the job heap after popping
	if jh.Len() != 2 {
		t.Errorf("Expected job heap length to be 2, got %d", jh.Len())
	}
}

func TestJobHeapPush(t *testing.T) {
	jh := NewJobHeap()

	// Add some job entries
	entry1 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job1",
			TriggerMS: 100,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	entry2 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job2",
			TriggerMS: 200,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	entry3 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job3",
			TriggerMS: 50,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	jh.AddJob(entry1)
	jh.AddJob(entry2)
	jh.AddJob(entry3)

	// Check the length of the job heap
	if jh.Len() != 3 {
		t.Errorf("Expected job heap length to be 3, got %d", jh.Len())
	}

	// Get the next job entry
	nextJob := jh.NextJob()

	// Check if the next job entry is correct
	if nextJob.job.ID != entry3.job.ID {
		t.Errorf("Expected next job entry to be %v, got %v", entry3.job.ID, nextJob.job.ID)
	}

	// Check the length of the job heap after popping
	if jh.Len() != 2 {
		t.Errorf("Expected job heap length to be 2, got %d", jh.Len())
	}

	// Get the next job entry
	nextJob = jh.NextJob()

	// Check if the next job entry is correct
	if nextJob.job.ID != entry1.job.ID {
		t.Errorf("Expected next job entry to be %v, got %v", entry1.job.ID, nextJob.job.ID)
	}
}

func TestJobHeapNextJob(t *testing.T) {
	jh := NewJobHeap()

	// Add some job entries
	entry1 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job1",
			TriggerMS: 100,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	entry2 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job2",
			TriggerMS: 200,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	entry3 := &jobEntry{
		deleted: false,
		version: 1,
		job: &jobmodels.Job{
			ID:        "job3",
			TriggerMS: 50,
			Meta:      json.RawMessage{},
			Route:     "route1",
		},
	}
	jh.AddJob(entry1)
	jh.AddJob(entry2)
	jh.AddJob(entry3)

	// Get the next job entry
	nextJob := jh.NextJob()

	// Check if the next job entry is correct
	if nextJob.job.ID != entry3.job.ID {
		t.Errorf("Expected next job entry to be %v, got %v", entry3.job.ID, nextJob.job.ID)
	}

	// Get the next job entry
	nextJob = jh.NextJob()

	// Check if the next job entry is correct
	if nextJob.job.ID != entry1.job.ID {
		t.Errorf("Expected next job entry to be %v, got %v", entry1.job.ID, nextJob.job.ID)
	}
}
