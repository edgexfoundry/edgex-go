package channel

import "time"

const (
	DefaultOutstandingConnects = 16
	DefaultQueuedConnects      = 1
	DefaultConnectTimeout      = 5 * time.Second

	MinQueuedConnects      = 1
	MinOutstandingConnects = 1
	MinConnectTimeout      = 30 * time.Millisecond

	MaxQueuedConnects      = 5000
	MaxOutstandingConnects = 1000
	MaxConnectTimeout      = 60000 * time.Millisecond

	DefaultOutQueueSize = 4
)
