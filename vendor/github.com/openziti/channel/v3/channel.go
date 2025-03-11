/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package channel

import (
	"context"
	"crypto/x509"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
	"io"
	"net"
	"time"
)

// Channel represents an asynchronous, message-passing framework, designed to sit on top of an underlay.
type Channel interface {
	Identity
	SetLogicalName(logicalName string)
	Sender
	io.Closer
	IsClosed() bool
	Underlay() Underlay
	StartRx()
	GetTimeSinceLastRead() time.Duration
}

type Sender interface {
	// Send will send the given Sendable. If the Sender is busy, it will wait until either the Sender
	// can process the Sendable, the channel is closed or the associated context.Context times out
	Send(s Sendable) error

	// TrySend will send the given Sendable. If the Sender is busy (outgoing message queue is full), it will return
	// immediately rather than wait. The boolean return indicates whether the message was queued or not
	TrySend(s Sendable) (bool, error)
}

// Sendable encapsulates all the data and callbacks that a Channel requires when sending a Message.
type Sendable interface {
	// Msg return the Message to send
	Msg() *Message

	// SetSequence sets a sequence number indicating in which order the message was received
	SetSequence(seq int32)

	// Sequence returns the sequence number
	Sequence() int32

	// Priority returns the Priority of the Message
	Priority() Priority

	// Context returns the Context used for timeouts/cancelling message sends, etc
	Context() context.Context

	// SendListener returns the SendListener to invoke at each stage of the send operation
	SendListener() SendListener

	// ReplyReceiver returns the ReplyReceiver to be invoked when a reply for the message or received, or nil if
	// no ReplyReceiver should be invoked if or when a reply is received
	ReplyReceiver() ReplyReceiver
}

// Envelope allows setting message priority and context. Message is an Envelope (as well as a Sendable)
type Envelope interface {
	// WithPriority returns an Envelope with the given priority
	WithPriority(p Priority) Envelope

	// WithTimeout returns a TimeoutEnvelope with a context using the given timeout
	WithTimeout(duration time.Duration) TimeoutEnvelope

	// Send sends the envelope on the given Channel
	Send(ch Channel) error

	// ReplyTo allows setting the reply header in a fluent style
	ReplyTo(msg *Message) Envelope

	// ToSendable converts the Envelope into a Sendable, which can be submitted to a Channel for sending
	ToSendable() Sendable
}

// TimeoutEnvelope has timeout related convenience methods, such as waiting for a Message to be written to
// the wire or waiting for a Message reply
type TimeoutEnvelope interface {
	Envelope

	// SendAndWaitForWire will wait until the configured timeout or until the message is sent, whichever comes first
	// If the timeout happens first, the context error will be returned, wrapped by a TimeoutError
	SendAndWaitForWire(ch Channel) error

	// SendForReply will wait until the configured timeout or until a reply is received, whichever comes first
	// If the timeout happens first, the context error will be returned, wrapped by a TimeoutError
	SendForReply(ch Channel) (*Message, error)
}

// SendListener is notified at the various stages of a message send
type SendListener interface {
	// Notify Queued is called when the message has been queued for send
	NotifyQueued()
	// NotifyBeforeWrite is called before send is called
	NotifyBeforeWrite()
	// NotifyAfterWrite is called after the message has been written to the Underlay
	NotifyAfterWrite()
	// NotifyErr is called if the Sendable context errors before send or if writing to the Underlay fails
	NotifyErr(error)
}

// ReplyReceiver is used to get notified when a Message reply arrives
type ReplyReceiver interface {
	AcceptReply(*Message)
}

type Identity interface {
	// The Id used to represent the identity of this channel to lower-level resources.
	//
	Id() string

	// The LogicalName represents the purpose or usage of this channel (i.e. 'ctrl', 'mgmt' 'r/001', etc.) Usually used
	// by humans in understand the logical purpose of a channel.
	//
	LogicalName() string

	// The ConnectionId represents the identity of this Channel to internal API components ("instance identifier").
	// Usually used by the Channel framework to differentiate Channel instances.
	//
	ConnectionId() string

	// Certificates contains the identity certificates provided by the peer.
	//
	Certificates() []*x509.Certificate

	// Label constructs a consistently-formatted string used for context logging purposes, from the components above.
	//
	Label() string
}

// UnderlayListener represents a component designed to listen for incoming peer connections.
type UnderlayListener interface {
	Listen(handlers ...ConnectionHandler) error
	UnderlayFactory
	io.Closer
}

// UnderlayFactory is used by Channel to obtain an Underlay instance. An underlay "dialer" or "listener" implement
// UnderlayFactory, to provide instances to Channel.
type UnderlayFactory interface {
	Create(timeout time.Duration) (Underlay, error)
}

// Underlay abstracts a physical communications channel, typically sitting on top of 'transport'.
type Underlay interface {
	Rx() (*Message, error)
	Tx(m *Message) error
	Identity
	io.Closer
	IsClosed() bool
	Headers() map[int32][]byte
	SetWriteTimeout(duration time.Duration) error
	SetWriteDeadline(time time.Time) error
	GetLocalAddr() net.Addr
	GetRemoteAddr() net.Addr
}

type classicUnderlay interface {
	Underlay
	getPeer() transport.Conn
	init(id string, connectionId string, headers Headers)
	rxHello() (*Message, error)
}

const AnyContentType = -1
const HelloSequence = -1

// TimeoutError is used to indicate a timeout happened
type TimeoutError struct {
	error
}

func (self TimeoutError) Unwrap() error {
	return self.error
}

func IsTimeout(err error) bool {
	return errors.As(err, &TimeoutError{})
}

type ClosedError struct{}

func (ClosedError) Error() string {
	return "channel closed"
}

var ListenerClosedError = listenerClosedError{}

type listenerClosedError struct{}

func (err listenerClosedError) Error() string {
	return "closed"
}

// BaseSendable is a type that may be used to provide default methods for Sendable implementation
type BaseSendable struct{}

func (BaseSendable) Msg() *Message {
	return nil
}

func (BaseSendable) Priority() Priority {
	return Standard
}

func (BaseSendable) Context() context.Context {
	return context.Background()
}

func (BaseSendable) SendListener() SendListener {
	return &BaseSendListener{}
}

func (BaseSendable) ReplyReceiver() ReplyReceiver {
	return nil
}

// BaseSendListener is a type that may be used to provide default methods for SendListener implementation
type BaseSendListener struct{}

func (BaseSendListener) NotifyQueued() {}

func (BaseSendListener) NotifyBeforeWrite() {}

func (BaseSendListener) NotifyAfterWrite() {}

func (BaseSendListener) NotifyErr(error) {}
