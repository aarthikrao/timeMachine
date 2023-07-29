package time

import "time"

// GetCurrentMillis returns the milliseconds since epoch
func GetCurrentMillis() int64 {
	return time.Now().UnixMilli()
}

// GetCurrentMinutes returns the minutes since epoch
func GetCurrentMinutes() int64 {
	return GetCurrentMillis() / 60000
}
