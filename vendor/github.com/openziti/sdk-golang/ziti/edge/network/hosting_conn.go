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
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v4"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/sdk-golang/inspect"
	"github.com/openziti/sdk-golang/pb/edge_client_pb"
	"github.com/openziti/sdk-golang/xgress"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/secretstream/kx"
	"github.com/sirupsen/logrus"
)

const (
	hostConnClosedFlag              = 0
	hostConnDoNotSaveDialerIdentity = 1
)

// edgeHostConn represents a service hosting connection that acts as a "receptionist"
// for incoming client dial requests. It implements edge.MsgSink to handle service-level
// messages and manages multiple client connections through its embedded ConnMux.
//
// Architecture:
//   - Receives dial requests from clients wanting to connect to the hosted service
//   - Creates individual edgeConn instances for each accepted client connection
//   - Routes ongoing client messages directly to their respective edgeConn via msgMux
//   - Manages service lifecycle (bind, unbind, close) with the edge router
//
// Message Flow:
//  1. Client sends dial request → edgeHostConn.AcceptMessage() handles it
//  2. edgeHostConn creates new edgeConn for the client
//  3. edgeConn is added to msgMux for future message routing
//  4. Client's data messages bypass edgeHostConn and go directly to edgeConn.Accept()
//
// Thread Safety: All methods are safe for concurrent use.
type edgeHostConn struct {
	// MsgChannel provides the underlying channel communication capabilities
	// for sending messages back to the edge router (bind requests, state changes, etc.)
	edge.MsgChannel

	// msgMux manages individual client connections created from dial requests.
	// Each accepted client gets an edgeConn that is registered with this mux.
	// Future messages for specific clients are routed directly to their edgeConn.
	msgMux edge.ConnMux[any]

	// serviceName is the name of the service being hosted by this connection.
	// Used for logging and debugging purposes.
	serviceName string

	// marker is a unique identifier for this hosting connection instance.
	// Used for tracing and debugging across the distributed system.
	marker string

	// crypto indicates whether end-to-end encryption is required for client connections.
	// When true, client connections must establish encrypted sessions using keyPair.
	crypto bool

	// keyPair contains the cryptographic keys used for end-to-end encryption
	// when crypto is enabled. Used during client connection handshake.
	keyPair *kx.KeyPair

	// data stores arbitrary service-level context information that can be
	// accessed by the hosting application. This might include service configuration,
	// authentication policies, metrics collectors, or other service-wide state.
	// Unlike client-specific data in edgeConn, this context applies to the entire service.
	data atomic.Value

	flags   concurrenz.AtomicBitSet
	service *rest_model.ServiceDetail
	acceptC chan edge.Conn

	routerInfo   edge.EdgeRouterInfo
	token        string
	manualStart  bool
	established  atomic.Bool
	eventHandler edge.ListenerEventHandler
	envF         func() xgress.Env
}

// GetEdgeRouterInfo returns the name and address of the edge router for this hosting connection.
func (conn *edgeHostConn) GetEdgeRouterInfo() edge.EdgeRouterInfo {
	return conn.routerInfo
}

// GetData retrieves arbitrary service-level context data associated with this hosting connection.
// This allows hosting applications to store and retrieve service-wide configuration,
// state, or metadata that applies to all clients of this service.
//
// Returns:
//   - any: the stored service context data, or nil if none has been set
//
// Examples of service-level data:
//   - Service configuration and feature flags
//   - Authentication and authorization policies
//   - Metrics collectors or connection limits
//   - Custom service handlers or middleware
func (conn *edgeHostConn) GetData() any {
	return conn.data.Load()
}

// SetData stores arbitrary service-level context data for this hosting connection.
// This data persists for the lifetime of the service and can be accessed during
// client connection handling or other service operations.
//
// Parameters:
//   - data: arbitrary context data to associate with this service
//
// Thread Safety: This method is safe for concurrent use.
func (conn *edgeHostConn) SetData(data any) {
	conn.data.Store(data)
}

// AcceptMessage implements edge.MsgSink and handles service-level messages for this hosting connection.
// This method acts as the "receptionist" that processes incoming requests and manages
// the service lifecycle. It does NOT handle individual client data messages - those
// are routed directly to the appropriate edgeConn via msgMux.
//
// Handled Message Types:
//   - ContentTypeDial: Creates new client connections for dial requests
//   - ContentTypeStateClosed: Handles service shutdown notifications
//   - ContentTypeBindSuccess: Confirms service binding and notifies listeners
//   - ContentTypeConnInspectRequest: Provides service inspection data
//
// Parameters:
//   - msg: the incoming message to process
//
// Thread Safety: This method is safe for concurrent use.
func (conn *edgeHostConn) AcceptMessage(msg *channel.Message) {
	conn.TraceMsg("AcceptMessage", msg)

	switch msg.ContentType {
	case edge.ContentTypeConnInspectRequest:
		go conn.handleInspect(msg)
	case edge.ContentTypeDial:
		newConnId, _ := msg.GetUint32Header(edge.RouterProvidedConnId)
		circuitId, _ := msg.GetStringHeader(edge.CircuitIdHeader)
		logrus.WithFields(edge.GetLoggerFields(msg)).
			WithField("circuitId", circuitId).
			WithField("newConnId", newConnId).Debug("received dial request")
		go conn.newChildConnection(msg)
	case edge.ContentTypeStateClosed:
		if err := edge.ErrorFromMsg(msg); err != nil {
			log := pfxlog.Logger().WithFields(edge.GetLoggerFields(msg)).
				WithField("retryHint", err.RetryHint.String())
			if errName, ok := edge_client_pb.Error_name[int32(err.Code)]; ok && err.Code > 0 {
				log = log.WithField("errorCode", errName)
			} else {
				log = log.WithField("errorCode", err.Code)
			}

			log.Errorf("router reported hosting error: %s", err.Message)

			switch err.RetryHint {
			case edge.RetryStartOver:
				conn.eventHandler.NotifyStartOver()
			case edge.RetryNotRetriable:
				conn.eventHandler.NotifyNotRetriable()
			}
		}

		conn.closeAndLogError(true)
	case edge.ContentTypeBindSuccess:
		conn.established.Store(true)
		conn.eventHandler.NotifyEstablished()
	}
}

func (conn *edgeHostConn) handleInspect(msg *channel.Message) {
	// note, until 1.5 this returned 0 for the connId
	resp := edge.NewConnInspectResponse(conn.Id(), edge.ConnTypeBind, conn.Inspect())
	if err := resp.ReplyTo(msg).Send(conn.GetControlSender()); err != nil {
		logrus.WithFields(edge.GetLoggerFields(msg)).WithError(err).
			Error("failed to send inspect response")
	}
}

func (conn *edgeHostConn) UpdateCost(cost uint16) error {
	return conn.updateCostAndPrecedence(&cost, nil)
}

func (conn *edgeHostConn) UpdatePrecedence(precedence edge.Precedence) error {
	return conn.updateCostAndPrecedence(nil, &precedence)
}

func (conn *edgeHostConn) UpdateCostAndPrecedence(cost uint16, precedence edge.Precedence) error {
	return conn.updateCostAndPrecedence(&cost, &precedence)
}

func (conn *edgeHostConn) updateCostAndPrecedence(cost *uint16, precedence *edge.Precedence) error {
	logger := pfxlog.Logger().
		WithField("connId", conn.Id()).
		WithField("serviceName", conn.serviceName).
		WithField("session", conn.token)

	logger.Debug("sending update bind request to edge router")
	request := edge.NewUpdateBindMsg(conn.Id(), conn.token, cost, precedence)
	conn.TraceMsg("updateCostAndPrecedence", request)
	return request.WithTimeout(5 * time.Second).SendAndWaitForWire(conn.GetControlSender())
}

func (conn *edgeHostConn) SendHealthEvent(pass bool) error {
	logger := pfxlog.Logger().
		WithField("connId", conn.Id()).
		WithField("serviceName", conn.serviceName).
		WithField("session", conn.token).
		WithField("health.status", pass)

	logger.Debug("sending health event to edge router")
	request := edge.NewHealthEventMsg(conn.Id(), conn.token, pass)
	conn.TraceMsg("healthEvent", request)
	return request.WithTimeout(5 * time.Second).SendAndWaitForWire(conn.GetControlSender())
}

func (conn *edgeHostConn) InspectSink() *inspect.VirtualConnDetail {
	return &inspect.VirtualConnDetail{
		ConnId:      conn.Id(),
		SinkType:    "host",
		ServiceName: conn.serviceName,
		Closed:      conn.flags.IsSet(hostConnClosedFlag),
	}
}

func (conn *edgeHostConn) Inspect() string {
	result := map[string]interface{}{}
	result["id"] = conn.Id()
	result["serviceName"] = conn.serviceName
	result["closed"] = conn.flags.IsSet(hostConnClosedFlag)
	result["encryptionRequired"] = conn.crypto

	result["listener"] = map[string]interface{}{
		"closed":      conn.flags.IsSet(hostConnClosedFlag),
		"manualStart": conn.manualStart,
		"serviceId":   *conn.service.ID,
		"serviceName": *conn.service.Name,
	}

	jsonOutput, err := json.Marshal(result)
	if err != nil {
		pfxlog.Logger().WithError(err).Error("unable to marshal inspect result")
	}
	return string(jsonOutput)
}

func (conn *edgeHostConn) newChildConnection(message *channel.Message) {
	token := string(message.Body)
	circuitId, _ := message.GetStringHeader(edge.CircuitIdHeader)
	logger := pfxlog.Logger().WithField("connId", conn.Id())
	if circuitId != "" {
		logger = logger.WithField("circuitId", circuitId)
	}
	logger.WithField("token", token).Debug("logging token")

	logger.Debug("checking token")
	if conn.token != token {
		logger.Warn("invalid token")
		reply := edge.NewDialFailedMsg(conn.Id(), "invalid token")
		reply.ReplyTo(message)
		if err := reply.WithPriority(channel.Highest).WithTimeout(5 * time.Second).SendAndWaitForWire(conn.GetControlSender()); err != nil {
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

	closeNotify := make(chan struct{})
	edgeCh := &edgeConn{
		closeNotify:    closeNotify,
		MsgChannel:     *edge.NewEdgeMsgChannel(conn.SdkChannel, id),
		readQ:          NewNoopSequencer[*channel.Message](closeNotify, 4),
		msgMux:         conn.msgMux,
		sourceIdentity: sourceIdentity,
		crypto:         conn.crypto,
		appData:        message.Headers[edge.AppDataHeader],
		marker:         marker,
		circuitId:      circuitId,
	}

	if !conn.flags.IsSet(hostConnDoNotSaveDialerIdentity) {
		if edgeCh.customState == nil {
			edgeCh.customState = map[int32][]byte{}
		}

		if dialerIdentityId, ok := message.Headers[edge.DialerIdentityId]; ok {
			edgeCh.customState[edge.DialerIdentityId] = dialerIdentityId
		}
		if dialerIdentityName, ok := message.Headers[edge.DialerIdentityName]; ok {
			edgeCh.customState[edge.DialerIdentityName] = dialerIdentityName
		}
	}

	cleanupAndReportError := func(description string, err error) {
		logger.WithError(err).Error(description)

		edgeCh.close(false)

		reply := edge.NewDialFailedMsg(conn.Id(), fmt.Sprintf("%s (%s)", description, err.Error()))
		reply.ReplyTo(message)
		if sendErr := reply.WithPriority(channel.Highest).WithTimeout(5 * time.Second).SendAndWaitForWire(conn.GetControlSender()); sendErr != nil {
			logger.WithError(sendErr).Error("failed to send reply to dial request")
		}
	}

	newConnLogger := pfxlog.Logger().
		WithField("marker", marker).
		WithField("connId", id).
		WithField("parentConnId", conn.Id()).
		WithField("token", token).
		WithField("circuitId", circuitId)

	// duplicate errors only happen on the server side, since client controls ids
	if err := conn.msgMux.Add(edgeCh); err != nil {
		newConnLogger.WithError(err).Error("invalid conn id, already in use")
		cleanupAndReportError("invalid connection id, already in use", err)
		return
	}

	if err := edgeCh.setupFlowControl(message, xgress.Terminator, conn.envF); err != nil {
		cleanupAndReportError("failed to start flow control", err)
		return
	}

	var txHeader []byte
	if edgeCh.crypto {
		newConnLogger.Debug("setting up crypto")
		clientKey := message.Headers[edge.PublicKeyHeader]
		method, _ := message.GetByteHeader(edge.CryptoMethodHeader)

		if clientKey != nil {
			var err error
			if txHeader, err = edgeCh.establishServerCrypto(conn.keyPair, clientKey, edge.CryptoMethod(method)); err != nil {
				cleanupAndReportError("failed to establish crypto session", err)
				return
			}
		} else {
			newConnLogger.Warnf("client did not send its key. connection is not end-to-end encrypted")
		}
	}

	connHandler := &newConnHandler{
		conn:                 conn,
		edgeCh:               edgeCh,
		message:              message,
		txHeader:             txHeader,
		routerProvidedConnId: routerProvidedConnId,
		circuitId:            circuitId,
	}

	if conn.manualStart {
		edgeCh.acceptCompleteHandler = connHandler
	} else if err, cleanupHandled := connHandler.dialSucceeded(); err != nil {
		if !cleanupHandled {
			cleanupAndReportError("failed to start connection", err)
		}
		return
	}

	newConnLogger.Debug("dial succeeded")

	conn.acceptC <- edgeCh
}

func (conn *edgeHostConn) HandleMuxClose() error {
	return conn.close(true)
}

func (conn *edgeHostConn) Close() error {
	return conn.close(false)
}

func (conn *edgeHostConn) closeAndLogError(closedByRemote bool) {
	if err := conn.close(closedByRemote); err != nil {
		pfxlog.Logger().WithError(err).Error("error closing connection")
	}
}

func (conn *edgeHostConn) close(closedByRemote bool) error {
	// everything in here should be safe to execute concurrently from outside the muxer loop,
	// except the remove from mux call
	if !conn.flags.CompareAndSet(hostConnClosedFlag, false, true) {
		return nil
	}

	conn.msgMux.Remove(conn)

	defer func() {
		conn.acceptC <- nil // signal listeners that listener is closed
	}()

	log := pfxlog.Logger().
		WithField("connId", conn.Id()).
		WithField("sessionId", conn.token).
		WithField("marker", conn.marker).
		WithField("serviceName", *conn.service.Name)
	log.Debug("removing listener for session")

	log.Debug("close: begin")
	defer log.Debug("close: end")

	var errList []error

	if !closedByRemote {
		unbindRequest := edge.NewUnbindMsg(conn.Id(), conn.token)
		conn.TraceMsg("close", unbindRequest)
		if err := unbindRequest.WithTimeout(5 * time.Second).SendAndWaitForWire(conn.GetControlSender()); err != nil {
			log.WithError(err).Error("unable to unbind session for conn")
			errList = append(errList, err)
		}

		msg := edge.NewStateClosedMsg(conn.Id(), "")
		if err := conn.SendState(msg); err != nil {
			log.WithError(err).Error("failed to send close message")
			errList = append(errList, err)
		}
	}

	return errors.Join(errList...)
}

func (conn *edgeHostConn) listen(session *rest_model.SessionDetail, service *rest_model.ServiceDetail, options *edge.ListenOptions) error {
	logger := pfxlog.ContextLogger(conn.GetChannel().Label()).
		WithField("connId", conn.Id()).
		WithField("serviceName", *service.Name).
		WithField("sessionId", *session.ID)

	success := false
	defer func() {
		if !success {
			logger.Debug("removing listener for session")
			conn.unbind(logger, conn.token)
		}
	}()

	logger.Debug("sending bind request to edge router")
	var pub []byte
	if conn.crypto {
		pub = conn.keyPair.Public()
	}
	bindRequest := edge.NewBindMsg(conn.Id(), *session.Token, pub, options)
	conn.TraceMsg("listen", bindRequest)
	replyMsg, err := bindRequest.WithTimeout(5 * time.Second).SendForReply(conn.GetControlSender())
	if err != nil {
		logger.WithError(err).Error("failed to bind")
		return err
	}

	if replyMsg.ContentType == edge.ContentTypeStateClosed {
		msg := string(replyMsg.Body)
		logger.Errorf("bind request resulted in disconnect. msg: (%v)", msg)
		return fmt.Errorf("attempt to use closed connection: %v", msg)
	}

	if replyMsg.ContentType != edge.ContentTypeStateConnected {
		logger.Errorf("unexpected response to connect attempt: %v", replyMsg.ContentType)
		return fmt.Errorf("unexpected response to connect attempt: %v", replyMsg.ContentType)
	}

	success = true
	logger.Debug("connected")

	return nil
}

func (conn *edgeHostConn) unbind(logger *logrus.Entry, token string) {
	logger.Debug("starting unbind")

	unbindRequest := edge.NewUnbindMsg(conn.Id(), token)
	if err := unbindRequest.WithTimeout(5 * time.Second).SendAndWaitForWire(conn.GetControlSender()); err != nil {
		logger.WithError(err).Error("unable to send unbind msg for conn")
	} else {
		logger.Debug("unbind message sent successfully")
	}
}
