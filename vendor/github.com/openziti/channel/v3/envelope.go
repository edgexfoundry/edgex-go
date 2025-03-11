package channel

import (
	"context"
	"github.com/michaelquigley/pfxlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type priorityEnvelopeImpl struct {
	msg *Message
	p   Priority
}

func (self *priorityEnvelopeImpl) SetSequence(seq int32) {
	self.msg.SetSequence(seq)
}

func (self *priorityEnvelopeImpl) Sequence() int32 {
	return self.msg.sequence
}

func (self *priorityEnvelopeImpl) Send(ch Channel) error {
	return ch.Send(self)
}

func (self *priorityEnvelopeImpl) ReplyTo(msg *Message) Envelope {
	self.msg.ReplyTo(msg)
	return self
}

func (self *priorityEnvelopeImpl) Msg() *Message {
	return self.msg
}

func (self *priorityEnvelopeImpl) Context() context.Context {
	return self.msg.Context()
}

func (self *priorityEnvelopeImpl) SendListener() SendListener {
	return self.msg.SendListener()
}

func (self *priorityEnvelopeImpl) ReplyReceiver() ReplyReceiver {
	return nil
}

func (self *priorityEnvelopeImpl) ToSendable() Sendable {
	return self
}

func (self *priorityEnvelopeImpl) Priority() Priority {
	return self.p
}

func (self *priorityEnvelopeImpl) WithPriority(p Priority) Envelope {
	self.p = p
	return self
}

func (self *priorityEnvelopeImpl) WithTimeout(duration time.Duration) TimeoutEnvelope {
	ctx, cancelF := context.WithTimeout(context.Background(), duration)
	return &envelopeImpl{
		msg:     self.msg,
		p:       self.p,
		context: ctx,
		cancelF: cancelF,
	}
}

type envelopeImpl struct {
	msg     *Message
	p       Priority
	context context.Context
	cancelF context.CancelFunc
}

func (self *envelopeImpl) SetSequence(seq int32) {
	self.msg.SetSequence(seq)
}

func (self *envelopeImpl) Sequence() int32 {
	return self.msg.sequence
}

func (self *envelopeImpl) Msg() *Message {
	return self.msg
}

func (self *envelopeImpl) ReplyTo(msg *Message) Envelope {
	self.msg.ReplyTo(msg)
	return self
}

func (self *envelopeImpl) ReplyReceiver() ReplyReceiver {
	return nil
}

func (self *envelopeImpl) ToSendable() Sendable {
	return self
}

func (self *envelopeImpl) SendListener() SendListener {
	return self
}

func (self *envelopeImpl) NotifyQueued() {}

func (self *envelopeImpl) NotifyBeforeWrite() {}

func (self *envelopeImpl) NotifyAfterWrite() {
	if self.cancelF != nil {
		self.cancelF()
	}
}

func (self *envelopeImpl) NotifyErr(error) {}

func (self *envelopeImpl) Priority() Priority {
	return self.p
}

func (self *envelopeImpl) WithPriority(p Priority) Envelope {
	self.p = p
	return self
}

func (self *envelopeImpl) Context() context.Context {
	return self.context
}

func (self *envelopeImpl) WithTimeout(duration time.Duration) TimeoutEnvelope {
	parent := self.context
	if parent == nil {
		parent = context.Background()
	}
	self.context, self.cancelF = context.WithTimeout(parent, duration)
	return self
}

func (self *envelopeImpl) Send(ch Channel) error {
	return ch.Send(self)
}

func (self *envelopeImpl) SendAndWaitForWire(ch Channel) error {
	waitSendContext := &sendWaitEnvelope{envelopeImpl: self}
	return waitSendContext.WaitForWire(ch)
}

func (self *envelopeImpl) SendForReply(ch Channel) (*Message, error) {
	replyContext := &replyEnvelope{envelopeImpl: self}
	return replyContext.WaitForReply(ch)
}

type sendWaitEnvelope struct {
	*envelopeImpl
	errC chan error
}

func (self *sendWaitEnvelope) ToSendable() Sendable {
	return self
}

func (self *sendWaitEnvelope) SendListener() SendListener {
	return self
}

func (self *sendWaitEnvelope) NotifyAfterWrite() {
	close(self.errC)
}

func (self *sendWaitEnvelope) NotifyErr(err error) {
	self.errC <- err
}

func (self *sendWaitEnvelope) WaitForWire(ch Channel) error {
	if err := self.context.Err(); err != nil {
		return err
	}

	defer self.cancelF()

	self.errC = make(chan error, 1)

	if err := ch.Send(self); err != nil {
		return err
	}
	select {
	case err := <-self.errC:
		return err
	case <-self.context.Done():
		if err := self.context.Err(); err != nil {
			return TimeoutError{errors.Wrap(err, "timeout waiting for message to be written to wire")}
		}
		return errors.New("timeout waiting for message to be written to wire")
	}
}

type replyEnvelope struct {
	*envelopeImpl
	errC   chan error
	replyC chan *Message
}

func (self *replyEnvelope) ToSendable() Sendable {
	return self
}

func (self *replyEnvelope) SendListener() SendListener {
	return self
}

func (self *replyEnvelope) ReplyReceiver() ReplyReceiver {
	return self
}

func (self *replyEnvelope) NotifyAfterWrite() {}

func (self *replyEnvelope) AcceptReply(message *Message) {
	select {
	case self.replyC <- message:
	default:
		logrus.
			WithField("seq", message.Sequence()).
			WithField("replyFor", message.ReplyFor()).
			WithField("contentType", message.ContentType).
			Error("could not send reply on reply channel, channel was busy")
	}
}

func (self *replyEnvelope) NotifyErr(err error) {
	self.errC <- err
}

func (self *replyEnvelope) WaitForReply(ch Channel) (*Message, error) {
	if err := self.context.Err(); err != nil {
		return nil, err
	}

	defer self.cancelF()

	self.errC = make(chan error, 1)
	self.replyC = make(chan *Message, 1)

	if err := ch.Send(self); err != nil {
		return nil, err
	}

	select {
	case err := <-self.errC:
		return nil, err
	case <-self.context.Done():
		if err := self.context.Err(); err != nil {
			return nil, TimeoutError{errors.Wrap(err, "timeout waiting for message reply")}
		}
		return nil, errors.New("timeout waiting for message reply")
	case reply := <-self.replyC:
		return reply, nil
	}
}

func NewErrorEnvelope(err error) Envelope {
	return &errorEnvelope{
		ctx: NewErrorContext(err),
	}
}

type errorEnvelope struct {
	ctx context.Context
}

func (self *errorEnvelope) SetSequence(seq int32) {}

func (self *errorEnvelope) Sequence() int32 {
	return 0
}

func (self *errorEnvelope) Msg() *Message {
	return nil
}

func (self *errorEnvelope) ReplyTo(msg *Message) Envelope {
	return self
}

func (self *errorEnvelope) Priority() Priority {
	return Standard
}

func (self *errorEnvelope) Context() context.Context {
	return self.ctx
}

func (self *errorEnvelope) SendListener() SendListener {
	return BaseSendListener{}
}

func (self *errorEnvelope) ReplyReceiver() ReplyReceiver {
	return nil
}

func (self *errorEnvelope) ToSendable() Sendable {
	return self
}

func (self *errorEnvelope) SendAndWaitForWire(Channel) error {
	return self.ctx.Err()
}

func (self *errorEnvelope) SendForReply(Channel) (*Message, error) {
	return nil, self.ctx.Err()
}

func (self *errorEnvelope) WithTimeout(time.Duration) TimeoutEnvelope {
	return self
}

func (self *errorEnvelope) Send(Channel) error {
	return self.ctx.Err()
}

func (self *errorEnvelope) WithPriority(Priority) Envelope {
	return self
}

func NewErrorContext(err error) context.Context {
	result := &errorContext{
		err:     err,
		closedC: make(chan struct{}),
	}
	close(result.closedC)
	return result
}

type errorContext struct {
	err     error
	closedC chan struct{}
}

func (self *errorContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (self *errorContext) Done() <-chan struct{} {
	return self.closedC
}

func (self *errorContext) Err() error {
	return self.err
}

func (self *errorContext) Value(interface{}) interface{} {
	// ignore for now. may need an implementation at some point
	pfxlog.Logger().Error("errorContext.Value called, but not implemented!!!")
	return nil
}
