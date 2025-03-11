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
	"fmt"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/secretstream/kx"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
)

type RouterConnOwner interface {
	OnClose(factory edge.RouterConn)
}

type routerConn struct {
	routerName string
	key        string
	ch         channel.Channel
	msgMux     edge.MsgMux
	owner      RouterConnOwner
}

func (conn *routerConn) GetBoolHeader(key int32) bool {
	val := conn.ch.Underlay().Headers()[key]
	return len(val) == 1 && val[0] == 1
}

func (conn *routerConn) Key() string {
	return conn.key
}

func (conn *routerConn) GetRouterName() string {
	return conn.routerName
}

func (conn *routerConn) HandleClose(channel.Channel) {
	if conn.owner != nil {
		conn.owner.OnClose(conn)
	}
}

func NewEdgeConnFactory(routerName, key string, owner RouterConnOwner) edge.RouterConn {
	connFactory := &routerConn{
		key:        key,
		routerName: routerName,
		msgMux:     edge.NewCowMapMsgMux(),
		owner:      owner,
	}

	return connFactory
}

func (conn *routerConn) BindChannel(binding channel.Binding) error {
	conn.ch = binding.GetChannel()

	binding.AddReceiveHandlerF(edge.ContentTypeDial, conn.msgMux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeStateClosed, conn.msgMux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeTraceRoute, conn.msgMux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeConnInspectRequest, conn.msgMux.HandleReceive)
	binding.AddReceiveHandlerF(edge.ContentTypeBindSuccess, conn.msgMux.HandleReceive)

	// Since data is the common message type, it gets to be dispatched directly
	binding.AddTypedReceiveHandler(conn.msgMux)
	binding.AddCloseHandler(conn.msgMux)
	binding.AddCloseHandler(conn)

	return nil
}

func (conn *routerConn) NewDialConn(service *rest_model.ServiceDetail) *edgeConn {
	id := conn.msgMux.GetNextId()

	edgeCh := &edgeConn{
		MsgChannel:  *edge.NewEdgeMsgChannel(conn.ch, id),
		readQ:       NewNoopSequencer[*channel.Message](4),
		msgMux:      conn.msgMux,
		serviceName: *service.Name,
		connType:    ConnTypeDial,
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

	err = conn.msgMux.AddMsgSink(edgeCh) // duplicate errors only happen on the server side, since client controls ids
	if err != nil {
		pfxlog.Logger().Warnf("error adding message sink %s[%d]: %v", *service.Name, id, err)
	}
	return edgeCh
}

func (conn *routerConn) UpdateToken(token []byte, timeout time.Duration) error {
	msg := edge.NewUpdateTokenMsg(token)
	resp, err := msg.WithTimeout(timeout).SendForReply(conn.ch)

	if err != nil {
		return err
	}

	if resp.ContentType == edge.ContentTypeUpdateTokenSuccess {
		return nil
	}

	if resp.ContentType == edge.ContentTypeUpdateTokenFailure {
		err = errors.New(string(resp.Body))
		return fmt.Errorf("could not update token for router [%s]: %w", conn.Key(), err)
	}

	err = fmt.Errorf("invalid content type response %d, expected one of [%d, %d]", resp.ContentType, edge.ContentTypeUpdateTokenSuccess, edge.ContentTypeUpdateTokenFailure)
	return fmt.Errorf("could not update token for router [%s]: %w", conn.Key(), err)
}

func (conn *routerConn) NewListenConn(service *rest_model.ServiceDetail, keyPair *kx.KeyPair) *edgeConn {
	id := conn.msgMux.GetNextId()

	edgeCh := &edgeConn{
		MsgChannel:  *edge.NewEdgeMsgChannel(conn.ch, id),
		readQ:       NewNoopSequencer[*channel.Message](4),
		msgMux:      conn.msgMux,
		serviceName: *service.Name,
		connType:    ConnTypeBind,
		keyPair:     keyPair,
		crypto:      keyPair != nil,
		hosting:     cmap.New[*edgeListener](),
	}

	// duplicate errors only happen on the server side, since client controls ids
	if err := conn.msgMux.AddMsgSink(edgeCh); err != nil {
		pfxlog.Logger().Warnf("error adding message sink %s[%d]: %v", *service.Name, id, err)
	}
	pfxlog.Logger().WithField("connId", id).
		WithField("routerName", conn.routerName).
		WithField("serviceId", *service.ID).
		WithField("serviceName", *service.Name).
		Debug("created new listener connection")
	return edgeCh
}

func (conn *routerConn) Connect(service *rest_model.ServiceDetail, session *rest_model.SessionDetail, options *edge.DialOptions) (edge.Conn, error) {
	ec := conn.NewDialConn(service)
	dialConn, err := ec.Connect(session, options)
	if err != nil {
		if err2 := ec.Close(); err2 != nil {
			pfxlog.Logger().Errorf("failed to cleanup connection for service '%v' (%v)", service.Name, err2)
		}
	}
	return dialConn, err
}

func (conn *routerConn) Listen(service *rest_model.ServiceDetail, session *rest_model.SessionDetail, options *edge.ListenOptions) (edge.Listener, error) {
	ec := conn.NewListenConn(service, options.KeyPair)

	log := pfxlog.Logger().
		WithField("connId", ec.Id()).
		WithField("router", conn.routerName).
		WithField("serviceId", *service.ID).
		WithField("serviceName", *service.Name)

	listener, err := ec.listen(session, service, options)
	if err != nil {
		log.WithError(err).Error("failed to establish listener")

		if err2 := ec.Close(); err2 != nil {
			log.WithError(err2).
				Error("failed to cleanup listener for service after failed bind")
		}
	} else {
		if !conn.GetBoolHeader(edge.SupportsBindSuccessHeader) {
			listener.established.Store(true)
		}
		log.Debug("established listener")
	}
	return listener, err
}

func (conn *routerConn) Close() error {
	if !conn.ch.IsClosed() {
		return conn.ch.Close()
	}

	return nil
}

func (conn *routerConn) IsClosed() bool {
	return conn.ch.IsClosed()
}
