package executor

import (
	"container/heap"
	"sync"
)

// jobHeap implements heap.Interface
type jobHeap struct {
	entries []*jobEntry
	length  int
	rw      sync.Mutex
}

type jobQueue interface {
	addJob(entry *jobEntry)
	nextJob() *jobEntry
}

func (jq *jobHeap) Less(i, j int) bool {
	return jq.entries[i].job.TriggerMS < jq.entries[j].job.TriggerMS
}

func (jq *jobHeap) Len() int {

	return jq.length
}

func (jq *jobHeap) Swap(i, j int) {

	jq.entries[i], jq.entries[j] = jq.entries[j], jq.entries[i]
}

func (jq *jobHeap) Pop() any {

	jq.length--
	if jq.length < 0 {
		return nil
	}
	return jq.entries[jq.length]
}

func (jq *jobHeap) Push(x any) {
	item := x.(*jobEntry)

	// Oh! We reached at the tip of entries
	// need to expand underlying array
	// TODO: We should squeeze this slice if it's idle and underutilised
	if jq.length+1 < len(jq.entries) {
		jq.entries = append(jq.entries, item)
	} else {
		jq.entries[jq.length] = x.(*jobEntry)
	}
	jq.length++
}

func (jq *jobHeap) nextJob() *jobEntry {
	jq.rw.Lock()
	defer jq.rw.Unlock()
	ientry := heap.Pop(jq)
	if ientry != nil {
		return ientry.(*jobEntry)
	}
	return nil
}

func (jq *jobHeap) addJob(entry *jobEntry) {
	jq.rw.Lock()
	defer jq.rw.Unlock()
	heap.Push(jq, entry)
}
