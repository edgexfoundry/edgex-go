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
	"context"
	"fmt"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v4"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/inspect"
	"github.com/openziti/sdk-golang/xgress"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/secretstream/kx"
	"github.com/pkg/errors"
)

type RouterConnOwner interface {
	OnClose(factory edge.RouterConn)
}

type routerConn struct {
	routerName string
	routerAddr string
	ch         edge.SdkChannel
	mux        edge.ConnMux[any]
	owner      RouterConnOwner
}

func (conn *routerConn) GetBoolHeader(key int32) bool {
	val := conn.ch.GetChannel().Headers()[key]
	return len(val) == 1 && val[0] == 1
}

func (conn *routerConn) GetRouterAddr() string {
	return conn.routerAddr
}

func (conn *routerConn) GetRouterName() string {
	return conn.routerName
}

func (conn *routerConn) Inspect() *inspect.RouterConnInspectDetail {
	result := &inspect.RouterConnInspectDetail{
		RouterName: conn.routerName,
		RouterAddr: conn.routerAddr,
		Closed:     conn.IsClosed(),
	}
	for _, sink := range conn.mux.GetSinks() {
		if inspectable, ok := sink.(interface {
			InspectSink() *inspect.VirtualConnDetail
		}); ok {
			result.VirtualConns = append(result.VirtualConns, inspectable.InspectSink())
		}
	}
	return result
}

func (conn *routerConn) HandleClose(channel.Channel) {
	if conn.owner != nil {
		conn.owner.OnClose(conn)
	}
}

func NewRouterConn(routerName, routerAddr string, owner RouterConnOwner, inspectF func() *inspect.ContextInspectResult) edge.RouterConn {
	conn := &routerConn{
		routerAddr: routerAddr,
		routerName: routerName,
		mux:        edge.NewChannelConnMapMux[any](inspectF),
		owner:      owner,
	}

	return conn
}

func (conn *routerConn) BindChannel(binding channel.Binding) error {
	if multiChannel, ok := binding.GetChannel().(channel.MultiChannel); ok {
		conn.ch = multiChannel.GetUnderlayHandler().(edge.SdkChannel)
	} else {
		conn.ch = edge.NewSingleSdkChannel(binding.GetChannel())
	}

	binding.AddReceiveHandlerF(edge.ContentTypeDial, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeStateClosed, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeTraceRoute, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeConnInspectRequest, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeBindSuccess, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeXgPayload, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeXgAcknowledgement, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeXgControl, conn.mux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeInspectRequest, conn.mux.HandleReceive)

	// Since data is the common message type, it gets to be dispatched directly
	binding.AddTypedReceiveHandler(conn.mux)
	binding.AddCloseHandler(conn.mux)
	binding.AddCloseHandler(conn)

	return nil
}

func (conn *routerConn) NewDialConn(service *rest_model.ServiceDetail) *edgeConn {
	id := conn.mux.GetNextId()

	closeNotify := make(chan struct{})
	edgeCh := &edgeConn{
		closeNotify: closeNotify,
		MsgChannel:  *edge.NewEdgeMsgChannel(conn.ch, id),
		readQ:       NewNoopSequencer[*channel.Message](closeNotify, 4),
		msgMux:      conn.mux,
		serviceName: *service.Name,
		marker:      newMarker(),
	}

	var err error
	if *service.EncryptionRequired {
		if edgeCh.keyPair, err = kx.NewKeyPair(); err == nil {
			edgeCh.crypto = true
		} else {
			pfxlog.Logger().Errorf("unable to setup encryption for edgeConn[%s] %v", *service.Name, err)
		}
	}

	err = conn.mux.Add(edgeCh) // duplicate errors only happen on the server side, since client controls ids
	if err != nil {
		pfxlog.Logger().Warnf("error adding message sink %s[%d]: %v", *service.Name, id, err)
	}
	return edgeCh
}

func (conn *routerConn) SendPosture(responses []rest_model.PostureResponseCreate) error {
	message := edge.NewPostureResponsesMsg(responses)
	sendErr := message.Send(conn.ch.GetControlSender())

	if sendErr != nil {
		return sendErr
	}

	return nil
}

func (conn *routerConn) UpdateToken(token []byte, timeout time.Duration) error {
	msg := edge.NewUpdateTokenMsg(token)
	resp, err := msg.WithTimeout(timeout).SendForReply(conn.ch.GetControlSender())

	if err != nil {
		return err
	}

	if resp.ContentType == edge.ContentTypeUpdateTokenSuccess {
		return nil
	}

	if resp.ContentType == edge.ContentTypeUpdateTokenFailure {
		err = errors.New(string(resp.Body))
		return fmt.Errorf("could not update token for router [%s]: %w", conn.GetRouterAddr(), err)
	}

	err = fmt.Errorf("invalid content type response %d, expected one of [%d, %d]", resp.ContentType, edge.ContentTypeUpdateTokenSuccess, edge.ContentTypeUpdateTokenFailure)
	return fmt.Errorf("could not update token for router [%s]: %w", conn.GetRouterAddr(), err)
}

func (conn *routerConn) NewListenConn(service *rest_model.ServiceDetail, session *rest_model.SessionDetail, options *edge.ListenOptions, envF func() xgress.Env) *edgeHostConn {
	id := conn.mux.GetNextId()

	edgeCh := &edgeHostConn{
		MsgChannel:   *edge.NewEdgeMsgChannel(conn.ch, id),
		msgMux:       conn.mux,
		serviceName:  *service.Name,
		routerInfo:   edge.EdgeRouterInfo{Name: conn.routerName, Addr: conn.routerAddr},
		keyPair:      options.KeyPair,
		crypto:       options.KeyPair != nil,
		service:      service,
		acceptC:      make(chan edge.Conn, 10),
		token:        *session.Token,
		manualStart:  options.ManualStart,
		eventHandler: options.EventHandler,
		envF:         envF,
	}

	if options.DoNotSaveDialerIdentity {
		edgeCh.flags.Set(hostConnDoNotSaveDialerIdentity, true)
	}

	// duplicate errors only happen on the server side, since the client controls ids
	if err := conn.mux.Add(edgeCh); err != nil {
		pfxlog.Logger().Warnf("error adding message sink %s[%d]: %v", *service.Name, id, err)
	}

	pfxlog.Logger().WithField("connId", id).
		WithField("routerName", conn.routerName).
		WithField("serviceId", *service.ID).
		WithField("serviceName", *service.Name).
		Debug("created new listener connection")

	return edgeCh
}

func (conn *routerConn) Connect(ctx context.Context, service *rest_model.ServiceDetail, session *rest_model.SessionDetail, options *edge.DialOptions, envF func() xgress.Env) (edge.Conn, error) {
	ec := conn.NewDialConn(service)
	dialConn, err := ec.Connect(ctx, session, options, envF)
	if err != nil {
		if !conn.ch.GetChannel().IsClosed() {
			if err2 := ec.Close(); err2 != nil {
				pfxlog.Logger().Errorf("failed to cleanup connection for service '%v' (%v)", service.Name, err2)
			}
		}
	}
	return dialConn, err
}

func (conn *routerConn) Listen(service *rest_model.ServiceDetail, session *rest_model.SessionDetail, options *edge.ListenOptions, envF func() xgress.Env) (edge.RouterHostConn, error) {
	ec := conn.NewListenConn(service, session, options, envF)

	log := pfxlog.Logger().
		WithField("connId", ec.Id()).
		WithField("router", conn.routerName).
		WithField("serviceId", *service.ID).
		WithField("serviceName", *service.Name)

	if err := ec.listen(session, service, options); err != nil {
		log.WithError(err).Error("failed to establish listener")

		if closeErr := ec.Close(); closeErr != nil {
			log.WithError(closeErr).Error("failed to cleanup listener for service after failed bind")
		}
		return nil, err
	}

	if !conn.GetBoolHeader(edge.SupportsBindSuccessHeader) {
		ec.established.Store(true)
	}

	log.Debug("established listener")
	return ec, nil
}

func (conn *routerConn) Close() error {
	return conn.ch.GetChannel().Close()
}

func (conn *routerConn) IsClosed() bool {
	return conn.ch.GetChannel().IsClosed()
}
