package executor

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
)

const ticksInMinute = 60000

type dispatcher struct {
	tickerBuckets sync.Map

	deleteJobs map[string]struct{}

	deleteLock  sync.RWMutex
	timerLocker sync.RWMutex

	now            int64
	milisecondTick int64

	ticker     chan int64
	dispatchCh chan *jobmodels.Job
}

var (
	ErrTooLate  = errors.New("Too late")
	ErrTooEarly = errors.New("Too early")
)

func New(dispatchCh chan *jobmodels.Job) Dispatcher {
	dis := &dispatcher{
		deleteJobs: make(map[string]struct{}),
		ticker:     make(chan int64),
		dispatchCh: dispatchCh,
	}
	go dis.startTimer()
	go dis.dispatcher()

	return dis
}

func (e *dispatcher) getCurrentTime() int64 {
	e.timerLocker.RLock()
	defer e.timerLocker.RUnlock()

	return e.now
}

func (e *dispatcher) initTimer() {
	e.timerLocker.Lock()
	defer e.timerLocker.Unlock()

	e.now = time.Now().UnixMilli()
	e.milisecondTick = 0
}

func (e *dispatcher) incrementTick() int64 {
	e.timerLocker.Lock()
	defer e.timerLocker.Unlock()

	e.now++
	e.milisecondTick++
	nextTick := e.milisecondTick

	return nextTick

}

func (e *dispatcher) startTimer() {
	e.initTimer()
	for {
		time.Sleep(time.Millisecond)
		e.ticker <- e.incrementTick()
	}
}

func (e *dispatcher) Add(job jobmodels.Job) error {
	tickDiff := job.TriggerMS - e.getCurrentTime()
	if tickDiff <= 0 {
		return ErrTooLate
	}
	if tickDiff > 2*ticksInMinute {
		return ErrTooEarly
	}

	e.addJob(tickDiff, job)

	return nil
}

func (e *dispatcher) Delete(jobId string) {
	e.deleteLock.Lock()
	defer e.deleteLock.Unlock()

	e.deleteJobs[jobId] = struct{}{}

}

func (e *dispatcher) dispatcher() {
	for tick := range e.ticker {
		go e.dispatchJobs(tick)
	}
}
func (e *dispatcher) dispatchJobs(tick int64) {
	bucket := e.retrive(tick)
	for bucket != nil && bucket.Len() != 0 {
		ele := bucket.Front()
		if ele != nil {

			val := bucket.Remove(ele)

			job, ok := val.(*jobmodels.Job)

			if ok {
				if e.shouldSkipJob(job.ID) {
					e.removeJob(job.ID)
					continue
				}
				e.dispatchCh <- job
			}
		}
	}
}

func (e *dispatcher) removeJob(id string) {
	e.deleteLock.Lock()
	defer e.deleteLock.Unlock()

	delete(e.deleteJobs, id)

}

func (e *dispatcher) shouldSkipJob(id string) bool {
	e.deleteLock.RLock()
	defer e.deleteLock.RUnlock()

	_, shouldSkip := e.deleteJobs[id]
	return shouldSkip
}

func (e *dispatcher) addJob(tickDiff int64, job jobmodels.Job) {
	bucket, _ := e.tickerBuckets.LoadOrStore(tickDiff, list.New())
	bucket.(*list.List).PushBack(job)
}

func (e *dispatcher) retrive(tick int64) *list.List {
	val, ok := e.tickerBuckets.LoadAndDelete(tick)
	if ok {
		return val.(*list.List)
	}
	return nil
}
