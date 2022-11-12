package jobmodels

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aarthikrao/timeMachine/utils/time"
)

type Job struct {
	ID          string          `json:"id,omitempty" bson:"id,omitempty"`
	TriggerTime int             `json:"trigger_time,omitempty" bson:"trigger_time,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty" bson:"meta,omitempty"`
	Route       string          `json:"route,omitempty" bson:"route,omitempty"`
}

func (j *Job) Valid() error {
	if j.ID == "" {
		return fmt.Errorf("invalid job id")
	}
	if j.TriggerTime < time.GetCurrentMillis() {
		return fmt.Errorf("trigger_time is in the past")
	}
	if j.Route == "" {
		return fmt.Errorf("invalid route")
	}

	return nil
}

func (j *Job) GetMinuteBucketName() []byte {
	// Get the minutes since epoch
	jobMinute := j.TriggerTime / 1000

	return []byte(strconv.Itoa(jobMinute))
}

// returns collection + "_" + job.ID
func (j *Job) GetUniqueKey(collection string) []byte {
	if collection == "" {
		return nil
	}

	return []byte(fmt.Sprintf("%s_%s", collection, j.ID))
}

func (j *Job) StringifyTriggerTime() []byte {
	return []byte(fmt.Sprintf("%d", j.TriggerTime))
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
