package time

import "time"

// GetCurrentMillis returns the milliseconds since epoch
func GetCurrentMillis() int64 {
	return time.Now().UnixNano() / time.Millisecond.Nanoseconds()
}

// GetCurrentMinutes returns the minutes since epoch
func GetCurrentMinutes() int64 {
	return GetCurrentMillis() / 60000
}

func GetEpochMinutes(timestampInMS int64) int {
	return int(timestampInMS) / 60000
}
