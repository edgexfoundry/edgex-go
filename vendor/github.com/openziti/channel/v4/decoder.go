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
	"github.com/michaelquigley/pfxlog"
)

type Decoder struct{}

const DECODER = "channel"

func (d Decoder) Decode(msg *Message) ([]byte, bool) {
	switch msg.ContentType {
	case ContentTypeHelloType:
		hello := UnmarshalHello(msg)
		meta := NewTraceMessageDecode(DECODER, "Hello")
		meta["id"] = hello.IdToken
		meta["headers"] = hello.Headers

		data, err := meta.MarshalTraceMessageDecode()
		if err != nil {
			pfxlog.Logger().Errorf("unexpected error (%s)", err)
			return nil, true
		}

		return data, true
	case ContentTypePingType:
		data, err := NewTraceMessageDecode(DECODER, "Ping").MarshalTraceMessageDecode()
		if err != nil {
			pfxlog.Logger().Errorf("unexpected error (%s)", err)
			return nil, true
		}

		return data, true

	case ContentTypeResultType:
		result := UnmarshalResult(msg)
		meta := NewTraceMessageDecode(DECODER, "Result")
		meta["success"] = result.Success
		if value, found := meta["message"]; found && value != "" {
			meta["message"] = result.Message
		}

		data, err := meta.MarshalTraceMessageDecode()
		if err != nil {
			pfxlog.Logger().Errorf("unexpected error (%s)", err)
			return nil, true
		}

		return data, true

	case ContentTypeLatencyType:
		meta := NewTraceMessageDecode(DECODER, "Latency")
		meta["probeTime"], _ = msg.GetUint64Header(128)
		data, err := meta.MarshalTraceMessageDecode()
		if err != nil {
			pfxlog.Logger().Errorf("unexpected error (%s)", err)
			return nil, true
		}

		return data, true

	case ContentTypeLatencyResponseType:
		meta := NewTraceMessageDecode(DECODER, "LatencyResponse")
		meta["probeTime"], _ = msg.GetUint64Header(128)
		data, err := meta.MarshalTraceMessageDecode()
		if err != nil {
			pfxlog.Logger().Errorf("unexpected error (%s)", err)
			return nil, true
		}
		return data, true

	case ContentTypeHeartbeat:
		meta := NewTraceMessageDecode(DECODER, "Heartbeat")
		data, err := meta.MarshalTraceMessageDecode()
		if err != nil {
			pfxlog.Logger().Errorf("unexpected error (%s)", err)
			return nil, true
		}
		return data, true
	}

	return nil, false
}
