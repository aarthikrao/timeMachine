package time

import "time"

// GetCurrentMillis returns the milliseconds since epoch
func GetCurrentMillis() int {
	return int(time.Now().UnixNano()) / int(time.Millisecond)
}

// GetCurrentMinutes returns the minutes since epoch
func GetCurrentMinutes() int {
	return GetCurrentMillis() / 60000
}

func GetEpochMinutes(timestampInMS int64) int {
	return int(timestampInMS) / 60000
}
