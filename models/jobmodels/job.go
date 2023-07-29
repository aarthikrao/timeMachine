package jobmodels

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aarthikrao/timeMachine/utils/time"
)

type Job struct {
	ID string `json:"id,omitempty" bson:"id,omitempty"`

	// Trigger time in milliseconds
	TriggerMS int64           `json:"trigger_ms,omitempty" bson:"trigger_ms,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty" bson:"meta,omitempty"`
	Route     string          `json:"route,omitempty" bson:"route,omitempty"`
}

func (j *Job) Valid() error {
	if j.ID == "" {
		return fmt.Errorf("invalid job id")
	}
	if j.TriggerMS < time.GetCurrentMillis() {
		return fmt.Errorf("trigger_time is in the past")
	}
	if j.Route == "" {
		return fmt.Errorf("invalid route")
	}

	return nil
}

func (j *Job) GetMinuteBucketName() []byte {
	// Get the minutes since epoch
	var jobMinute int64 = j.TriggerMS / 60000

	return []byte(strconv.FormatInt(jobMinute, 10))
}

// returns collection + "_" + job.ID
func (j *Job) GetUniqueKey(collection string) []byte {
	if collection == "" {
		return nil
	}

	return []byte(fmt.Sprintf("%s_%s", collection, j.ID))
}

func (j *Job) StringifyTriggerTime() []byte {
	return []byte(fmt.Sprintf("%d", j.TriggerMS))
}

// TODO: Change to msgpack later
func (j *Job) ToBytes() ([]byte, error) {
	return json.Marshal(&j)
}

// GetJobFromBytes returns the job struct from byte array
// TODO: Change to msgpack later
func GetJobFromBytes(by []byte) (*Job, error) {
	var j Job
	err := json.Unmarshal(by, &j)
	if err != nil {
		return nil, err
	}

	return &j, nil
}
