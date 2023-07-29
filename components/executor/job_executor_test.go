package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

func setupTestServer(t *testing.T, jsonData *json.RawMessage,
	doneFn context.CancelFunc) *http.Server {
	var server = http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer doneFn()
			var err error
			*jsonData, err = io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("error while reading request body: %v", err)
			}
		}),
		Addr: ":9415",
	}
	go server.ListenAndServe()
	return &server
}

func createTestJob(triggerWindowMs int) jobmodels.Job {
	testJob := jobmodels.Job{Route: "http://localhost:9415/testJob"}
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
	testJob := createTestJob(50)
	randomMilisec := time.Duration(rand.Intn(50))
	schduleTime := time.Now().Add(randomMilisec * time.Millisecond)
	testJob.TriggerMS = schduleTime.UnixMilli()

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

}
