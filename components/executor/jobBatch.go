package executor

import (
	"container/list"
	"sync"
	"time"
)

type jobBatch struct {
	batchDuration, slotDuration time.Duration
	slots                       int
	slotRing                    []list.List
	muRing                      []sync.Mutex
	batchStartTime              time.Time
	rwRingSwap                  sync.RWMutex
}

type jobDispatch func(*list.List) error

func NewJobBatch(batchDuration, slotDuration time.Duration) *jobBatch {

	slotCount := int(batchDuration / slotDuration)
	batch := jobBatch{
		batchDuration:  batchDuration,
		slotDuration:   slotDuration,
		slots:          slotCount,
		batchStartTime: time.Now(),
		slotRing:       make([]list.List, 2*slotCount),
		muRing:         make([]sync.Mutex, 2*slotCount),
	}
	return &batch
}

// add, adds a job version to slot rings
// safe for concurrenct usage
func (jb *jobBatch) add(jov *jobVersion, triggerMs int64) error {
	scheduleTime := time.UnixMilli(triggerMs)
	slot := jb.getSlot(scheduleTime)
	if slot != -1 {
		jb.rwRingSwap.RLock()
		defer jb.rwRingSwap.RUnlock()

		jb.muRing[slot].Lock()
		defer jb.muRing[slot].Unlock()
		jb.slotRing[slot].PushBack(jov)
	}
	return ErrNoSlot
}

// checkRings, checks if current needs to swapped or not
// ideally, batchCount in the following calcuation should not more than 1
// i.e. more than one batch has left until last access
func (jb *jobBatch) checkRings() {
	sinceBatchStart := time.Since(jb.batchStartTime)
	if currentBatchPassed := sinceBatchStart > jb.batchDuration; currentBatchPassed { // current batch is passed
		jb.rwRingSwap.Lock() // provides safety via swapping slot rings
		defer jb.rwRingSwap.Unlock()
		batchCount := sinceBatchStart / jb.batchDuration
		if batchCount > 1 { // since more than one batch has left, we can reset the slot ring
			jb.slotRing = make([]list.List, 2*jb.slots)
		} else { // since only one batch has left, so swapping would help us
			var tempSlots = make([]list.List, jb.slots)
			copy(tempSlots, jb.slotRing[:jb.slots])
			copy(jb.slotRing[:jb.slots], jb.slotRing[jb.slots:])
			copy(jb.slotRing[jb.slots:], tempSlots)
		}
		jb.batchStartTime = jb.batchStartTime.Add(batchCount * jb.batchDuration) // setting latest batch start time for future reference
	}
}

// getSlot, retrives the slot reponsibel for given time
func (jb *jobBatch) getSlot(scheduleTime time.Time) int {
	untilScheduleTime := time.Until(scheduleTime)
	if jb.batchDuration < untilScheduleTime || untilScheduleTime < 0 {
		// Too early or Too late to access batch
		return -1
	}
	jb.checkRings()
	sinceBatchStartTime := scheduleTime.Sub(scheduleTime)
	slotIndex := int(sinceBatchStartTime / jb.slotDuration)
	currentBatch := sinceBatchStartTime <= jb.batchDuration
	if currentBatch {
		return slotIndex
	}
	return jb.slots + slotIndex
}

func (jb *jobBatch) iterateBatch(scheduleTime time.Time, dispatcher jobDispatch) error {
	slot := jb.getSlot(scheduleTime)
	if slot > -1 {
		jb.rwRingSwap.RLock()
		defer jb.rwRingSwap.RUnlock()

		jb.muRing[slot].Lock()
		defer jb.muRing[slot].Unlock()

		return dispatcher(&jb.slotRing[slot])
	}
	return nil
}
