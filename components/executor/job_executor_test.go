package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

func setupTestServer(t *testing.T, jsonData *json.RawMessage,
	doneFn context.CancelFunc) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer doneFn()
		var err error
		*jsonData, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error while reading request body: %v", err)
		}
	}))

	t.Log("web server started")
	return server
}

func createTestJob(triggerWindowMs int, listener net.Listener) jobmodels.Job {
	route := fmt.Sprintf("http://%s/testJob", listener.Addr().String())
	testJob := jobmodels.Job{Route: route}
	randomData := strconv.Itoa(rand.Intn(100000))
	testJob.Meta = json.RawMessage(randomData)
	testJob.TriggerMS = time.Now().Add(time.Duration(triggerWindowMs) * time.Millisecond).UnixMilli()
	return testJob
}

func TestRunJob(t *testing.T) {
	var exe = NewJobExecutor()
	err := exe.Run(jobmodels.Job{})
	if err != nil {
		t.Errorf("error occured while running the job %v", err)
	}

	var reqMetaData json.RawMessage
	done, doneFn := context.WithCancel(context.TODO())

	server := setupTestServer(t, &reqMetaData, doneFn)
	defer server.Close()
	exe.SetClient(server.Client())
	testJob := createTestJob(500, server.Listener)

	err = exe.Run(testJob)

	if err != nil {
		t.Errorf("error occured while running the job %v", err)
	}

	<-done.Done()

	if !bytes.Equal(reqMetaData, testJob.Meta) {
		t.Errorf("meta data didn't match, recieved: %s, expected: %s", reqMetaData, testJob.Meta)
	}

}

func TestDeleteJob(t *testing.T) {
	var exe = NewJobExecutor()
	var job jobmodels.Job
	err := exe.Delete(job.ID)
	if err != nil {
		t.Errorf("error while deleting job: %s", err)
	}
	var jsonData json.RawMessage
	done, doneFn := context.WithCancel(context.TODO())
	server := setupTestServer(t, &jsonData, doneFn)
	defer server.Close()
	exe.SetClient(server.Client())
	jobTriggerWindow := 200
	job = createTestJob(jobTriggerWindow, server.Listener)
	job.ID = strconv.Itoa(rand.Intn(10))
	err = exe.Run(job)
	if err != nil {
		t.Errorf("error while adding job: %s", err)
		return
	}
	// simulates delay between run and del requests
	deadline := time.After(time.Millisecond * time.Duration(jobTriggerWindow*2))
	time.Sleep(time.Millisecond * time.Duration(jobTriggerWindow/2))
	exe.Delete(job.ID)

	select {
	case <-done.Done():
		t.Errorf("couldn't delete job, it got executed")
	case <-deadline:
		t.Logf("time elapsed")
	}

}

func TestUpdatePostponedJob(t *testing.T) {
	testUpdateJob(t, 150, 1000)
}

func TestUpdatePrePonedJob(t *testing.T) {
	testUpdateJob(t, 1000, 500)
}

func testUpdateJob(t *testing.T, oldJobTriggerWindow, newJobTriggerWindow int) {
	var exe = NewJobExecutor()
	var job jobmodels.Job
	err := exe.Update(job.ID, job)
	if err != nil {
		t.Errorf("error while updating job: %s", err)
	}
	var jsonData json.RawMessage
	done, doneFn := context.WithCancel(context.TODO())
	server := setupTestServer(t, &jsonData, doneFn)
	defer server.Close()
	job = createTestJob(oldJobTriggerWindow, server.Listener)
	job.ID = strconv.Itoa(rand.Intn(10))

	updateDone, updateDoneFn := context.WithCancel(context.TODO())
	updatedJobServer := setupTestServer(t, &jsonData, updateDoneFn)
	defer updatedJobServer.Close()
	updatedJob := createTestJob(newJobTriggerWindow, updatedJobServer.Listener)
	updatedJob.ID = job.ID
	err = exe.Run(job)
	if err != nil {
		t.Errorf("error while adding job: %s", err)
		return
	}
	var smallerTime, biggerTime int
	var oldClient, newClient *http.Client
	smallerTime, biggerTime = oldJobTriggerWindow, newJobTriggerWindow
	oldClient, newClient = server.Client(), updatedJobServer.Client()
	if smallerTime > newJobTriggerWindow {
		smallerTime, biggerTime = newJobTriggerWindow, oldJobTriggerWindow
		oldClient, newClient = updatedJobServer.Client(), server.Client()

	}
	go func() { // update client for updated server
		time.Sleep(time.Millisecond * time.Duration(smallerTime))
		exe.SetClient(newClient)
	}()
	exe.SetClient(oldClient)
	// simulates delay between run and del requests
	deadline := time.After(time.Millisecond * time.Duration(biggerTime*2))
	exe.Update(job.ID, updatedJob)

	select {
	case <-done.Done():
		t.Errorf("couldn't update job, out-dated job got executed")
	case <-deadline:
		t.Errorf("time elapsed, updated job didn't ran")
	case <-updateDone.Done():
		t.Logf("job updated and executed, successfully")
	}

}
