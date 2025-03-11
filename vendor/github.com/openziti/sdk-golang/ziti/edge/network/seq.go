package network

import (
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"
)

var ErrClosed = errors.New("sequencer closed")

type ReadTimout struct{}

func (r ReadTimout) Error() string {
	return "read timed out"
}

func (r ReadTimout) Timeout() bool {
	return true
}

func (r ReadTimout) Temporary() bool {
	return true
}

func NewNoopSequencer[T any](channelDepth int) *noopSeq[T] {
	return &noopSeq[T]{
		ch:             make(chan T, channelDepth),
		closeNotify:    make(chan struct{}),
		deadlineNotify: make(chan struct{}),
	}
}

type noopSeq[T any] struct {
	ch             chan T
	closeNotify    chan struct{}
	deadlineNotify chan struct{}
	deadline       concurrenz.AtomicValue[time.Time]
	closed         atomic.Bool
	readInProgress atomic.Bool
}

func (seq *noopSeq[T]) PutSequenced(event T) error {
	select {
	case seq.ch <- event:
		return nil
	case <-seq.closeNotify:
		return ErrClosed
	}
}

func (seq *noopSeq[T]) SetReadDeadline(deadline time.Time) {
	seq.deadline.Store(deadline)
	if seq.readInProgress.Load() {
		select {
		case seq.deadlineNotify <- struct{}{}:
		case <-time.After(5 * time.Millisecond):
		}
	} else {
		select {
		case seq.deadlineNotify <- struct{}{}:
		default:
		}
	}
}

func (seq *noopSeq[T]) GetNext() (T, error) {
	seq.readInProgress.Store(true)
	defer seq.readInProgress.Store(false)

	var val T

	for {
		deadline := seq.deadline.Load()

		var timeoutCh <-chan time.Time

		if !deadline.IsZero() {
			timeoutCh = time.After(time.Until(deadline))
		}

		select {
		case val = <-seq.ch:
			return val, nil
		case <-seq.closeNotify:
			// If we're closed, return any buffered values, otherwise return nil
			select {
			case val = <-seq.ch:
				return val, nil
			default:
				return val, ErrClosed
			}
		case <-seq.deadlineNotify:
			continue
		case <-timeoutCh:
			// If we're timing out, return any buffered values, otherwise return nil
			select {
			case val = <-seq.ch:
				return val, nil
			default:
				return val, &ReadTimout{}
			}
		}
	}
}

func (seq *noopSeq[T]) Close() {
	if seq.closed.CompareAndSwap(false, true) {
		close(seq.closeNotify)
	}
}
