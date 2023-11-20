package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

const contentType = "application/json"

type ExecutorOptions struct {
	RequestTimeout  time.Duration
	IdleConnTimeout time.Duration
	// Go-Routine pool size
	PoolSize int
}
type httpReq struct {
	route       string
	requestBody json.RawMessage
}
type executorPool struct {
	timeout    time.Duration
	headers    http.Header
	maxSize    int
	cursize    int
	pendingReq chan httpReq
	mu         *sync.Mutex
}

func newpool(requestTimeout, idleConnTimeout time.Duration, poolSize int) *executorPool {
	var h = make(http.Header)

	h.Set("keep-alive", fmt.Sprintf("timeout=%d", int(idleConnTimeout.Seconds())))
	h.Set("content-type", contentType)

	var pool = &executorPool{
		timeout:    requestTimeout,
		headers:    h,
		maxSize:    poolSize,
		pendingReq: make(chan httpReq),
		mu:         new(sync.Mutex),
	}
	return pool
}
func (ep *executorPool) run(route string, requestBody json.RawMessage) {
	ep.mu.Lock()
	if ep.cursize < ep.maxSize {
		ep.cursize++
		go ep.startDispatcher()
	}
	ep.mu.Unlock()
	ep.pendingReq <- httpReq{route, requestBody}
}
func (ep *executorPool) close() {
	close(ep.pendingReq)
}

func (ep *executorPool) startDispatcher() {

	// we also need to make idle connection timeout tweakable
	var cli = http.DefaultClient

	cli.Timeout = ep.timeout

	for req := range ep.pendingReq {
		var hreq, err = http.NewRequest(http.MethodPost, req.route, bytes.NewReader(req.requestBody))
		if err != nil {
			// TODO: report this error somewhere
			continue
		}
		hreq.Header = ep.headers
		resp, err := cli.Do(hreq)
		if err != nil {
			// TODO: report this error somewhere
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			// TODO: report this error somewhere
			continue

		}
	}
}

// StartExecutionPool starts executor pool, it can be closed by cancelling the passed context
// This needs to be run in a seperate go-routine
func StartExecutionPool(ctx context.Context, dispatchQueue DispatchQueue, options ExecutorOptions) {
	var pool = newpool(options.RequestTimeout, options.IdleConnTimeout, options.PoolSize)
	var nextJob jobmodels.Job
	for {
		select {
		case <-ctx.Done():
			pool.close()
			return
		default:
			if dispatchQueue.Next(&nextJob) {
				pool.run(nextJob.Route, nextJob.Meta) // this unblocks as soon as we have aleast one go-routine to serve the request
			} else {
				time.Sleep(time.Millisecond) // since we didn't find anything in dispatch, we are waiting for new jobs
			}

		}
	}
}
