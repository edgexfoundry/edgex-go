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
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

type TraceMessageDecoder interface {
	Decode(msg *Message) ([]byte, bool)
}

type TraceMessageDecode map[string]interface{}

const DecoderFieldName = "__decoder__"
const MessageFieldName = "__message__"

func NewTraceMessageDecode(decoder, message string) TraceMessageDecode {
	meta := make(map[string]interface{})
	meta[DecoderFieldName] = decoder
	meta[MessageFieldName] = message
	return meta
}

func (d TraceMessageDecode) MarshalTraceMessageDecode() ([]byte, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d TraceMessageDecode) MarshalResult() ([]byte, bool) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, true
	}
	return data, true
}

func decodeTraceAndFormat(decode []byte) string {
	if len(decode) > 0 {
		meta := make(map[string]interface{})
		err := json.Unmarshal(decode, &meta)
		if err != nil {
			panic(err)
		}

		out := fmt.Sprintf("%-24s", fmt.Sprintf("%-8s %s", meta[DecoderFieldName], meta[MessageFieldName]))

		if len(meta) > 2 {
			keys := make([]string, 0)
			for k := range meta {
				if k != DecoderFieldName && k != MessageFieldName {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)

			out += " {"
			for i := 0; i < len(keys); i++ {
				k := keys[i]
				if i > 0 {
					out += " "
				}
				out += k
				out += "=["
				v := meta[k]
				switch v := v.(type) {
				case string:
					out += v
				case float64:
					out += fmt.Sprintf("%0.0f", v)
				case bool:
					out += fmt.Sprintf("%t", v)
				default:
					out += fmt.Sprintf("<%s>", reflect.TypeOf(v))
				}
				out += "]"
			}
			out += "}"
		}

		return out
	} else {
		return ""
	}
}
