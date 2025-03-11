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
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3/trace"
	"github.com/openziti/channel/v3/trace/pb"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

type TraceHandler struct {
	f         *os.File
	id        string
	decoders  []TraceMessageDecoder
	logTraces bool
}

func NewTraceHandler(path string, id string) (*TraceHandler, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &TraceHandler{
		f:        f,
		id:       id,
		decoders: make([]TraceMessageDecoder, 0),
	}, nil
}

func (h *TraceHandler) AddDecoder(decoder TraceMessageDecoder) {
	h.decoders = append(h.decoders, decoder)
}

func (h *TraceHandler) Connect(ch Channel, remoteAddress string) {
	if err := trace.WriteChannelState(&trace_pb.ChannelState{
		Timestamp:     time.Now().UnixNano(),
		Identity:      h.id,
		Channel:       ch.Label(),
		RemoteAddress: remoteAddress,
		Connected:     true,
	}, h.f); err != nil {
		pfxlog.ContextLogger(ch.Label()).Errorf("unexpected error (%s)", err)
	}
}

func (h TraceHandler) Rx(msg *Message, ch Channel) {
	h.writeChannelMessage(msg, ch, true)
}

func (h TraceHandler) Tx(msg *Message, ch Channel) {
	h.writeChannelMessage(msg, ch, false)
}

func (h TraceHandler) Close(ch Channel) {
	if err := trace.WriteChannelState(&trace_pb.ChannelState{
		Timestamp: time.Now().UnixNano(),
		Identity:  h.id,
		Channel:   ch.Label(),
		Connected: false,
	}, h.f); err != nil {
		pfxlog.ContextLogger(ch.Label()).Errorf("unexpected error (%s)", err)
	}
}

func (h TraceHandler) writeChannelMessage(msg *Message, ch Channel, rx bool) {
	var decode []byte
	for _, decoder := range h.decoders {
		if str, ok := decoder.Decode(msg); ok {
			decode = str
			break
		}
	}

	t := &trace_pb.ChannelMessage{
		Timestamp:   time.Now().UnixNano(),
		Identity:    h.id,
		Channel:     ch.Label(),
		IsRx:        rx,
		ContentType: msg.ContentType,
		Sequence:    msg.sequence,
		ReplyFor:    msg.ReplyFor(),
		Length:      int32(len(msg.Body)),
		Decode:      decode,
	}

	if err := trace.WriteChannelMessage(t, h.f); err != nil {
		pfxlog.ContextLogger(ch.Label()).Errorf("unexpected error (%s)", err)
	}

	if h.logTraces {
		logrus.Info(h.msgToString(t))
	}
}

func (h TraceHandler) msgToString(msg *trace_pb.ChannelMessage) string {
	flow := "->"
	if msg.IsRx {
		flow = "<-"
	}
	replyFor := ""
	if msg.ReplyFor != -1 {
		replyFor = fmt.Sprintf(">%d", msg.ReplyFor)
	}
	return fmt.Sprintf("%-16s %8s %s #%-5d %5s | %s\n", msg.Identity, msg.Channel, flow, msg.Sequence, replyFor, decodeTraceAndFormat(msg.Decode))
}
