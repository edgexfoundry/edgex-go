package network

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v4"
	"github.com/openziti/sdk-golang/edgexg"
	"github.com/openziti/sdk-golang/xgress"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/sirupsen/logrus"
)

type XgAdapter struct {
	conn         *edgeConn
	readC        chan []byte
	env          xgress.Env
	xg           *xgress.Xgress
	writeAdapter *xgress.WriteAdapter
}

func (self *XgAdapter) HandleXgressClose(x *xgress.Xgress) {
	xgCloseMsg := channel.NewMessage(edge.ContentTypeXgClose, []byte(self.xg.CircuitId()))
	if err := xgCloseMsg.WithTimeout(5 * time.Second).Send(self.conn.GetControlSender()); err != nil {
		pfxlog.Logger().WithError(err).Error("failed to send close xg close message")
	}

	// see note in close
	self.conn.msgMux.Remove(self.conn)
}

func (self *XgAdapter) ForwardPayload(payload *xgress.Payload, _ *xgress.Xgress, ctx context.Context) {
	msg := payload.Marshall()
	msg.PutUint32Header(edge.ConnIdHeader, self.conn.Id())

	if err := msg.WithContext(ctx).SendAndWaitForWire(self.conn.GetDefaultSender()); err != nil {
		pfxlog.Logger().WithField("circuitId", payload.CircuitId).WithError(err).Error("failed to send payload")
	}
}

func (self *XgAdapter) RetransmitPayload(srcAddr xgress.Address, payload *xgress.Payload) error {
	msg := payload.Marshall()
	sent, err := self.conn.MsgChannel.GetDefaultSender().TrySend(msg)
	if err != nil {
		// if the channel is closed, close the xgress
		if self.conn.MsgChannel.GetChannel().IsClosed() {
			self.xg.Close()
		}
		return err
	}

	if !sent {
		pfxlog.Logger().WithField("circuitId", payload.CircuitId).WithError(err).Debug("payload dropped")
	}

	return nil
}

func (self *XgAdapter) ForwardControlMessage(control *xgress.Control, x *xgress.Xgress) {
	msg := control.Marshall()
	if err := self.conn.MsgChannel.GetDefaultSender().Send(msg); err != nil {
		pfxlog.Logger().WithError(err).Error("failed to forward control message")
	}
}

func (self *XgAdapter) ForwardAcknowledgement(ack *xgress.Acknowledgement, address xgress.Address) {
	msg := ack.Marshall()
	if err := self.conn.MsgChannel.GetDefaultSender().Send(msg); err != nil {
		pfxlog.Logger().WithError(err).Error("failed to send acknowledgement")
	}
}

func (self *XgAdapter) GetRetransmitter() *xgress.Retransmitter {
	return self.env.GetRetransmitter()
}

func (self *XgAdapter) GetPayloadIngester() *xgress.PayloadIngester {
	return self.env.GetPayloadIngester()
}

func (self *XgAdapter) GetMetrics() xgress.Metrics {
	return self.env.GetMetrics()
}

func (self *XgAdapter) Close() error {
	return nil
}

func (self *XgAdapter) LogContext() string {
	return fmt.Sprintf("xg/%s", self.conn.GetCircuitId())
}

func (self *XgAdapter) ReadPayload() ([]byte, map[uint8][]byte, error) {
	return nil, nil, errors.New("should never be called")
}

func (self *XgAdapter) WritePayload(bytes []byte, headers map[uint8][]byte) (int, error) {
	var msgUUID []byte
	var edgeHdrs map[int32][]byte

	if headers != nil {
		msgUUID = headers[xgress.HeaderKeyUUID]

		edgeHdrs = make(map[int32][]byte)
		for k, v := range headers {
			if edgeHeader, found := edgexg.HeadersFromFabric[k]; found {
				edgeHdrs[edgeHeader] = v
			}
		}
	}

	msg := edge.NewDataMsg(self.conn.Id(), bytes)
	if msgUUID != nil {
		msg.Headers[edge.UUIDHeader] = msgUUID
	}

	for k, v := range edgeHdrs {
		msg.Headers[k] = v
	}

	if err := self.conn.readQ.PutSequenced(msg); err != nil {
		logrus.WithFields(edge.GetLoggerFields(msg)).WithError(err).
			WithField("circuitId", self.conn.circuitId).
			Error("error pushing edge message to sequencer")
		return 0, err
	}

	logrus.WithFields(edge.GetLoggerFields(msg)).Debugf("received %v bytes (msg type: %v)", len(msg.Body), msg.ContentType)
	return len(msg.Body), nil
}

func (self *XgAdapter) FlowFromFabricToXgressClosed() {
	pfxlog.Logger().WithField("circuitId", self.conn.circuitId).
		Debug("fabric to sdk flow complete")
	self.conn.readQ.Close()
}

func (self *XgAdapter) HandleControlMsg(controlType xgress.ControlType, headers channel.Headers, responder xgress.ControlReceiver) error {
	//TODO implement me
	panic("implement me")
}
