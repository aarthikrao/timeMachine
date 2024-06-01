package executor

import (
	"container/heap"
	"sync"
)

type jobList []*jobEntry

func (jq jobList) Less(i, j int) bool {
	return jq[i].job.TriggerMS < jq[j].job.TriggerMS
}

func (jq jobList) Len() int {
	return len(jq)
}

func (jq jobList) Swap(i, j int) {
	jq[i], jq[j] = jq[j], jq[i]
}

func (jq *jobList) Pop() any {
	old := *jq
	n := len(old)
	x := old[n-1]
	*jq = old[0 : n-1]
	return x
}

// Push adds an element to the jobHeap.
// If the underlying array is full, it replaces the last element with the new element.
// The length of the jobHeap is incremented by 1.
func (jq *jobList) Push(x interface{}) {
	*jq = append(*jq, x.(*jobEntry))
}

// jobHeap implements heap.Interface
type jobHeap struct {
	entries jobList
	rw      sync.Mutex
}

type JobQueue interface {
	AddJob(entry *jobEntry)
	NextJob() *jobEntry
	Len() int
}

func NewJobHeap() *jobHeap {
	h := &jobHeap{
		entries: make([]*jobEntry, 0),
	}

	heap.Init(&h.entries)

	return h
}

func (jq *jobHeap) NextJob() *jobEntry {
	jq.rw.Lock()
	defer jq.rw.Unlock()

	if jq.entries.Len() == 0 {
		return nil
	}

	ientry := heap.Pop(&jq.entries)
	if ientry != nil {
		return ientry.(*jobEntry)
	}
	return nil
}

func (jq *jobHeap) AddJob(entry *jobEntry) {
	jq.rw.Lock()
	defer jq.rw.Unlock()

	heap.Push(&jq.entries, entry)
}

func (jq *jobHeap) Len() int {
	jq.rw.Lock()
	defer jq.rw.Unlock()

	return jq.entries.Len()
}
