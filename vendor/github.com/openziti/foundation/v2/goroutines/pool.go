package goroutines

import (
	"fmt"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"
)

type strErr string

func (s strErr) Error() string {
	return string(s)
}

const (
	TimeoutError     = strErr("timed out")
	QueueFullError   = strErr("queue full")
	PoolStoppedError = strErr("pool shutdown")
)

// Pool represents a goroutine worker pool that can be configured with a queue size and min and max sizes.
//
//	The pool will start with min size goroutines and will add more if the queue isn't staying empty.
//	After a worker has been idle for a configured time, it will stop
type Pool interface {
	// Queue submits a unit of work to the pool. It will return an error if the pool is shutdown
	Queue(func()) error

	// QueueWithTimeout submits a unit of work to the pool. It will return an error if the pool is shutdown or
	// if the work cannot be submitted to the work queue before the given timeout elapses
	QueueWithTimeout(func(), time.Duration) error

	// QueueOrError submits a unit of work to the pool. It will return an error if the pool is shutdown or
	// if the work cannot be submitted to the work queue immediately
	QueueOrError(func()) error

	// GetWorkerCount returns the current number of goroutines servicing the work queue
	GetWorkerCount() uint32

	// GetQueueSize returns the current number of work items in the work queue
	GetQueueSize() uint32

	// GetBusyWorkers returns the current number workers busy doing work from the work queue
	GetBusyWorkers() uint32

	// Shutdown stops all workers as they finish work and prevents new work from being submitted to the queue
	Shutdown()
}

// PoolConfig is used to configure a new Pool
type PoolConfig struct {
	// The size of the channel feeding the worker pool
	QueueSize uint32
	// The minimum number of goroutines
	MinWorkers uint32
	// The maximum number of workers
	MaxWorkers uint32
	// IdleTime how long a goroutine should be idle before exiting
	IdleTime time.Duration
	// Provides a way to join shutdown of the pool with other components.
	// The pool also be shut down independently using the Shutdown method
	CloseNotify <-chan struct{}
	// Provides a way to specify what happens if a worker encounters a panic
	// if no PanicHandler is provided, panics will not be caught
	PanicHandler func(err interface{})
	// Optional callback which is called whenever work completes, with the
	// time the work took to complete
	OnWorkCallback func(workTime time.Duration)
	// Optional callback which is called when the pool is created
	OnCreate func(Pool)
}

func (self *PoolConfig) Validate() error {
	if self.MaxWorkers < 1 {
		return fmt.Errorf("max workers must be at least 1")
	}
	if self.MinWorkers > self.MaxWorkers {
		return fmt.Errorf("min workers must be less than or equal to max workers. min workers=%v, max workers=%v", self.MinWorkers, self.MaxWorkers)
	}
	return nil
}

func NewPool(config PoolConfig) (Pool, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	result := &pool{
		queue:               make(chan func(), int(config.QueueSize)),
		minWorkers:          config.MinWorkers,
		maxWorkers:          config.MaxWorkers,
		maxIdle:             config.IdleTime,
		externalCloseNotify: config.CloseNotify,
		closeNotify:         make(chan struct{}),
		panicHandler:        config.PanicHandler,
		onWorkCallback:      config.OnWorkCallback,
	}
	if config.OnCreate != nil {
		config.OnCreate(result)
	}
	for result.count < result.minWorkers {
		result.tryAddWorker()
	}
	return result, nil
}

type pool struct {
	queue               chan func()
	queueSize           uint32
	count               uint32
	minWorkers          uint32
	maxWorkers          uint32
	busyWorkers         uint32
	maxIdle             time.Duration
	stopped             atomic.Bool
	externalCloseNotify <-chan struct{}
	closeNotify         chan struct{}
	panicHandler        func(err interface{})
	onWorkCallback      func(workTime time.Duration)
}

func (self *pool) Queue(work func()) error {
	return self.queueImpl(work, nil)
}

func (self *pool) QueueWithTimeout(work func(), timeout time.Duration) error {
	return self.queueImpl(work, time.After(timeout))
}

func (self *pool) queueImpl(work func(), timeoutC <-chan time.Time) error {
	self.ensureNoStarvation()
	select {
	case self.queue <- work:
		self.incrQueueSize()
		self.ensureNoStarvation()
		return nil
	case <-self.closeNotify:
		return errors.Wrap(PoolStoppedError, "cannot queue")
	case <-self.externalCloseNotify:
		return errors.Wrap(PoolStoppedError, "cannot queue, pool stopped externally")
	case <-timeoutC:
		return errors.Wrap(TimeoutError, "cannot queue")
	}
}

func (self *pool) QueueOrError(work func()) error {
	select {
	case self.queue <- work:
		self.incrQueueSize()
		self.ensureNoStarvation()
		return nil
	case <-self.closeNotify:
		return errors.Wrap(PoolStoppedError, "cannot queue")
	case <-self.externalCloseNotify:
		return errors.Wrap(PoolStoppedError, "cannot queue, pool stopped externally")
	default:
		return errors.Wrap(QueueFullError, "cannot queue")
	}
}

func (self *pool) ensureNoStarvation() {
	if self.minWorkers == 0 && self.GetWorkerCount() == 0 {
		self.tryAddWorker()
	}
}

func (self *pool) Shutdown() {
	if self.stopped.CompareAndSwap(false, true) {
		close(self.closeNotify)
	}
}

func (self *pool) worker(initialWork func()) {
	defer func() {
		if err := recover(); err != nil {
			if self.panicHandler != nil {
				self.panicHandler(err)
			} else {
				fmt.Printf("panic during pool worker executing (%+v)\n", err)
			}
			self.tryAddWorker()
		}
	}()

	defer func() {
		// There's a small race condition where the last worker can exit due to idle
		// right as something is queued. If we're the last worker, check again, just
		// to be sure there's nothing queued.
		if newCount := self.decrementCount(); newCount == 0 {
			time.AfterFunc(100*time.Millisecond, self.startExtraWorkerIfQueueBusy)
		}
	}()

	if initialWork != nil {
		self.runWork(initialWork)
	}

	for {
		select {
		case work := <-self.queue:
			self.decrQueueSize()
			self.startExtraWorkerIfQueueBusy()
			self.runWork(work)
		case <-time.After(self.maxIdle):
			if self.GetWorkerCount() > self.minWorkers {
				return
			}
		case <-self.closeNotify:
			return
		case <-self.externalCloseNotify:
			return
		}
	}
}

func (self *pool) startExtraWorkerIfQueueBusy() {
	if self.GetWorkerCount() < self.maxWorkers {
		if self.incrementCount() <= self.maxWorkers {
			select {
			case work := <-self.queue:
				self.decrQueueSize()
				go self.worker(work)
			default:
				self.decrementCount()
			}
		} else {
			self.decrementCount()
		}
	}
}

func (self *pool) tryAddWorker() {
	if self.GetWorkerCount() < self.maxWorkers {
		if self.incrementCount() <= self.maxWorkers {
			go self.worker(nil)
		} else {
			self.decrementCount()
		}
	}
}

func (self *pool) runWork(work func()) {
	self.incrBusyWorkers()
	defer self.decrBusyWorkers()
	if self.onWorkCallback != nil {
		start := time.Now()
		work()
		self.onWorkCallback(time.Since(start))
	} else {
		work()
	}
}

func (self *pool) GetWorkerCount() uint32 {
	return atomic.LoadUint32(&self.count)
}

func (self *pool) incrementCount() uint32 {
	return atomic.AddUint32(&self.count, 1)
}

func (self *pool) decrementCount() uint32 {
	return atomic.AddUint32(&self.count, ^uint32(0))
}

func (self *pool) GetQueueSize() uint32 {
	return atomic.LoadUint32(&self.queueSize)
}

func (self *pool) incrQueueSize() uint32 {
	return atomic.AddUint32(&self.queueSize, 1)
}

func (self *pool) decrQueueSize() uint32 {
	return atomic.AddUint32(&self.queueSize, ^uint32(0))
}

func (self *pool) GetBusyWorkers() uint32 {
	return atomic.LoadUint32(&self.busyWorkers)
}

func (self *pool) incrBusyWorkers() uint32 {
	return atomic.AddUint32(&self.busyWorkers, 1)
}

func (self *pool) decrBusyWorkers() uint32 {
	return atomic.AddUint32(&self.busyWorkers, ^uint32(0))
}
