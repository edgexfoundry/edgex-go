package xgress

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/goroutines"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"time"
)

type PayloadIngester struct {
	pool goroutines.Pool
}

func NewPayloadIngester(closeNotify <-chan struct{}) *PayloadIngester {
	return NewPayloadIngesterWithConfig(1, closeNotify)
}

func NewPayloadIngesterWithConfig(maxWorkers uint32, closeNotify <-chan struct{}) *PayloadIngester {
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	poolConfig := goroutines.PoolConfig{
		QueueSize:   uint32(64),
		MinWorkers:  1,
		MaxWorkers:  maxWorkers,
		IdleTime:    30 * time.Second,
		CloseNotify: closeNotify,
		PanicHandler: func(err interface{}) {
			pfxlog.Logger().WithField(logrus.ErrorKey, err).WithField("backtrace", string(debug.Stack())).Error("panic during payload ingest")
		},
		WorkerFunction: payloadIngesterWorker,
	}

	pool, err := goroutines.NewPool(poolConfig)
	if err != nil {
		panic(fmt.Errorf("error creating payload ingester handler pool (%w)", err))
	}

	pi := &PayloadIngester{
		pool: pool,
	}

	return pi
}

func payloadIngesterWorker(_ uint32, f func()) {
	f()
}

func (self *PayloadIngester) ingest(payload *Payload, x *Xgress) {
	_ = self.pool.Queue(func() {
		x.acceptPayload(payload)
	})
}
