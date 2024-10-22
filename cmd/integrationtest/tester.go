// This package contains tools for basic testing of time machine DB
//
// It contains a client that publishes jobs to the time machine DB,
// and a server which listens to the callback even
//
// Usage: go run cmd/tester/tester.go -c 10
// Usage: ./tester -c 10
// This will create 10 jobs in the time machine DB
//
// The server listens on port 4000 for callbacks
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/utils/httpclient"
	timeUtils "github.com/aarthikrao/timeMachine/utils/time"
)

var (
	// Route ID for this service
	TesterRoute = "TesterRoute"

	// URLs
	JobURL   = "http://localhost:8001/job/test"
	RouteURL = "http://localhost:8001/route"
)

type Tester struct {
	client *httpclient.HTTPClient

	mu    sync.Mutex
	count int
}

func (t *Tester) IncrementRecievedCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.count++
	return t.count
}

func main() {
	count := flag.Int("c", 0, "number of jobs to create")
	delay := flag.Int("d", 60000, "delay in milliseconds for job trigger time")
	routines := flag.Int("r", 10, "number of routines to create jobs")
	flag.Parse()

	if count == nil || *count == 0 {
		flag.PrintDefaults()
		fmt.Println("Please provide a valid count")
		os.Exit(1)
	}

	t := &Tester{
		client: httpclient.NewHTTPClient(30*time.Second, 10),
		count:  0,
	}

	srv := &http.Server{
		Addr: ":4000",
	}
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
		wg.Done()
	}()

	go func() {
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			// Handle callback logic here
			var payload map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)

			callbackCount := t.IncrementRecievedCount()

			fmt.Printf("Count: %d Recieved callback:, %v \n", callbackCount, payload)
			if callbackCount == *count {
				fmt.Println("All callbacks recieved")
				wg.Done()
			}

		})

		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Create a route
	t.createRoute()

	time.Sleep(2 * time.Second)

	// Create jobs
	t.createJobs(*count, *delay, *routines)

	fmt.Println("Listening for callbacks on port 4000... Press Ctrl + C to exit")

	wg.Wait()

	err := srv.Shutdown(context.Background())
	if err != nil {
		fmt.Println("Error shutting down server:", err)
	}

	fmt.Println("Exiting...")
}

func (t *Tester) createRoute() {
	// Create a route
	route := routemodels.Route{
		ID:         TesterRoute,
		Type:       routemodels.Http,
		WebhookURL: "http://localhost:4000/callback",
	}

	by, statusCode, err := t.client.Post(RouteURL, route)
	if err != nil {
		fmt.Println("Error creating route:", err)
		return
	}
	if statusCode != http.StatusOK {
		fmt.Println("Error creating route. Status code:", statusCode, string(by))
		return
	}
	fmt.Println("Created route:", string(by))

}

func (t *Tester) createJobs(count int, delayMS int, routines int) {
	var wg sync.WaitGroup
	jobChan := make(chan int, count)

	// Start goroutines
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobChan {
				var job jobmodels.Job = jobmodels.Job{
					ID:        fmt.Sprintf("job-%d", i),
					TriggerMS: timeUtils.GetCurrentMillis() + delayMS,
					Meta:      []byte(`{"foo": "bar"}`),
					Route:     TesterRoute,
				}

				by, statusCode, err := t.client.Post(JobURL, job)
				if err != nil {
					fmt.Println("Error creating job:", err)
					continue
				}
				if statusCode != http.StatusOK {
					fmt.Println("Error creating job. Status code:", statusCode, string(by))
					continue
				}
				fmt.Println("Created job", i, " : ", string(by), " with payload: ", job)
			}
		}()
	}

	// Send jobs to the jobChan
	for i := 0; i < count; i++ {
		jobChan <- i
	}
	close(jobChan)

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Printf("Created %d jobs...\n", count)
}
