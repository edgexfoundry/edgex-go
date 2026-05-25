package channel

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/pkg/errors"
	"io"
	"sync/atomic"
	"time"
)

var ErrClosed = errors.New("channel closed")

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

func NewReadAdapter(label string, channelDepth int) *ReadAdapter {
	return &ReadAdapter{
		label:          label,
		ch:             make(chan []byte, channelDepth),
		closeNotify:    make(chan struct{}),
		deadlineNotify: make(chan struct{}),
	}
}

type ReadAdapter struct {
	label          string
	ch             chan []byte
	closeNotify    chan struct{}
	deadlineNotify chan struct{}
	deadline       concurrenz.AtomicValue[time.Time]
	closed         atomic.Bool
	readInProgress atomic.Bool
	leftover       []byte
}

func (self *ReadAdapter) PushData(data []byte) error {
	select {
	case self.ch <- data:
		return nil
	case <-self.closeNotify:
		return ErrClosed
	}
}

func (self *ReadAdapter) SetReadDeadline(deadline time.Time) error {
	self.deadline.Store(deadline)
	if self.readInProgress.Load() {
		select {
		case self.deadlineNotify <- struct{}{}:
		case <-time.After(5 * time.Millisecond):
		}
	} else {
		select {
		case self.deadlineNotify <- struct{}{}:
		default:
		}
	}
	return nil
}

func (self *ReadAdapter) GetNext() ([]byte, error) {
	self.readInProgress.Store(true)
	defer self.readInProgress.Store(false)

	for {
		deadline := self.deadline.Load()

		var timeoutCh <-chan time.Time

		if !deadline.IsZero() {
			timeoutCh = time.After(time.Until(deadline))
		}

		select {
		case data := <-self.ch:
			return data, nil
		case <-self.closeNotify:
			// If we're closed, return any buffered values, otherwise return nil
			select {
			case data := <-self.ch:
				return data, nil
			default:
				return nil, ErrClosed
			}
		case <-self.deadlineNotify:
			continue
		case <-timeoutCh:
			// If we're timing out, return any buffered values, otherwise return nil
			select {
			case data := <-self.ch:
				return data, nil
			default:
				return nil, &ReadTimout{}
			}
		}
	}
}

func (self *ReadAdapter) Read(b []byte) (n int, err error) {
	log := pfxlog.Logger().WithField("label", self.label)
	if self.closed.Load() {
		return 0, io.EOF
	}

	log.Tracef("read buffer = %d bytes", len(b))
	if len(self.leftover) > 0 {
		log.Tracef("found %d leftover bytes", len(self.leftover))
		n = copy(b, self.leftover)
		self.leftover = self.leftover[n:]
		return n, nil
	}

	d, err := self.GetNext()
	if err != nil {
		return 0, err
	}
	log.Tracef("got buffer from sequencer %d bytes", len(d))

	n = copy(b, d)
	self.leftover = d[n:]

	log.Tracef("saving %d bytes for leftover", len(self.leftover))
	log.Tracef("reading %v bytes", n)
	return n, nil
}

func (self *ReadAdapter) Close() {
	if self.closed.CompareAndSwap(false, true) {
		close(self.closeNotify)
	}
}
