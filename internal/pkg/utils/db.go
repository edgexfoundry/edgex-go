package utils

import "time"

// MakeTimestamp returns the Unix timestamp in milliseconds
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
