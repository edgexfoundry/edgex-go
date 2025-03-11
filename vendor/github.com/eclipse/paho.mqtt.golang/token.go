/*
 * Copyright (c) 2021 IBM Corp and others.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.
 *
 * Contributors:
 *    Allan Stockdill-Mander
 */

package mqtt

import (
	"errors"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// PacketAndToken is a struct that contains both a ControlPacket and a
// Token. This struct is passed via channels between the client interface
// code and the underlying code responsible for sending and receiving
// MQTT messages.
type PacketAndToken struct {
	p packets.ControlPacket
	t tokenCompletor
}

// Token defines the interface for the tokens used to indicate when
// actions have completed.
type Token interface {
	// Wait will wait indefinitely for the Token to complete, ie the Publish
	// to be sent and confirmed receipt from the broker.
	Wait() bool

	// WaitTimeout takes a time.Duration to wait for the flow associated with the
	// Token to complete, returns true if it returned before the timeout or
	// returns false if the timeout occurred. In the case of a timeout the Token
	// does not have an error set in case the caller wishes to wait again.
	WaitTimeout(time.Duration) bool

	// Done returns a channel that is closed when the flow associated
	// with the Token completes. Clients should call Error after the
	// channel is closed to check if the flow completed successfully.
	//
	// Done is provided for use in select statements. Simple use cases may
	// use Wait or WaitTimeout.
	Done() <-chan struct{}

	Error() error
}

type TokenErrorSetter interface {
	setError(error)
}

type tokenCompletor interface {
	Token
	TokenErrorSetter
	flowComplete()
}

type baseToken struct {
	m        sync.RWMutex
	complete chan struct{}
	err      error
}

// Wait implements the Token Wait method.
func (b *baseToken) Wait() bool {
	<-b.complete
	return true
}

// WaitTimeout implements the Token WaitTimeout method.
func (b *baseToken) WaitTimeout(d time.Duration) bool {
	timer := time.NewTimer(d)
	select {
	case <-b.complete:
		if !timer.Stop() {
			<-timer.C
		}
		return true
	case <-timer.C:
	}

	return false
}

// Done implements the Token Done method.
func (b *baseToken) Done() <-chan struct{} {
	return b.complete
}

func (b *baseToken) flowComplete() {
	select {
	case <-b.complete:
	default:
		close(b.complete)
	}
}

func (b *baseToken) Error() error {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.err
}

func (b *baseToken) setError(e error) {
	b.m.Lock()
	b.err = e
	b.flowComplete()
	b.m.Unlock()
}

func newToken(tType byte) tokenCompletor {
	switch tType {
	case packets.Connect:
		return &ConnectToken{baseToken: baseToken{complete: make(chan struct{})}}
	case packets.Subscribe:
		return &SubscribeToken{baseToken: baseToken{complete: make(chan struct{})}, subResult: make(map[string]byte)}
	case packets.Publish:
		return &PublishToken{baseToken: baseToken{complete: make(chan struct{})}}
	case packets.Unsubscribe:
		return &UnsubscribeToken{baseToken: baseToken{complete: make(chan struct{})}}
	case packets.Disconnect:
		return &DisconnectToken{baseToken: baseToken{complete: make(chan struct{})}}
	}
	return nil
}

// ConnectToken is an extension of Token containing the extra fields
// required to provide information about calls to Connect()
type ConnectToken struct {
	baseToken
	returnCode     byte
	sessionPresent bool
}

// ReturnCode returns the acknowledgement code in the connack sent
// in response to a Connect()
func (c *ConnectToken) ReturnCode() byte {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.returnCode
}

// SessionPresent returns a bool representing the value of the
// session present field in the connack sent in response to a Connect()
func (c *ConnectToken) SessionPresent() bool {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.sessionPresent
}

// PublishToken is an extension of Token containing the extra fields
// required to provide information about calls to Publish()
type PublishToken struct {
	baseToken
	messageID uint16
}

// MessageID returns the MQTT message ID that was assigned to the
// Publish packet when it was sent to the broker
func (p *PublishToken) MessageID() uint16 {
	return p.messageID
}

// SubscribeToken is an extension of Token containing the extra fields
// required to provide information about calls to Subscribe()
type SubscribeToken struct {
	baseToken
	subs      []string
	subResult map[string]byte
	messageID uint16
}

// Result returns a map of topics that were subscribed to along with
// the matching return code from the broker. This is either the Qos
// value of the subscription or an error code.
func (s *SubscribeToken) Result() map[string]byte {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.subResult
}

// UnsubscribeToken is an extension of Token containing the extra fields
// required to provide information about calls to Unsubscribe()
type UnsubscribeToken struct {
	baseToken
	messageID uint16
}

// DisconnectToken is an extension of Token containing the extra fields
// required to provide information about calls to Disconnect()
type DisconnectToken struct {
	baseToken
}

// TimedOut is the error returned by WaitTimeout when the timeout expires
var TimedOut = errors.New("context canceled")

// WaitTokenTimeout is a utility function used to simplify the use of token.WaitTimeout
// token.WaitTimeout may return `false` due to time out but t.Error() still results
// in nil.
// `if t := client.X(); t.WaitTimeout(time.Second) && t.Error() != nil {` may evaluate
// to false even if the operation fails.
// It is important to note that if TimedOut is returned, then the operation may still be running
// and could eventually complete successfully.
func WaitTokenTimeout(t Token, d time.Duration) error {
	if !t.WaitTimeout(d) {
		return TimedOut
	}
	return t.Error()
}
