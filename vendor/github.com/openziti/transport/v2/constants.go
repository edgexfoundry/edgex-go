package transport

import "time"

const (
	DefaultWsWriteTimeout      = time.Second * 10
	DefaultWsReadTimeout       = time.Second * 5
	DefaultWsIdleTimeout       = time.Second * 120
	DefaultWsPongTimeout       = time.Second * 60
	DefaultWsPingInterval      = (DefaultWsPongTimeout * 9) / 10
	DefaultWsHandshakeTimeout  = time.Second * 10
	DefaultWsReadBufferSize    = 4096
	DefaultWsWriteBufferSize   = 4096
	DefaultWsEnableCompression = true
)
