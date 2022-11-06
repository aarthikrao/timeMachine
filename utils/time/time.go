package time

import "time"

// GetCurrentMillis returns the milliseconds since epoch
func GetCurrentMillis() int {
	return int(time.Now().UnixNano()) / int(time.Millisecond)
}
