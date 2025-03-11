/*
	Copyright 2019 NetFoundry Inc.

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

package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/foundation/v2/info"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/secretstream"
	"github.com/openziti/secretstream/kx"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var unsupportedCrypto = errors.New("unsupported crypto")

type ConnType byte

const (
	ConnTypeDial ConnType = 1
	ConnTypeBind ConnType = 2
)

var _ edge.Conn = &edgeConn{}

type edgeConn struct {
	edge.MsgChannel
	readQ                 *noopSeq[*channel.Message]
	inBuffer              [][]byte
	msgMux                edge.MsgMux
	hosting               cmap.ConcurrentMap[string, *edgeListener]
	flags                 uint32
	closed                atomic.Bool
	readFIN               atomic.Bool
	sentFIN               atomic.Bool
	serviceName           string
	sourceIdentity        string
	acceptCompleteHandler *newConnHandler
	connType              ConnType
	marker                string
	circuitId             string
	customState           map[int32][]byte

	crypto   bool
	keyPair  *kx.KeyPair
	rxKey    []byte
	receiver secretstream.Decryptor
	sender   secretstream.Encryptor
	appData  []byte
	sync.Mutex
}

func (conn *edgeConn) Write(data []byte) (int, error) {
	if conn.sentFIN.Load() {
		return 0, errors.New("calling Write() after CloseWrite()")
	}

	if conn.sender != nil {
		conn.Lock()
		defer conn.Unlock()

		cipherData, err := conn.sender.Push(data, secretstream.TagMessage)
		if err != nil {
			return 0, err
		}

		_, err = conn.MsgChannel.Write(cipherData)
		return len(data), err
	} else {
		return conn.MsgChannel.Write(data)
	}
}

func (conn *edgeConn) CloseWrite() error {
	if conn.sentFIN.CompareAndSwap(false, true) {
		headers := channel.Headers{}
		headers.PutUint32Header(edge.FlagsHeader, edge.FIN)
		_, err := conn.MsgChannel.WriteTraced(nil, nil, headers)
		return err
	}

	return nil
}

func (conn *edgeConn) Inspect() string {
	result := map[string]interface{}{}
	result["id"] = conn.Id()
	result["serviceName"] = conn.serviceName
	result["closed"] = conn.closed.Load()
	result["encryptionRequired"] = conn.crypto

	if conn.connType == ConnTypeDial {
		result["encrypted"] = conn.rxKey != nil || conn.receiver != nil
		result["readFIN"] = conn.readFIN.Load()
		result["sentFIN"] = conn.sentFIN.Load()
	}

	if conn.connType == ConnTypeBind {
		hosting := map[string]interface{}{}
		for entry := range conn.hosting.IterBuffered() {
			hosting[entry.Key] = map[string]interface{}{
				"closed":      entry.Val.closed.Load(),
				"manualStart": entry.Val.manualStart,
				"serviceId":   *entry.Val.service.ID,
				"serviceName": *entry.Val.service.Name,
			}
		}
		result["hosting"] = hosting
	}

	jsonOutput, err := json.Marshal(result)
	if err != nil {
		pfxlog.Logger().WithError(err).Error("unable to marshal inspect result")
	}
	return string(jsonOutput)
}

func (conn *edgeConn) Accept(msg *channel.Message) {
	conn.TraceMsg("Accept", msg)

	if msg.ContentType == edge.ContentTypeConnInspectRequest {
		resp := edge.NewConnInspectResponse(0, edge.ConnType(conn.connType), conn.Inspect())
		if err := resp.ReplyTo(msg).Send(conn.Channel); err != nil {
			logrus.WithFields(edge.GetLoggerFields(msg)).WithError(err).
				Error("failed to send inspect response")
		}
		return
	}

	switch conn.connType {
	case ConnTypeDial:
		if msg.ContentType == edge.ContentTypeStateClosed {
			conn.sentFIN.Store(true) // if we're not closing until all reads are done, at least prevent more writes
		}

		if msg.ContentType == edge.ContentTypeTraceRoute {
			hops, _ := msg.GetUint32Header(edge.TraceHopCountHeader)
			if hops > 0 {
				hops--
				msg.PutUint32Header(edge.TraceHopCountHeader, hops)
			}

			ts, _ := msg.GetUint64Header(edge.TimestampHeader)
			connId, _ := msg.GetUint32Header(edge.ConnIdHeader)
			resp := edge.NewTraceRouteResponseMsg(connId, hops, ts, "sdk/golang", "")

			sourceRequestId, _ := msg.GetUint32Header(edge.TraceSourceRequestIdHeader)
			resp.PutUint32Header(edge.TraceSourceRequestIdHeader, sourceRequestId)

			if msgUUID := msg.Headers[edge.UUIDHeader]; msgUUID != nil {
				resp.Headers[edge.UUIDHeader] = msgUUID
			}

			if err := conn.Send(resp); err != nil {
				logrus.WithFields(edge.GetLoggerFields(msg)).WithError(err).
					Error("failed to send trace route response")
			}
			return
		}

		if err := conn.readQ.PutSequenced(msg); err != nil {
			logrus.WithFields(edge.GetLoggerFields(msg)).WithError(err).
				Error("error pushing edge message to sequencer")
		} else {
			logrus.WithFields(edge.GetLoggerFields(msg)).Debugf("received %v bytes (msg type: %v)", len(msg.Body), msg.ContentType)
		}

	case ConnTypeBind:
		if msg.ContentType == edge.ContentTypeDial {
			newConnId, _ := msg.GetUint32Header(edge.RouterProvidedConnId)
			logrus.WithFields(edge.GetLoggerFields(msg)).WithField("newConnId", newConnId).Debug("received dial request")
			go conn.newChildConnection(msg)
		} else if msg.ContentType == edge.ContentTypeStateClosed {
			conn.close(true)
		} else if msg.ContentType == edge.ContentTypeBindSuccess {
			for entry := range conn.hosting.IterBuffered() {
				entry.Val.established.Store(true)
				event := &edge.ListenerEvent{
					EventType: edge.ListenerEstablished,
				}
				select {
				case entry.Val.eventC <- event:
				default:
					logrus.WithFields(edge.GetLoggerFields(msg)).Warn("unable to send listener established event")
				}
			}
		}
	default:
		logrus.WithFields(edge.GetLoggerFields(msg)).Errorf("invalid connection type: %v", conn.connType)
	}
}

func (conn *edgeConn) IsClosed() bool {
	return conn.closed.Load()
}

func (conn *edgeConn) Network() string {
	return conn.serviceName
}

func (conn *edgeConn) String() string {
	return fmt.Sprintf("zitiConn connId=%v svcId=%v sourceIdentity=%v", conn.Id(), conn.serviceName, conn.sourceIdentity)
}

func (conn *edgeConn) LocalAddr() net.Addr {
	return conn
}

func (conn *edgeConn) RemoteAddr() net.Addr {
	return &edge.Addr{MsgCh: conn.MsgChannel}
}

func (conn *edgeConn) SourceIdentifier() string {
	return conn.sourceIdentity
}

func (conn *edgeConn) SetDeadline(t time.Time) error {
	if err := conn.SetReadDeadline(t); err != nil {
		return err
	}
	return conn.SetWriteDeadline(t)
}

func (conn *edgeConn) SetReadDeadline(t time.Time) error {
	conn.readQ.SetReadDeadline(t)
	return nil
}

func (conn *edgeConn) HandleMuxClose() error {
	conn.close(true)
	return nil
}

func (conn *edgeConn) GetCircuitId() string {
	return conn.circuitId
}

func (conn *edgeConn) GetStickinessToken() []byte {
	return conn.customState[edge.StickinessTokenHeader]
}

func (conn *edgeConn) HandleClose(channel.Channel) {
	logger := pfxlog.Logger().WithField("connId", conn.Id()).WithField("marker", conn.marker)
	defer logger.Debug("received HandleClose from underlying channel, marking conn closed")
	conn.readQ.Close()
	conn.closed.Store(true)
	conn.sentFIN.Store(true)
	conn.readFIN.Store(true)
}

func (conn *edgeConn) Connect(session *rest_model.SessionDetail, options *edge.DialOptions) (edge.Conn, error) {
	logger := pfxlog.Logger().
		WithField("marker", conn.marker).
		WithField("connId", conn.Id()).
		WithField("sessionId", session.ID)

	var pub []byte
	if conn.crypto {
		pub = conn.keyPair.Public()
	}
	connectRequest := edge.NewConnectMsg(conn.Id(), *session.Token, pub, options)
	connectRequest.Headers[edge.ConnectionMarkerHeader] = []byte(conn.marker)
	conn.TraceMsg("connect", connectRequest)
	replyMsg, err := connectRequest.WithTimeout(options.ConnectTimeout).SendForReply(conn.Channel)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if replyMsg.ContentType == edge.ContentTypeStateClosed {
		return nil, errors.Errorf("dial failed: %v", string(replyMsg.Body))
	}

	if replyMsg.ContentType != edge.ContentTypeStateConnected {
		return nil, errors.Errorf("unexpected response to connect attempt: %v", replyMsg.ContentType)
	}

	if conn.crypto {
		// There is no race condition where we can receive the other side crypto header
		// because the processing of the crypto header takes place in Conn.Read which
		// can't happen until we return the conn to the user. So as long as we send
		// the header and set rxkey before we return, we should be safe
		method, _ := replyMsg.GetByteHeader(edge.CryptoMethodHeader)
		hostPubKey := replyMsg.Headers[edge.PublicKeyHeader]
		if hostPubKey != nil {
			logger.Debug("setting up end-to-end encryption")
			if err = conn.establishClientCrypto(conn.keyPair, hostPubKey, edge.CryptoMethod(method)); err != nil {
				logger.WithError(err).Error("crypto failure")
				_ = conn.Close()
				return nil, err
			}
			logger.Debug("client tx encryption setup done")
		} else {
			logger.Warn("connection is not end-to-end-encrypted")
		}
	}
	conn.circuitId, _ = replyMsg.GetStringHeader(edge.CircuitIdHeader)
	if stickinessToken, ok := replyMsg.Headers[edge.StickinessTokenHeader]; ok {
		if conn.customState == nil {
			conn.customState = map[int32][]byte{}
		}
		conn.customState[edge.StickinessTokenHeader] = stickinessToken
	}
	logger.Debug("connected")

	return conn, nil
}

func (conn *edgeConn) establishClientCrypto(keypair *kx.KeyPair, peerKey []byte, method edge.CryptoMethod) error {
	var err error
	var rx, tx []byte

	if method != edge.CryptoMethodLibsodium {
		return unsupportedCrypto
	}

	if rx, tx, err = keypair.ClientSessionKeys(peerKey); err != nil {
		return errors.Wrap(err, "failed key exchange")
	}

	var txHeader []byte
	if conn.sender, txHeader, err = secretstream.NewEncryptor(tx); err != nil {
		return errors.Wrap(err, "failed to establish crypto stream")
	}

	conn.rxKey = rx

	if _, err = conn.MsgChannel.Write(txHeader); err != nil {
		return errors.Wrap(err, "failed to write crypto header")
	}

	pfxlog.Logger().
		WithField("connId", conn.Id()).
		WithField("marker", conn.marker).
		Debug("crypto established")
	return nil
}

func (conn *edgeConn) establishServerCrypto(keypair *kx.KeyPair, peerKey []byte, method edge.CryptoMethod) ([]byte, error) {
	var err error
	var rx, tx []byte

	if method != edge.CryptoMethodLibsodium {
		return nil, unsupportedCrypto
	}
	if rx, tx, err = keypair.ServerSessionKeys(peerKey); err != nil {
		return nil, errors.Wrap(err, "failed key exchange")
	}

	var txHeader []byte
	if conn.sender, txHeader, err = secretstream.NewEncryptor(tx); err != nil {
		return nil, errors.Wrap(err, "failed to establish crypto stream")
	}

	conn.rxKey = rx

	return txHeader, nil
}

func (conn *edgeConn) listen(session *rest_model.SessionDetail, service *rest_model.ServiceDetail, options *edge.ListenOptions) (*edgeListener, error) {
	logger := pfxlog.ContextLogger(conn.Channel.Label()).
		WithField("connId", conn.Id()).
		WithField("serviceName", *service.Name).
		WithField("sessionId", *session.ID)

	listener := &edgeListener{
		baseListener: baseListener{
			service: service,
			acceptC: make(chan edge.Conn, 10),
			errorC:  make(chan error, 1),
		},
		token:       *session.Token,
		edgeChan:    conn,
		manualStart: options.ManualStart,
		eventC:      options.GetEventChannel(),
	}
	logger.Debug("adding listener for session")
	conn.hosting.Set(*session.Token, listener)

	success := false
	defer func() {
		if !success {
			logger.Debug("removing listener for session")
			conn.unbind(logger, listener.token)
		}
	}()

	logger.Debug("sending bind request to edge router")
	var pub []byte
	if conn.crypto {
		pub = conn.keyPair.Public()
	}
	bindRequest := edge.NewBindMsg(conn.Id(), *session.Token, pub, options)
	conn.TraceMsg("listen", bindRequest)
	replyMsg, err := bindRequest.WithTimeout(5 * time.Second).SendForReply(conn.Channel)
	if err != nil {
		logger.WithError(err).Error("failed to bind")
		return nil, err
	}

	if replyMsg.ContentType == edge.ContentTypeStateClosed {
		msg := string(replyMsg.Body)
		logger.Errorf("bind request resulted in disconnect. msg: (%v)", msg)
		return nil, errors.Errorf("attempt to use closed connection: %v", msg)
	}

	if replyMsg.ContentType != edge.ContentTypeStateConnected {
		logger.Errorf("unexpected response to connect attempt: %v", replyMsg.ContentType)
		return nil, errors.Errorf("unexpected response to connect attempt: %v", replyMsg.ContentType)
	}

	success = true
	logger.Debug("connected")

	return listener, nil
}

func (conn *edgeConn) unbind(logger *logrus.Entry, token string) {
	logger.Debug("starting unbind")

	conn.hosting.Remove(token)

	unbindRequest := edge.NewUnbindMsg(conn.Id(), token)
	if err := unbindRequest.WithTimeout(5 * time.Second).SendAndWaitForWire(conn.Channel); err != nil {
		logger.WithError(err).Error("unable to send unbind msg for conn")
	} else {
		logger.Debug("unbind message sent successfully")
	}
}

func (conn *edgeConn) Read(p []byte) (int, error) {
	log := pfxlog.Logger().WithField("connId", conn.Id()).WithField("marker", conn.marker)
	if conn.closed.Load() {
		return 0, io.EOF
	}

	log.Tracef("read buffer = %d bytes", len(p))
	if len(conn.inBuffer) > 0 {
		first := conn.inBuffer[0]
		log.Tracef("found %d buffered bytes", len(first))
		n := copy(p, first)
		first = first[n:]
		if len(first) == 0 {
			conn.inBuffer = conn.inBuffer[1:]
		} else {
			conn.inBuffer[0] = first
		}
		return n, nil
	}

	for {
		if conn.readFIN.Load() {
			return 0, io.EOF
		}

		msg, err := conn.readQ.GetNext()
		if errors.Is(err, ErrClosed) {
			log.Debug("sequencer closed, closing connection")
			conn.closed.Store(true)
			return 0, io.EOF
		} else if err != nil {
			log.Debugf("unexpected sequencer err (%v)", err)
			return 0, err
		}

		flags, _ := msg.GetUint32Header(edge.FlagsHeader)
		if flags&edge.FIN != 0 {
			conn.readFIN.Store(true)
		}
		conn.flags = conn.flags | (flags & (edge.STREAM | edge.MULTIPART))

		switch msg.ContentType {

		case edge.ContentTypeStateClosed:
			log.Debug("received ConnState_CLOSED message, closing connection")
			conn.close(true)
			continue

		case edge.ContentTypeData:
			d := msg.Body
			log.Tracef("got buffer from sequencer %d bytes", len(d))
			if len(d) == 0 && conn.readFIN.Load() {
				return 0, io.EOF
			}

			multipart := (flags & edge.MULTIPART_MSG) != 0

			// first data message should contain crypto header
			if conn.rxKey != nil {
				if len(d) != secretstream.StreamHeaderBytes {
					return 0, errors.Errorf("failed to receive crypto header bytes: read[%d]", len(d))
				}
				conn.receiver, err = secretstream.NewDecryptor(conn.rxKey, d)
				if err != nil {
					return 0, errors.Wrap(err, "failed to init decryptor")
				}
				conn.rxKey = nil
				continue
			}

			if conn.receiver != nil {
				d, _, err = conn.receiver.Pull(d)
				if err != nil {
					log.WithFields(edge.GetLoggerFields(msg)).Errorf("crypto failed on msg of size=%v, headers=%+v err=(%v)", len(msg.Body), msg.Headers, err)
					return 0, err
				}
			}
			n := 0
			if multipart && len(d) > 0 {
				var parts [][]byte
				for len(d) > 0 {
					l := binary.LittleEndian.Uint16(d[0:2])
					d = d[2:]
					part := d[0:l]
					d = d[l:]
					parts = append(parts, part)
				}
				n = copy(p, parts[0])
				parts[0] = parts[0][n:]
				if len(parts[0]) == 0 {
					parts = parts[1:]
				}
				conn.inBuffer = append(conn.inBuffer, parts...)
			} else {
				n = copy(p, d)
				d = d[n:]
				if len(d) > 0 {
					conn.inBuffer = append(conn.inBuffer, d)
				}
			}

			log.Tracef("%d chunks in incoming buffer", len(conn.inBuffer))
			log.Debugf("read %v bytes", n)
			return n, nil

		default:
			log.WithField("type", msg.ContentType).Error("unexpected message")
		}
	}
}

func (conn *edgeConn) Close() error {
	conn.close(false)
	return nil
}

func (conn *edgeConn) close(closedByRemote bool) {
	// everything in here should be safe to execute concurrently from outside the muxer loop,
	// except the remove from mux call
	if !conn.closed.CompareAndSwap(false, true) {
		return
	}
	conn.readFIN.Store(true)
	conn.sentFIN.Store(true)

	log := pfxlog.Logger().WithField("connId", conn.Id()).WithField("marker", conn.marker)
	log.Debug("close: begin")
	defer log.Debug("close: end")

	if !closedByRemote {
		msg := edge.NewStateClosedMsg(conn.Id(), "")
		if err := conn.SendState(msg); err != nil {
			log.WithError(err).Error("failed to send close message")
		}
	}

	conn.readQ.Close()
	conn.msgMux.RemoveMsgSink(conn) // if we switch back to ChMsgMux will need to be done async again, otherwise we may deadlock

	if conn.connType == ConnTypeBind {
		for entry := range conn.hosting.IterBuffered() {
			listener := entry.Val
			if err := listener.close(closedByRemote); err != nil {
				log.WithError(err).WithField("serviceName", *listener.service.Name).Error("failed to close listener")
			}
		}
	}
}

func (conn *edgeConn) getListener(token string) (*edgeListener, bool) {
	return conn.hosting.Get(token)
}

func (conn *edgeConn) newChildConnection(message *channel.Message) {
	token := string(message.Body)
	circuitId, _ := message.GetStringHeader(edge.CircuitIdHeader)
	logger := pfxlog.Logger().WithField("connId", conn.Id())
	if circuitId != "" {
		logger = logger.WithField("circuitId", circuitId)
	}
	logger.WithField("token", token).Debug("logging token")

	logger.Debug("looking up listener")
	listener, found := conn.getListener(token)
	if !found {
		logger.Warn("listener not found")
		reply := edge.NewDialFailedMsg(conn.Id(), "invalid token")
		reply.ReplyTo(message)
		if err := reply.WithPriority(channel.Highest).WithTimeout(5 * time.Second).SendAndWaitForWire(conn.Channel); err != nil {
			logger.WithError(err).Error("failed to send reply to dial request")
		}
		return
	}

	logger.Debug("listener found. checking for router provided connection id")

	id, routerProvidedConnId := message.GetUint32Header(edge.RouterProvidedConnId)
	if routerProvidedConnId {
		logger.Debugf("using router provided connection id %v", id)
	} else {
		id = conn.msgMux.GetNextId()
		logger.Debugf("listener found. generating id for new connection: %v", id)
	}

	sourceIdentity, _ := message.GetStringHeader(edge.CallerIdHeader)
	marker, _ := message.GetStringHeader(edge.ConnectionMarkerHeader)

	edgeCh := &edgeConn{
		MsgChannel:     *edge.NewEdgeMsgChannel(conn.Channel, id),
		readQ:          NewNoopSequencer[*channel.Message](4),
		msgMux:         conn.msgMux,
		sourceIdentity: sourceIdentity,
		crypto:         conn.crypto,
		appData:        message.Headers[edge.AppDataHeader],
		connType:       ConnTypeDial,
		marker:         marker,
		circuitId:      circuitId,
	}

	newConnLogger := pfxlog.Logger().
		WithField("marker", marker).
		WithField("connId", id).
		WithField("parentConnId", conn.Id()).
		WithField("token", token).
		WithField("circuitId", token)

	err := conn.msgMux.AddMsgSink(edgeCh) // duplicate errors only happen on the server side, since client controls ids
	if err != nil {
		newConnLogger.WithError(err).Error("invalid conn id, already in use")
		reply := edge.NewDialFailedMsg(conn.Id(), err.Error())
		reply.ReplyTo(message)
		if err := reply.WithPriority(channel.Highest).WithTimeout(5 * time.Second).SendAndWaitForWire(conn.Channel); err != nil {
			logger.WithError(err).Error("failed to send reply to dial request")
		}
		return
	}

	var txHeader []byte
	if edgeCh.crypto {
		newConnLogger.Debug("setting up crypto")
		clientKey := message.Headers[edge.PublicKeyHeader]
		method, _ := message.GetByteHeader(edge.CryptoMethodHeader)

		if clientKey != nil {
			if txHeader, err = edgeCh.establishServerCrypto(conn.keyPair, clientKey, edge.CryptoMethod(method)); err != nil {
				logger.WithError(err).Error("failed to establish crypto session")
			}
		} else {
			newConnLogger.Warnf("client did not send its key. connection is not end-to-end encrypted")
		}
	}

	if err != nil {
		newConnLogger.WithError(err).Error("failed to establish connection")
		reply := edge.NewDialFailedMsg(conn.Id(), err.Error())
		reply.ReplyTo(message)
		if err := reply.WithPriority(channel.Highest).WithTimeout(5 * time.Second).SendAndWaitForWire(conn.Channel); err != nil {
			logger.WithError(err).Error("failed to send reply to dial request")
		}
		return
	}

	connHandler := &newConnHandler{
		conn:                 conn,
		edgeCh:               edgeCh,
		message:              message,
		txHeader:             txHeader,
		routerProvidedConnId: routerProvidedConnId,
		circuitId:            circuitId,
	}

	if listener.manualStart {
		edgeCh.acceptCompleteHandler = connHandler
	} else if err := connHandler.dialSucceeded(); err != nil {
		logger.Debug("calling dial succeeded")
		return
	}

	listener.acceptC <- edgeCh
}

func (conn *edgeConn) GetAppData() []byte {
	return conn.appData
}

func (conn *edgeConn) CompleteAcceptSuccess() error {
	if conn.acceptCompleteHandler != nil {
		result := conn.acceptCompleteHandler.dialSucceeded()
		conn.acceptCompleteHandler = nil
		return result
	}
	return nil
}

func (conn *edgeConn) CompleteAcceptFailed(err error) {
	if conn.acceptCompleteHandler != nil {
		conn.acceptCompleteHandler.dialFailed(err)
		conn.acceptCompleteHandler = nil
	}
}

func (conn *edgeConn) TraceRoute(hops uint32, timeout time.Duration) (*edge.TraceRouteResult, error) {
	msg := edge.NewTraceRouteMsg(conn.Id(), hops, uint64(info.NowInMilliseconds()))
	resp, err := msg.WithTimeout(timeout).SendForReply(conn.Channel)
	if err != nil {
		return nil, err
	}
	if resp.ContentType != edge.ContentTypeTraceRouteResponse {
		return nil, errors.Errorf("unexpected response: %v", resp.ContentType)
	}
	hops, _ = resp.GetUint32Header(edge.TraceHopCountHeader)
	ts, _ := resp.GetUint64Header(edge.TimestampHeader)
	elapsed := time.Duration(0)
	if ts > 0 {
		elapsed = time.Duration(info.NowInMilliseconds()-int64(ts)) * time.Millisecond
	}
	hopType, _ := resp.GetStringHeader(edge.TraceHopTypeHeader)
	hopId, _ := resp.GetStringHeader(edge.TraceHopIdHeader)
	hopErr, _ := resp.GetStringHeader(edge.TraceError)

	result := &edge.TraceRouteResult{
		Hops:    hops,
		Time:    elapsed,
		HopType: hopType,
		HopId:   hopId,
		Error:   hopErr,
	}
	return result, nil
}

type newConnHandler struct {
	conn                 *edgeConn
	edgeCh               *edgeConn
	message              *channel.Message
	txHeader             []byte
	routerProvidedConnId bool
	circuitId            string
}

func (self *newConnHandler) dialFailed(err error) {
	token := string(self.message.Body)
	logger := pfxlog.Logger().WithField("connId", self.conn.Id()).WithField("token", token)

	newConnLogger := pfxlog.Logger().
		WithField("connId", self.edgeCh.Id()).
		WithField("parentConnId", self.conn.Id()).
		WithField("token", token)

	newConnLogger.WithError(err).Error("Failed to establish connection")
	reply := edge.NewDialFailedMsg(self.conn.Id(), err.Error())
	reply.ReplyTo(self.message)
	if err := reply.WithPriority(channel.Highest).WithTimeout(5 * time.Second).SendAndWaitForWire(self.conn.Channel); err != nil {
		logger.WithError(err).Error("Failed to send reply to dial request")
	}
}

func (self *newConnHandler) dialSucceeded() error {
	logger := pfxlog.Logger().WithField("connId", self.conn.Id()).WithField("circuitId", self.circuitId)

	newConnLogger := pfxlog.Logger().
		WithField("connId", self.edgeCh.Id()).
		WithField("marker", self.edgeCh.marker).
		WithField("parentConnId", self.conn.Id()).
		WithField("circuitId", self.circuitId)

	newConnLogger.Debug("new connection established")

	reply := edge.NewDialSuccessMsg(self.conn.Id(), self.edgeCh.Id())
	reply.ReplyTo(self.message)

	if !self.routerProvidedConnId {
		startMsg, err := reply.WithPriority(channel.Highest).WithTimeout(5 * time.Second).SendForReply(self.conn.Channel)
		if err != nil {
			logger.WithError(err).Error("Failed to send reply to dial request")
			return err
		}

		if startMsg.ContentType != edge.ContentTypeStateConnected {
			logger.Errorf("failed to receive start after dial. got %v", startMsg)
			return errors.Errorf("failed to receive start after dial. got %v", startMsg)
		}
	} else if err := reply.WithPriority(channel.Highest).WithTimeout(time.Second * 5).SendAndWaitForWire(self.conn.Channel); err != nil {
		logger.WithError(err).Error("Failed to send reply to dial request")
		return err
	}

	if self.txHeader != nil {
		newConnLogger.Debug("sending crypto header")
		if _, err := self.edgeCh.MsgChannel.Write(self.txHeader); err != nil {
			newConnLogger.WithError(err).Error("failed to write crypto header")
			return err
		}
		newConnLogger.Debug("tx crypto established")
	}
	return nil
}

// make a random 8 byte string
func newMarker() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
