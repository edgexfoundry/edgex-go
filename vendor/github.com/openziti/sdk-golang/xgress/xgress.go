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

package xgress

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v4"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/foundation/v2/debugz"
	"github.com/openziti/foundation/v2/info"
	"github.com/sirupsen/logrus"
)

const (
	HeaderKeyUUID = 0

	closedFlag            = 0
	rxerStartedFlag       = 1
	endOfCircuitRecvdFlag = 2
	endOfCircuitSentFlag  = 3
	closedTxer            = 4
	rxPushModeFlag        = 5 // false == pull, use rx(), 1 == push, use WriteAdapter

	// bits 8-16 reserved for peer capabilities (same values used in wire format)
	CapabilityEOFIndex = 8
	CapabilityEOFMask  = 1 << CapabilityEOFIndex
	capabilitiesMask   = 0x1FF00 // bits 8-16
)

var ErrWriteClosed = errors.New("write closed")
var ErrPeerClosed = errors.New("peer closed")

type Address string

type AckSender interface {
	SendAck(ack *Acknowledgement, address Address)
}

type OptionsData map[interface{}]interface{}

// BindHandler is an interface invoked to install the appropriate handlers.
type BindHandler interface {
	HandleXgressBind(x *Xgress)
}

type ControlReceiver interface {
	HandleControlReceive(controlType ControlType, headers channel.Headers)
}

type Env interface {
	GetRetransmitter() *Retransmitter
	GetPayloadIngester() *PayloadIngester
	GetMetrics() Metrics
}

// DataPlaneAdapter is invoked by an xgress whenever messages need to be sent to the data plane. Generally a DataPlaneAdapter
// is implemented to connect the xgress to a data plane data transmission system.
type DataPlaneAdapter interface {
	// ForwardPayload is used to forward data payloads onto the data-plane from an xgress
	ForwardPayload(payload *Payload, x *Xgress, ctx context.Context)

	// RetransmitPayload is used to retransmit data payloads onto the data-plane from an xgress
	RetransmitPayload(srcAddr Address, payload *Payload) error

	// ForwardControlMessage is used to forward control messages onto the data-plane from an xgress
	ForwardControlMessage(control *Control, x *Xgress)

	// ForwardAcknowledgement is used to forward acks onto the data-plane from an xgress
	ForwardAcknowledgement(ack *Acknowledgement, address Address)

	Env
}

// CloseHandler is invoked by an xgress when the connected peer terminates the communication.
type CloseHandler interface {
	// HandleXgressClose is invoked when the connected peer terminates the communication.
	//
	HandleXgressClose(x *Xgress)
}

// CloseHandlerF is the function version of CloseHandler
type CloseHandlerF func(x *Xgress)

func (self CloseHandlerF) HandleXgressClose(x *Xgress) {
	self(x)
}

// PeekHandler allows registering watcher to react to data flowing an xgress instance
type PeekHandler interface {
	Rx(x *Xgress, payload *Payload)
	Tx(x *Xgress, payload *Payload)
	Close(x *Xgress)
}

type Connection interface {
	io.Closer
	LogContext() string
	ReadPayload() ([]byte, map[uint8][]byte, error)
	WritePayload([]byte, map[uint8][]byte) (int, error)

	HandleControlMsg(controlType ControlType, headers channel.Headers, responder ControlReceiver) error
}

type SignalConnection interface {
	Connection
	FlowFromFabricToXgressClosed()
}

type Xgress struct {
	timeOfLastRxFromLink int64 // must be first for 64-bit atomic operations on 32-bit machines
	dataPlane            DataPlaneAdapter
	circuitId            string
	ctrlId               string
	address              Address
	peer                 Connection
	originator           Originator
	Options              *Options
	closeNotify          chan struct{}
	rxSequence           uint64
	rxSequenceLock       sync.Mutex
	payloadBuffer        *LinkSendBuffer
	linkRxBuffer         *LinkReceiveBuffer
	closeHandlers        []CloseHandler
	peekHandlers         []PeekHandler
	flags                concurrenz.AtomicBitSet
	tags                 map[string]string
	lastBufferSizeSent   uint32
}

func (self *Xgress) GetDestinationType() string {
	return "xgress"
}

func (self *Xgress) GetIntervalId() string {
	return self.circuitId
}

func (self *Xgress) GetTags() map[string]string {
	return self.tags
}

func NewXgress(circuitId string, ctrlId string, address Address, peer Connection, originator Originator, options *Options, tags map[string]string) *Xgress {
	result := &Xgress{
		circuitId:            circuitId,
		ctrlId:               ctrlId,
		address:              address,
		peer:                 peer,
		originator:           originator,
		Options:              options,
		closeNotify:          make(chan struct{}),
		rxSequence:           0,
		linkRxBuffer:         NewLinkReceiveBuffer(options.TxQueueSize),
		timeOfLastRxFromLink: time.Now().UnixMilli(),
		tags:                 tags,
	}
	result.payloadBuffer = NewLinkSendBuffer(result)
	return result
}

func (self *Xgress) GetTimeOfLastRxFromLink() int64 {
	return atomic.LoadInt64(&self.timeOfLastRxFromLink)
}

func (self *Xgress) CircuitId() string {
	return self.circuitId
}

func (self *Xgress) CtrlId() string {
	return self.ctrlId
}

func (self *Xgress) Address() Address {
	return self.address
}

func (self *Xgress) Originator() Originator {
	return self.originator
}

func (self *Xgress) IsTerminator() bool {
	return self.originator == Terminator
}

func (self *Xgress) SetDataPlaneAdapter(dataPlaneAdapter DataPlaneAdapter) {
	self.dataPlane = dataPlaneAdapter
}

func (self *Xgress) AddCloseHandler(closeHandler CloseHandler) {
	self.closeHandlers = append(self.closeHandlers, closeHandler)
}

func (self *Xgress) AddPeekHandler(peekHandler PeekHandler) {
	self.peekHandlers = append(self.peekHandlers, peekHandler)
}

func (self *Xgress) IsEndOfCircuitReceived() bool {
	return self.flags.IsSet(endOfCircuitRecvdFlag)
}

func (self *Xgress) markCircuitEndReceived() {
	self.flags.Set(endOfCircuitRecvdFlag, true)
}

func (self *Xgress) capabilities() uint32 {
	return CapabilityEOFMask
}

func (self *Xgress) capabilitiesHeader() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, self.capabilities())
	return buf
}

func (self *Xgress) setPeerCapabilities(caps []byte) {
	if len(caps) >= 4 {
		v := binary.BigEndian.Uint32(caps) & capabilitiesMask
		for {
			current := self.flags.Load()
			next := current | v
			if self.flags.CompareAndSetAll(current, next) {
				return
			}
		}
	}
}

func (self *Xgress) peerSupportsEOF() bool {
	return self.flags.IsSet(CapabilityEOFIndex)
}

func (self *Xgress) IsCircuitStarted() bool {
	return !self.IsTerminator() || self.flags.IsSet(rxerStartedFlag)
}

func (self *Xgress) isRxStarted() bool {
	return self.flags.CompareAndSet(rxerStartedFlag, false, true)
}

func (self *Xgress) NewWriteAdapter() *WriteAdapter {
	self.flags.Set(rxPushModeFlag, true)
	return NewWriteAdapter(self)
}

func (self *Xgress) Start() {
	log := pfxlog.ContextLogger(self.Label())
	if self.IsTerminator() {
		log.Debug("terminator: waiting for circuit start before starting receiver")
		if self.Options.CircuitStartTimeout > time.Second {
			time.AfterFunc(self.Options.CircuitStartTimeout, self.terminateIfNotStarted)
		}
	} else if self.isRxStarted() {
		log.Debug("initiator: sending circuit start")
		go self.payloadBuffer.run()
		_ = self.forwardPayload(self.GetStartCircuit(), context.Background())

		if !self.flags.IsSet(rxPushModeFlag) {
			go self.rx()
		}
	}
	go self.tx()
}

func (self *Xgress) terminateIfNotStarted() {
	if !self.IsCircuitStarted() {
		logrus.WithField("xgress", self.Label()).Warn("xgress circuit not started in time, closing")
		self.Close()
	}
}

func (self *Xgress) Label() string {
	return fmt.Sprintf("{c/%s|@/%s}<%s>", self.circuitId, string(self.address), self.originator.String())
}

func (self *Xgress) GetStartCircuit() *Payload {
	startCircuit := &Payload{
		CircuitId: self.circuitId,
		Flags:     SetOriginatorFlag(uint32(PayloadFlagCircuitStart), self.originator),
		Sequence:  int32(self.nextReceiveSequence()),
		Data:      nil,
		Headers: map[uint8][]byte{
			HeaderKeyCapabilities: self.capabilitiesHeader(),
		},
	}
	return startCircuit
}

func (self *Xgress) GetEndCircuit() *Payload {
	endCircuit := &Payload{
		CircuitId: self.circuitId,
		Flags:     SetOriginatorFlag(uint32(PayloadFlagCircuitEnd), self.originator),
		Sequence:  int32(self.nextReceiveSequence()),
		Data:      nil,
	}
	return endCircuit
}

func (self *Xgress) ForwardEndOfCircuit(sendF func(payload *Payload) bool) {
	// for now always send end of circuit. too many is better than not enough
	if self.flags.CompareAndSet(endOfCircuitSentFlag, false, true) {
		sendF(self.GetEndCircuit())
	}
}

func (self *Xgress) IsEndOfCircuitSent() bool {
	return self.flags.IsSet(endOfCircuitSentFlag)
}

func (self *Xgress) CloseRxTimeout() {
	self.sendEOF()
	self.payloadBuffer.CloseWhenEmpty()
}

func (self *Xgress) Unrouted() {
	// if we're unrouted no more data is inbound. this still allows already queued data to flow to the client
	self.CloseXgToClient()

	// If we're unrouted, we can't send any more data, so close the payload buffer
	self.payloadBuffer.Close()
}

func (self *Xgress) CloseXgToClient() {
	pfxlog.ContextLogger(self.Label()).Debug("close xg to client")
	if self.flags.CompareAndSet(closedTxer, false, true) {
		close(self.closeNotify)
	}

	if self.payloadBuffer.IsClosed() {
		self.Close()
	}
}

/*
Close should only be called once both sides of the circuit are complete.
*/
func (self *Xgress) Close() {
	log := pfxlog.ContextLogger(self.Label())

	if self.flags.CompareAndSet(closedFlag, false, true) {
		log.Debug("closing xgress")

		self.sendEndOfCircuit()

		if err := self.peer.Close(); err != nil {
			log.WithError(err).Warn("error while closing xgress peer")
		}

		log.Debug("closing tx queue")
		if self.flags.CompareAndSet(closedTxer, false, true) {
			close(self.closeNotify)
		}

		self.payloadBuffer.Close()

		for _, peekHandler := range self.peekHandlers {
			peekHandler.Close(self)
		}

		if len(self.closeHandlers) != 0 {
			for _, closeHandler := range self.closeHandlers {
				closeHandler.HandleXgressClose(self)
			}
		} else {
			pfxlog.ContextLogger(self.Label()).Warn("no close handler")
		}
	}
}

func (self *Xgress) PeerClosed() {
	log := pfxlog.ContextLogger(self.Label())
	log.Debug("peer closed")
	self.CloseXgToClient()
	self.CloseRxTimeout()
}

func (self *Xgress) closeIfRxAndTxDone() {
	if self.payloadBuffer.IsClosed() && self.flags.IsSet(closedTxer) {
		self.Close()
	}
}

func (self *Xgress) CloseSendBuffer() {
	self.payloadBuffer.Close()
}

func (self *Xgress) Closed() bool {
	return self.flags.IsSet(closedFlag)
}

func (self *Xgress) IsClosed() bool {
	return self.flags.IsSet(closedFlag)
}

func (self *Xgress) SendPayload(payload *Payload, _ time.Duration, _ PayloadType) error {
	if self.Closed() {
		return nil
	}

	if payload.IsCircuitEndFlagSet() {
		pfxlog.ContextLogger(self.Label()).Debug("received end of circuit payload")
	}

	if payload.IsFlagWriteFailedSet() {
		pfxlog.ContextLogger(self.Label()).Debug("received write failed payload")
		self.payloadBuffer.Close()
	}

	atomic.StoreInt64(&self.timeOfLastRxFromLink, time.Now().UnixMilli())
	self.dataPlane.GetPayloadIngester().ingest(payload, self)

	return nil
}

func (self *Xgress) SendAcknowledgement(acknowledgement *Acknowledgement) error {
	self.dataPlane.GetMetrics().MarkAckReceived()
	self.payloadBuffer.ReceiveAcknowledgement(acknowledgement)
	return nil
}

func (self *Xgress) SendControl(control *Control) error {
	return self.peer.HandleControlMsg(control.Type, control.Headers, self)
}

func (self *Xgress) HandleControlReceive(controlType ControlType, headers channel.Headers) {
	control := &Control{
		Type:      controlType,
		CircuitId: self.circuitId,
		Headers:   headers,
	}
	self.dataPlane.ForwardControlMessage(control, self)
}

func (self *Xgress) acceptPayload(payload *Payload) {
	if payload.IsCircuitStartFlagSet() && self.isRxStarted() {
		pfxlog.ContextLogger(self.Label()).Debug("start received")

		var peerSentCapabilities bool
		if caps, ok := payload.Headers[HeaderKeyCapabilities]; ok {
			self.setPeerCapabilities(caps)
			peerSentCapabilities = true
		}

		go self.payloadBuffer.run()

		if peerSentCapabilities {
			self.sendCapabilitiesResponse()
		}

		if !self.flags.IsSet(rxPushModeFlag) {
			go self.rx()
		}
	}

	if !self.Options.RandomDrops || rand.Int31n(self.Options.Drop1InN) != 1 {
		self.PayloadReceived(payload)
	}
}

func (self *Xgress) tx() {
	log := pfxlog.ContextLogger(self.Label())

	log.Debug("started")
	defer log.Debug("exited")
	defer func() {
		if signalConn, ok := self.peer.(SignalConnection); ok {
			signalConn.FlowFromFabricToXgressClosed()
		}
		if !self.IsEndOfCircuitReceived() {
			self.sendWriteFailed()
		}
	}()

	clearPayloadFromSendBuffer := func(payload *Payload) {
		payloadSize := len(payload.Data)
		size := atomic.AddUint32(&self.linkRxBuffer.size, ^uint32(payloadSize-1)) // subtraction for uint32

		payloadLogger := log.WithFields(payload.GetLoggerFields())
		payloadLogger.Debugf("payload %v of size %v removed from rx buffer, new size: %v", payload.Sequence, payloadSize, size)

		lastBufferSizeSent := self.getLastBufferSizeSent()
		if lastBufferSizeSent > 10000 && (lastBufferSizeSent>>1) > size {
			self.SendEmptyAck()
		}
	}

	sendPayload := func(payload *Payload) bool {
		payloadLogger := log.WithFields(payload.GetLoggerFields())

		if payload.IsCircuitEndFlagSet() || payload.IsFlagEOFSet() {
			self.markCircuitEndReceived()
			self.CloseXgToClient()
			payloadLogger.Debug("circuit end payload received, exiting")
			return false
		}

		// Intercept capabilities header from peer
		if caps, ok := payload.Headers[HeaderKeyCapabilities]; ok {
			self.setPeerCapabilities(caps)
			delete(payload.Headers, HeaderKeyCapabilities)
			payloadLogger.Debug("peer capabilities received")
			if len(payload.Data) == 0 && !payload.IsCircuitStartFlagSet() {
				return true // capabilities-only payload, consume it
			}
		}

		payloadLogger.Debug("sending")

		for _, peekHandler := range self.peekHandlers {
			peekHandler.Tx(self, payload)
		}

		if !payload.IsCircuitStartFlagSet() {
			start := time.Now()
			n, err := self.peer.WritePayload(payload.Data, payload.Headers)
			if err != nil {
				payloadLogger.Warnf("write failed (%s), closing xgress", err)
				self.Close()
				return false
			} else {
				self.dataPlane.GetMetrics().PayloadWritten(time.Since(start))
				payloadLogger.Debugf("payload sent [%s]", info.ByteCount(int64(n)))
			}
		}
		return true
	}

	var payload *Payload
	var payloadChunk *Payload

	payloadStarted := false
	payloadComplete := false
	var payloadSize uint64
	var payloadWriteOffset int

	for {
		payloadChunk = self.linkRxBuffer.NextPayload(self.closeNotify)

		if payloadChunk == nil {
			log.Debug("nil payload received, exiting")
			return
		}

		if !isFlagSet(payloadChunk.GetFlags(), PayloadFlagChunk) {
			if !sendPayload(payloadChunk) {
				return
			}
			clearPayloadFromSendBuffer(payloadChunk)
			continue
		}

		var payloadReadOffset int
		if !payloadStarted {
			payloadSize, payloadReadOffset = binary.Uvarint(payloadChunk.Data)

			if len(payloadChunk.Data) == 0 || payloadSize+uint64(payloadReadOffset) == uint64(len(payloadChunk.Data)) {
				payload = payloadChunk
				payload.Data = payload.Data[payloadReadOffset:]
				payloadComplete = true
			} else {
				payload = &Payload{
					CircuitId: payloadChunk.CircuitId,
					Flags:     payloadChunk.Flags,
					RTT:       payloadChunk.RTT,
					Sequence:  payloadChunk.Sequence,
					Headers:   payloadChunk.Headers,
					Data:      make([]byte, payloadSize),
				}
			}
			payloadStarted = true
		}

		if !payloadComplete {
			chunkData := payloadChunk.Data[payloadReadOffset:]
			copy(payload.Data[payloadWriteOffset:], chunkData)
			payloadWriteOffset += len(chunkData)
			payloadComplete = uint64(payloadWriteOffset) == payloadSize
		}

		payloadLogger := log.WithFields(payload.GetLoggerFields())
		payloadLogger.Debugf("received payload chunk. seq: %d, first: %v, complete: %v, chunk size: %d, payload size: %d, writeOffset: %d",
			payloadChunk.Sequence, len(payload.Data) == 0 || payloadReadOffset > 0, payloadComplete, len(payloadChunk.Data), payloadSize, payloadWriteOffset)

		if !payloadComplete {
			clearPayloadFromSendBuffer(payloadChunk)
			continue
		}

		payloadStarted = false
		payloadComplete = false
		payloadWriteOffset = 0

		if !sendPayload(payload) {
			return
		}
		clearPayloadFromSendBuffer(payloadChunk)
	}
}

func (self *Xgress) sendCapabilitiesResponse() {
	payload := &Payload{
		CircuitId: self.circuitId,
		Flags:     SetOriginatorFlag(0, self.originator),
		Sequence:  int32(self.nextReceiveSequence()),
		Data:      nil,
		Headers: map[uint8][]byte{
			HeaderKeyCapabilities: self.capabilitiesHeader(),
		},
	}
	_ = self.forwardPayload(payload, context.Background())
}

func (self *Xgress) sendEOF() {
	log := pfxlog.ContextLogger(self.Label())
	log.Debug("sendEOF")

	if self.payloadBuffer.closed.Load() {
		// Avoid spurious 'failed to forward payload' error if the buffer is already closed
		return
	}

	var flag Flag
	if self.peerSupportsEOF() {
		flag = PayloadFlagEOF
		log.Debug("peer supports EOF, sending EOF flag")
	} else {
		flag = PayloadFlagCircuitEnd
		log.Debug("peer does not support EOF, falling back to CircuitEnd")
	}

	payload := &Payload{
		CircuitId: self.circuitId,
		Flags:     SetOriginatorFlag(uint32(flag), self.originator),
		Sequence:  int32(self.nextReceiveSequence()),
		Data:      nil,
	}

	_ = self.forwardPayload(payload, context.Background())
}

func (self *Xgress) sendWriteFailed() {
	log := pfxlog.ContextLogger(self.Label())
	log.Debug("sendWriteFailed")

	if self.payloadBuffer.closed.Load() {
		return
	}

	payload := &Payload{
		CircuitId: self.circuitId,
		Flags:     SetOriginatorFlag(uint32(PayloadFlagWriteFailed), self.originator),
		Sequence:  int32(self.nextReceiveSequence()),
		Data:      nil,
	}

	log.Debug("sending end of circuit payload")
	_ = self.forwardPayload(payload, context.Background())
}

func (self *Xgress) sendEndOfCircuit() {
	log := pfxlog.ContextLogger(self.Label())
	log.Debug("sendEndOfCircuit")
	self.dataPlane.ForwardPayload(self.GetEndCircuit(), self, context.Background())
}

/**
	Payload format

	Field 1: 1 byte - version and flags
      Masks
      * 00000000 - Always 0 to indicate type. The standard channel header 4 byte protocol indicator has a 1 in bit 0 of the first byte
      * 00000110 - Version, v0-v3. Assumption is that if we ever get to v4, we can roll back to 0, since everything
                   should have upgraded past v0 by that point
      * 00001000 - Terminator Flag - indicates the payload origin, initiator (0) or terminator (1)
      * 00010000 - RTT Flag. Indicates if the payload contains an RTT. We don't need to send RTT on every payload.
      * 00100000 - Chunk Flag. Indicates if this payload is chunked.
      * 01000000 - Headers flag. Indicates this payload contains headers.
      * 10000000 - Heartbeat Flag. Indicates the payload contains a heartbeat

    Field 2: 1 byte, Circuit id size
      Masks
      * 00001111 - Number of bytes in circuit id. Supports circuit ids which take up to 15 bytes.
                   Circuits ids are currently at 9 bytes.
      * 11110000 - currently unused

    Field 3: RTT (optional)
      - 2 bytes

    Field 4: CircuitId
      - direct bytes representation of string encoded circuit id

    Field 5: Sequence number
      - Encoded using binary.PutUvarint

    Field 6: Headers
      - Presence indicated by headers flag in first field
      length - encoded with binary.PutUvarint
      for each key/value pair -
         key - 1 byte
         value length - encoded with binary.PutUvarint
         value - byte array, directly appended


    Field 7: Data

    Field 8: Heartbeat
      - 8 bytes
      - only included if there's extra room
*/

const (
	VersionMask        byte = 0b00000110
	TerminatorFlagMask byte = 0b00001000
	RttFlagMask        byte = 0b00010000
	ChunkFlagMask      byte = 0b00100000
	HeadersFlagMask    byte = 0b01000000
	HeartbeatFlagMask  byte = 0b10000000

	CircuitIdSizeMask     byte = 0b00001111
	PayloadProtocolV1     byte = 1
	PayloadProtocolOffset byte = 1
)

func (self *Xgress) rx() {
	log := pfxlog.ContextLogger(self.Label())

	log.Debugf("started with peer: %v", self.peer.LogContext())
	defer log.Debug("exited")

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("send on closed channel. error: (%v)", r)
			return
		}
	}()

	defer self.CloseRxTimeout()

	for {
		buffer, headers, err := self.peer.ReadPayload()
		log.Debugf("payload read: %d bytes read", len(buffer))
		n := len(buffer)

		// if we got an EOF, but also some data, ignore the EOF, next read we'll get 0, EOF
		if err != nil && (n == 0 || err != io.EOF) {
			if err == io.EOF || errors.Is(err, ErrPeerClosed) {
				if errors.Is(err, ErrPeerClosed) { // if the peer closed, we need to close the txer as well
					self.CloseXgToClient()
				}
				log.Debugf("EOF, exiting xgress.rx loop")
			} else {
				log.Warnf("read failed (%s)", err)
			}

			return
		}

		if self.Closed() {
			return
		}

		if err = self.Write(buffer, headers, nil); err != nil {
			return
		}
	}
}

func (self *Xgress) Write(buffer []byte, headers map[uint8][]byte, ctx context.Context) error {
	log := pfxlog.ContextLogger(self.Label())

	log.Debugf("payload read: %d bytes read", len(buffer))
	n := len(buffer)

	if self.Closed() {
		return ErrWriteClosed
	}

	if self.Options.Mtu == 0 {
		return self.sendUnchunkedBuffer(buffer, headers, ctx)
	}

	first := true
	chunked := false
	var err error

	for len(buffer) > 0 || (first && len(headers) > 0) {
		seq := self.nextReceiveSequence()

		chunk := make([]byte, self.Options.Mtu)

		flagsHeader := VersionMask & (PayloadProtocolV1 << PayloadProtocolOffset)
		var sizesHeader byte
		if self.originator == Terminator {
			flagsHeader |= TerminatorFlagMask
		}

		written := 2
		rest := chunk[2:]
		includeRtt := seq%5 == 0
		if includeRtt {
			flagsHeader |= RttFlagMask
			written += 2
			rest = rest[2:]
		}

		size := copy(rest, self.circuitId)
		sizesHeader |= CircuitIdSizeMask & uint8(size)
		written += size
		rest = rest[size:]
		size = binary.PutUvarint(rest, seq)
		rest = rest[size:]
		written += size

		if first && len(headers) > 0 {
			flagsHeader |= HeadersFlagMask
			size, err = writeU8ToBytesMap(headers, rest)
			if err != nil {
				log.WithError(err).Error("payload encoding error, closing")
				return err
			}
			rest = rest[size:]
			written += size
		}

		data := rest
		dataLen := 0
		if first && len(rest) < len(buffer) {
			chunked = true
			size = binary.PutUvarint(rest, uint64(n))
			dataLen += size
			written += size
			rest = rest[size:]
		}

		if chunked {
			flagsHeader |= ChunkFlagMask
		}

		size = copy(rest, buffer)
		written += size
		dataLen += size

		buffer = buffer[size:]

		// check if there's room for a heartbeat
		if written+8 <= len(chunk) {
			flagsHeader |= HeartbeatFlagMask
			written += 8
		}

		chunk[0] = flagsHeader
		chunk[1] = sizesHeader

		payload := &Payload{
			CircuitId: self.circuitId,
			Flags:     SetOriginatorFlag(0, self.originator),
			Sequence:  int32(seq),
			Data:      data[:dataLen],
			raw:       chunk[:written],
		}

		if chunked {
			payload.Flags = setPayloadFlag(payload.Flags, PayloadFlagChunk)
		}

		if first {
			payload.Headers = headers
		}

		log.Debugf("sending payload chunk. seq: %d, first: %v, chunk size: %d, payload size: %d, remainder: %d", payload.Sequence, first, len(payload.Data), n, len(buffer))
		first = false

		// if the payload buffer is closed, we can't forward any more data, so might as well exit the rx loop
		// The txer will still have a chance to flush any already received data
		if err = self.forwardPayload(payload, ctx); err != nil {
			return err
		}

		payloadLogger := log.WithFields(payload.GetLoggerFields())
		payloadLogger.Debugf("forwarded [%s]", info.ByteCount(int64(n)))
	}

	logrus.Debugf("received payload for [%d] bytes", n)
	return nil
}

func (self *Xgress) sendUnchunkedBuffer(buf []byte, headers map[uint8][]byte, ctx context.Context) error {
	log := pfxlog.ContextLogger(self.Label())

	payload := &Payload{
		CircuitId: self.circuitId,
		Flags:     SetOriginatorFlag(0, self.originator),
		Sequence:  int32(self.nextReceiveSequence()),
		Data:      buf,
		Headers:   headers,
	}

	log.Debugf("sending unchunked payload. seq: %d, payload size: %d", payload.Sequence, len(payload.Data))

	// if the payload buffer is closed, we can't forward any more data, so might as well exit the rx loop
	// The txer will still have a chance to flush any already received data
	if err := self.forwardPayload(payload, ctx); err != nil {
		return err
	}

	payloadLogger := log.WithFields(payload.GetLoggerFields())
	payloadLogger.Debugf("forwarded [%s]", info.ByteCount(int64(len(buf))))
	return nil
}

func (self *Xgress) forwardPayload(payload *Payload, ctx context.Context) error {
	var sendCallback func()
	var err error

	if ctx == nil {
		sendCallback, err = self.payloadBuffer.BufferPayload(payload)
	} else {
		sendCallback, err = self.payloadBuffer.BufferPayloadWithDeadline(payload, ctx)
	}

	if err != nil {
		if !payload.IsCircuitEndFlagSet() && !payload.IsFlagEOFSet() {
			pfxlog.ContextLogger(self.Label()).WithError(err).Error("failure to buffer payload")
		}
		return err
	}

	for _, peekHandler := range self.peekHandlers {
		peekHandler.Rx(self, payload)
	}

	self.dataPlane.ForwardPayload(payload, self, ctx)
	sendCallback()
	return nil
}

func (self *Xgress) nextReceiveSequence() uint64 {
	self.rxSequenceLock.Lock()
	defer self.rxSequenceLock.Unlock()

	next := self.rxSequence
	self.rxSequence++

	return next
}

func (self *Xgress) PayloadReceived(payload *Payload) {
	log := pfxlog.ContextLogger(self.Label()).WithFields(payload.GetLoggerFields())
	log.Debug("payload received")
	if self.originator == payload.GetOriginator() {
		// a payload sent from this xgress has arrived back at this xgress, instead of the other end
		log.Warn("ouroboros (circuit cycle) detected, dropping payload")
	} else if self.linkRxBuffer.ReceiveUnordered(self, payload, self.Options.RxBufferSize) {
		log.Debug("ready to acknowledge")

		ack := NewAcknowledgement(self.circuitId, self.originator)
		ack.RecvBufferSize = self.linkRxBuffer.Size()
		ack.Sequence = append(ack.Sequence, payload.Sequence)
		ack.RTT = payload.RTT

		atomic.StoreUint32(&self.lastBufferSizeSent, ack.RecvBufferSize)
		self.dataPlane.ForwardAcknowledgement(ack, self.address)
	} else {
		log.Debug("dropped")
	}
}

func (self *Xgress) SendEmptyAck() {
	pfxlog.ContextLogger(self.Label()).WithField("circuit", self.circuitId).Debug("sending empty ack")
	ack := NewAcknowledgement(self.circuitId, self.originator)
	ack.RecvBufferSize = self.linkRxBuffer.Size()
	atomic.StoreUint32(&self.lastBufferSizeSent, ack.RecvBufferSize)
	self.dataPlane.ForwardAcknowledgement(ack, self.address)
}

func (self *Xgress) GetSequence() uint64 {
	self.rxSequenceLock.Lock()
	defer self.rxSequenceLock.Unlock()
	return self.rxSequence
}

func (self *Xgress) getLastBufferSizeSent() uint32 {
	return atomic.LoadUint32(&self.lastBufferSizeSent)
}

func (self *Xgress) InspectCircuit(detail *CircuitInspectDetail) {
	detail.AddXgressDetail(self.GetInspectDetail(detail.includeGoroutines))
}

func (self *Xgress) GetInspectDetail(includeGoroutines bool) *InspectDetail {
	timeSinceLastRxFromLink := time.Duration(time.Now().UnixMilli()-atomic.LoadInt64(&self.timeOfLastRxFromLink)) * time.Millisecond
	xgressDetail := &InspectDetail{
		Address:               string(self.address),
		Originator:            self.originator.String(),
		TimeSinceLastLinkRx:   timeSinceLastRxFromLink.String(),
		SendBufferDetail:      self.payloadBuffer.Inspect(),
		RecvBufferDetail:      self.linkRxBuffer.Inspect(),
		XgressPointer:         fmt.Sprintf("%p", self),
		LinkSendBufferPointer: fmt.Sprintf("%p", self.payloadBuffer),
		Sequence:              self.GetSequence(),
		Flags:                 strconv.FormatUint(uint64(self.flags.Load()), 2),
		LastSizeSent:          self.getLastBufferSizeSent(),
	}

	if includeGoroutines {
		xgressDetail.Goroutines = self.getRelatedGoroutines(xgressDetail.XgressPointer, xgressDetail.LinkSendBufferPointer)
	}

	return xgressDetail
}

func (self *Xgress) getRelatedGoroutines(contains ...string) []string {
	reader := bytes.NewBufferString(debugz.GenerateStack())
	scanner := bufio.NewScanner(reader)
	var result []string
	var buf *bytes.Buffer
	xgressRelated := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "goroutine") && strings.HasSuffix(line, ":") {
			result = self.addGoroutineIfRelated(buf, xgressRelated, result, contains...)
			buf = &bytes.Buffer{}
			xgressRelated = false
		}

		if buf != nil {
			if strings.Contains(line, "xgress") {
				xgressRelated = true
			}
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	result = self.addGoroutineIfRelated(buf, xgressRelated, result, contains...)
	if err := scanner.Err(); err != nil {
		result = append(result, "goroutine parsing error: %v", err.Error())
	}
	return result
}

func (self *Xgress) addGoroutineIfRelated(buf *bytes.Buffer, xgressRelated bool, result []string, contains ...string) []string {
	if !xgressRelated {
		return result
	}
	if buf != nil {
		gr := buf.String()
		// ignore the current goroutine
		if strings.Contains(gr, "GenerateStack") {
			return result
		}

		for _, s := range contains {
			if strings.Contains(gr, s) {
				result = append(result, gr)
				break
			}
		}
	}
	return result
}

func UnmarshallPacketPayload(buf []byte) (*channel.Message, error) {
	flagsField := buf[0]
	if flagsField&1 != 0 {
		return channel.ReadV2(bytes.NewBuffer(buf))
	}
	version := (flagsField & VersionMask) >> 1
	if version != PayloadProtocolV1 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}
	sizeField := buf[1]
	circuitIdSize := CircuitIdSizeMask & sizeField
	rest := buf[2:]

	var rtt *uint16
	if flagsField&RttFlagMask != 0 {
		b0 := rest[0]
		b1 := rest[1]
		rest = rest[2:]
		val := uint16(b0) | (uint16(b1) << 8)
		rtt = &val
	}

	var heartbeat *uint64
	if flagsField&HeartbeatFlagMask != 0 {
		val := binary.BigEndian.Uint64(rest[len(rest)-8:])
		heartbeat = &val
		rest = rest[:len(rest)-8]
	}

	circuitId := string(rest[:circuitIdSize])
	rest = rest[circuitIdSize:]
	seq, read := binary.Uvarint(rest)
	rest = rest[read:]

	var headers map[uint8][]byte
	if flagsField&HeadersFlagMask != 0 {
		var err error
		headers, rest, err = readU8ToBytesMap(rest)
		if err != nil {
			return nil, err
		}
	}

	msg := channel.NewMessage(ContentTypePayloadType, rest)
	addPayloadHeadersToMsg(msg, headers)
	msg.PutStringHeader(HeaderKeyCircuitId, circuitId)
	msg.PutUint64Header(HeaderKeySequence, seq)
	if heartbeat != nil {
		msg.PutUint64Header(channel.HeartbeatHeader, *heartbeat)
	}
	msg.Headers[HeaderPayloadRaw] = buf

	flags := uint32(0)

	if flagsField&ChunkFlagMask != 0 {
		flags = setPayloadFlag(flags, PayloadFlagChunk)
	}

	if flagsField&TerminatorFlagMask != 0 {
		flags = setPayloadFlag(flags, PayloadFlagOriginator)
	}

	if flags != 0 {
		msg.PutUint32Header(HeaderKeyFlags, flags)
	}

	if rtt != nil {
		msg.PutUint16Header(HeaderKeyRTT, *rtt)
	}

	return msg, nil
}

func writeU8ToBytesMap(m map[uint8][]byte, buf []byte) (int, error) {
	written := binary.PutUvarint(buf, uint64(len(m)))
	buf = buf[written:]
	for k, v := range m {
		if len(buf) < 10 {
			return 0, fmt.Errorf("header too large, no space for header keys, payload has only %d bytes left", len(buf))
		}
		buf[0] = k
		written++
		buf = buf[1:]

		fieldLen := binary.PutUvarint(buf, uint64(len(v)))
		buf = buf[fieldLen:]
		written += fieldLen
		if len(buf) < len(v) {
			return 0, fmt.Errorf("header too large, no space for header value of size %d, only %d bytes available", len(v), len(buf))
		}

		fieldLen = copy(buf, v)
		buf = buf[fieldLen:]
		written += fieldLen
	}

	return written, nil
}

func readU8ToBytesMap(buf []byte) (map[uint8][]byte, []byte, error) {
	result := map[uint8][]byte{}
	count, offset := binary.Uvarint(buf)
	if offset < 1 {
		return nil, nil, errors.New("error reading payload header map length")
	}
	buf = buf[offset:]
	for i := range count {
		if len(buf) < 2 {
			return nil, nil, fmt.Errorf("payload header error, ran out of space reading header %d", i)
		}
		k := buf[0]
		valSize, read := binary.Uvarint(buf[1:])
		if read < 1 {
			return nil, nil, fmt.Errorf("payload header error, ran out of space reading header %d", i)
		}
		buf = buf[read+1:]
		if len(buf) < int(valSize) {
			return nil, nil, fmt.Errorf("payload header error, ran out of space reading header %d", i)
		}
		result[k] = buf[:valSize]
		buf = buf[valSize:]
	}

	return result, buf, nil
}

func NewWriteAdapter(x *Xgress) *WriteAdapter {
	result := &WriteAdapter{
		x: x,
	}
	result.doneNotify.Store(make(chan struct{}))
	return result
}

type WriteAdapter struct {
	x                *Xgress
	deadline         concurrenz.AtomicValue[time.Time]
	doneNotify       concurrenz.AtomicValue[chan struct{}]
	doneNotifyClosed bool
	lock             sync.Mutex
}

func (self *WriteAdapter) Deadline() (deadline time.Time, ok bool) {
	deadline = self.deadline.Load()
	return deadline, !deadline.IsZero()
}

func (self *WriteAdapter) Done() <-chan struct{} {
	return self.doneNotify.Load()
}

func (self *WriteAdapter) Err() error {
	return nil
}

func (self *WriteAdapter) Value(any) any {
	return nil
}

func (self *WriteAdapter) SetWriteDeadline(t time.Time) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.deadline.Store(t)
	if t.IsZero() {
		if self.doneNotifyClosed {
			self.doneNotify.Store(make(chan struct{}))
			self.doneNotifyClosed = false
		}
		return nil
	}
	d := time.Until(t)
	if d > 0 {
		if self.doneNotifyClosed {
			self.doneNotify.Store(make(chan struct{}))
			self.doneNotifyClosed = false
		}

		time.AfterFunc(d, func() {
			self.lock.Lock()
			defer self.lock.Unlock()

			if t.Equal(self.deadline.Load()) {
				if !self.doneNotifyClosed {
					close(self.doneNotify.Load())
					self.doneNotifyClosed = true
				}
			}
		})
	} else {
		if !self.doneNotifyClosed {
			close(self.doneNotify.Load())
			self.doneNotifyClosed = true
		}
	}

	return nil
}

func (self *WriteAdapter) Write(b []byte) (n int, err error) {
	if err = self.x.Write(b, nil, self); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (self *WriteAdapter) WriteToXgress(b []byte, header map[uint8][]byte) (n int, err error) {
	if err = self.x.Write(b, header, self); err != nil {
		return 0, err
	}
	return len(b), nil
}
