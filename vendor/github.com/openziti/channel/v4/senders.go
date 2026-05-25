package channel

import (
	"errors"
	"fmt"
)

type SenderContext interface {
	NextSequence() int32
	GetCloseNotify() chan struct{}
}

func NewSingleChSender(ctx SenderContext, msgC chan<- Sendable) Sender {
	return &singleChSender{ctx: ctx, msgC: msgC}
}

type singleChSender struct {
	ctx  SenderContext
	msgC chan<- Sendable
}

func (self *singleChSender) CloseNotify() <-chan struct{} {
	return self.ctx.GetCloseNotify()
}

func (self *singleChSender) TrySend(s Sendable) (bool, error) {
	if err := s.Context().Err(); err != nil {
		return false, err
	}

	s.SetSequence(self.ctx.NextSequence())

	select {
	case <-s.Context().Done():
		if err := s.Context().Err(); err != nil {
			return false, TimeoutError{error: fmt.Errorf("timeout waiting to put message in send queue (%w)", err)}
		}
		return false, TimeoutError{error: errors.New("timeout waiting to put message in send queue")}
	case <-self.ctx.GetCloseNotify():
		return false, ClosedError{}
	case self.msgC <- s:
		return true, nil
	default:
		return false, nil
	}
}

func (self *singleChSender) Send(s Sendable) error {
	if err := s.Context().Err(); err != nil {
		return err
	}

	s.SetSequence(self.ctx.NextSequence())

	select {
	case <-s.Context().Done():
		if err := s.Context().Err(); err != nil {
			return TimeoutError{error: fmt.Errorf("timeout waiting to put message in send queue (%w)", err)}
		}
		return TimeoutError{error: errors.New("timeout waiting to put message in send queue")}
	case <-self.ctx.GetCloseNotify():
		return ClosedError{}
	case self.msgC <- s:
	}
	return nil
}
