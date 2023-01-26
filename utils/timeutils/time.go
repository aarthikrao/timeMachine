package timeutils

import "time"

// GetCurrentMillis returns the milliseconds since epoch
func GetCurrentMillis() int64 {
	return time.Now().UnixMilli()
}
