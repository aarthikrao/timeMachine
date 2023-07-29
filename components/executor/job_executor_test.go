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

func TestAddJob(t *testing.T) {
	var exe = NewJobExecutor()
	err := exe.Run(jobmodels.Job{})
	if err != nil {
		t.Errorf("error occured while running the job %v", err)
	}

	testJob := jobmodels.Job{Route: "http://localhost:9415/testJob"}
	randomMilisec := time.Duration(rand.Intn(5000))
	randomData := strconv.Itoa(rand.Intn(100000))
	testJob.Meta = json.RawMessage(randomData)

	var reqMetaData json.RawMessage
	done, doneFn := context.WithCancel(context.TODO())
	var server = http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer doneFn()
			reqMetaData, err = io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("error while reading request body: %v", err)
			}
		}),
		Addr: ":9415",
	}
	go server.ListenAndServe()
	defer server.Close()

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
