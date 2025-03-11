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
	"os"
)

type messageHandler interface {
	Handle(msg interface{}) error
}

func Read(path string, handler messageHandler) error {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}

	for {
		typeData := make([]byte, 4)
		n, err := f.Read(typeData)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if n != 4 {
			return errors.New("short read")
		}
		buf := bytes.NewBuffer(typeData)
		var messageType int32
		err = binary.Read(buf, binary.LittleEndian, &messageType)
		if err != nil {
			return err
		}

		lenData := make([]byte, 4)
		n, err = f.Read(lenData)
		if err != nil {
			return err
		}
		if n != 4 {
			return errors.New("short read")
		}
		buf = bytes.NewBuffer(lenData)
		var len int32 = -1
		err = binary.Read(buf, binary.LittleEndian, &len)
		if err != nil {
			return err
		}

		data := make([]byte, len)
		n, err = f.Read(data)
		if err != nil {
			return err
		}
		if int32(n) != len {
			return errors.New("short read")
		}

		switch messageType {
		case int32(trace_pb.MessageType_ChannelStateType):
			cs := &trace_pb.ChannelState{}
			if err = proto.Unmarshal(data, cs); err != nil {
				return err
			}
			if err = handler.Handle(cs); err != nil {
				return err
			}

		case int32(trace_pb.MessageType_ChannelMessageType):
			cm := &trace_pb.ChannelMessage{}
			if err = proto.Unmarshal(data, cm); err != nil {
				return err
			}
			if err = handler.Handle(cm); err != nil {
				return err
			}

		default:
			return errors.New("unexpected message")
		}
	}
}
