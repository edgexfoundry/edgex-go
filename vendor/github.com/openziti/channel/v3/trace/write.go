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

package trace

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/openziti/channel/v3/trace/pb"
	"google.golang.org/protobuf/proto"
	"io"
)

func WriteChannelState(s *trace_pb.ChannelState, writer io.Writer) error {
	data, err := proto.Marshal(s)
	if err != nil {
		return err
	}

	out := new(bytes.Buffer)
	if err = binary.Write(out, binary.LittleEndian, int32(trace_pb.MessageType_ChannelStateType)); err != nil {
		return err
	}
	if err = binary.Write(out, binary.LittleEndian, int32(len(data))); err != nil {
		return err
	}
	out.Write(data)

	n, err := writer.Write(out.Bytes())
	if err != nil {
		return err
	}
	if n != out.Len() {
		return errors.New("short write")
	}

	return nil
}

func WriteChannelMessage(t *trace_pb.ChannelMessage, writer io.Writer) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}

	out := new(bytes.Buffer)
	if err = binary.Write(out, binary.LittleEndian, int32(trace_pb.MessageType_ChannelMessageType)); err != nil {
		return err
	}
	if err = binary.Write(out, binary.LittleEndian, int32(len(data))); err != nil {
		return err
	}
	out.Write(data)

	n, err := writer.Write(out.Bytes())
	if err != nil {
		return err
	}
	if n != out.Len() {
		return errors.New("short write")
	}

	return nil
}
