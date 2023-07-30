package datastore

import (
	"errors"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	"os"
	"testing"
)

// TestCreateBoltDataStore tests the creation of BoltDataStore.
func TestCreateBoltDataStore(t *testing.T) {
	path := tempfile()
	defer os.RemoveAll(path)

	// Attempt to create a Bolt Data Store.
	dbStore, err := CreateBoltDataStore(path)
	if err != nil {
		t.Errorf("error while opening the db %v", err)
		t.Fail()
		return
	}

	// Attempt to close the database.
	err = dbStore.Close()
	if err != nil {
		t.Errorf("error while closing the db %v", err)
		t.Fail()
	}
}

// TestBoltDataStore_CreateGetSetClose tests the functionality of setting, getting, deleting jobs and closing the BoltDataStore.
func TestBoltDataStore_CreateGetSetDeleteJobCycle(t *testing.T) {
	path := tempfile()       // Create a temporary file path for testing.
	defer os.RemoveAll(path) // Make sure to remove all files from the path when finished.

	// Define a test job.
	testJob := &jm.Job{
		ID:        "test_job",
		TriggerMS: 100,
		Meta:      []byte(`{"id":"test_id","data":["str1","str2",99,{"my":"obj","id":42}]}`),
		Route:     "test_route",
	}

	// Create a Bolt Data Store.
	dbStore, err := CreateBoltDataStore(path)
	if err != nil {
		t.Errorf("error while opening the db %v", err)
		t.Fail()
		return
	}

	// Try to set a job in the database.
	err = dbStore.SetJob(path, testJob)
	if err != nil {
		t.Errorf("error while setting the job in the db %v", err)
		t.Fail()
		return
	}

	// Try to retrieve the same job from the database.
	retrievedJob, err := dbStore.GetJob(path, testJob.ID)
	if err != nil {
		t.Errorf("error while getting the job from the db %v", err)
		t.Fail()
		return
	}

	// Compare the retrieved job with the original one.
	if testJob == retrievedJob {
		t.Errorf("the db store retrievedJob was not equal to the testJob")
		t.Fail()
		return
	}

	// Try to delete the job.
	err = dbStore.DeleteJob(path, testJob.ID)
	if err != nil {
		t.Errorf("error while deleting the job from the db %v", err)
		t.Fail()
		return
	}

	// Make sure the job is indeed deleted by attempting to retrieve it again.
	_, err = dbStore.GetJob(path, testJob.ID)
	if err == nil {
		t.Errorf("job was not deleted from the dbstore because it can be retrieved still")
		t.Fail()
		return
	}

	// The expected error after deleting should be ErrKeyNotFound.
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("error returned was not ErrKeyNotFound from GetJob with no existing id")
		t.Fail()
		return
	}

	// Attempt to close the database.
	err = dbStore.Close()
	if err != nil {
		t.Errorf("error while closing the db %v", err)
		t.Fail()
	}
}

// This function tests if we can set a job in the BoltDB data store without first creating the data store.
func TestBoltDataStore_SetJobWithNewCollection(t *testing.T) {
	path := tempfile()
	defer os.RemoveAll(path)

	dbStore, err := CreateBoltDataStore(path)
	defer dbStore.Close()

	if err != nil {
		t.Errorf("error while opening the db %v", err)
		t.Fail()
		return
	}

	// Define a test job that we'll try to store in the data store.
	testJob := &jm.Job{
		ID:        "test_job",
		TriggerMS: 100,
		Meta:      []byte(`{"id":"test_id","data":["str1","str2",99,{"my":"obj","id":42}]}`),
		Route:     "test_route",
	}

	// Create another temporary file path where we'll try to set the job.
	differentPath := tempfile()
	defer os.RemoveAll(differentPath)

	// Try to set the job in the data store for a new collection path
	err = dbStore.SetJob(differentPath, testJob)
	if err != nil {
		t.Errorf("error while setting the job in the db %v", err)
		t.Fail()
		return
	}
}

// TestBoltDataStore_GetJobNonExistingCollection is a unit test function for testing the behavior of the GetJob
// method when attempting to retrieve a job from a non-existing collection in the BoltDB data store.
func TestBoltDataStore_GetJobNonExistingCollection(t *testing.T) {
	path := tempfile()
	defer os.RemoveAll(path)

	dbStore, err := CreateBoltDataStore(path)
	defer dbStore.Close()

	if err != nil {
		t.Errorf("error while opening the db %v", err)
		t.Fail()
		return
	}

	differentPath := tempfile()
	defer os.RemoveAll(differentPath)

	_, err = dbStore.GetJob(differentPath, "test_id")
	if err == nil {
		t.Errorf("job was returned when no job should have been returned from collection")
		t.Fail()
		return
	}

	// The expected error after deleting should be ErrBucketNotFound.
	if !errors.Is(err, ErrBucketNotFound) {
		t.Errorf("error returned was not ErrBucketNotFound from GetJob call")
		t.Fail()
		return
	}
}

// tempfile returns a temporary file path.
// It creates a temporary file, closes it and then removes it just to get a temporary file path.
func tempfile() string {
	f, err := os.CreateTemp("", "_bolt_temp")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
