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
	"github.com/pkg/errors"
	"io"
	"time"
)

type MessageStrategy interface {
	GetMarshaller() MessageMarshaller
	GetStreamProducer() StreamMessageProducer
	GetPacketProducer() PacketMessageProducer
}

type MessageMarshaller func(m *Message) ([]byte, error)
type StreamMessageProducer func(r io.Reader) (*Message, error)
type PacketMessageProducer func(b []byte) (*Message, error)

type Options struct {
	OutQueueSize int
	ConnectOptions
	DelayRxStart    bool
	WriteTimeout    time.Duration
	MessageStrategy MessageStrategy
}

func DefaultOptions() *Options {
	return &Options{
		OutQueueSize:   DefaultOutQueueSize,
		ConnectOptions: DefaultConnectOptions(),
	}
}

func DefaultConnectOptions() ConnectOptions {
	return ConnectOptions{
		MaxQueuedConnects:      DefaultQueuedConnects,
		MaxOutstandingConnects: DefaultOutstandingConnects,
		ConnectTimeout:         DefaultConnectTimeout,
	}
}

func LoadOptions(data map[interface{}]interface{}) (*Options, error) {
	options := DefaultOptions()

	if value, found := data["outQueueSize"]; found {
		if floatValue, ok := value.(float64); ok {
			options.OutQueueSize = int(floatValue)
		}
	}

	if value, found := data["maxQueuedConnects"]; found {
		if intVal, ok := value.(int); ok {
			options.MaxQueuedConnects = intVal
		}
	}

	if value, found := data["maxOutstandingConnects"]; found {
		if intVal, ok := value.(int); ok {
			options.MaxOutstandingConnects = intVal
		}
	}

	if value, found := data["connectTimeoutMs"]; found {
		if intVal, ok := value.(int); ok {
			options.ConnectTimeout = time.Duration(intVal) * time.Millisecond
		}
	}

	if value, found := data["writeTimeout"]; found {
		if strVal, ok := value.(string); ok {
			if d, err := time.ParseDuration(strVal); err == nil {
				options.WriteTimeout = d
			} else {
				return nil, errors.Wrapf(err, "invalid value for writeTimeout: %v", value)
			}
		} else {
			return nil, errors.Errorf("invalid (non-string) value for writeTimeout: %v", value)
		}
	}

	return options, nil
}

func (o Options) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type ConnectOptions struct {
	MaxQueuedConnects      int
	MaxOutstandingConnects int
	ConnectTimeout         time.Duration
}

func (co *ConnectOptions) Validate() error {
	if err := co.validateConnectTimeout(); err != nil {
		return err
	}

	if err := co.validateOutstandingConnects(); err != nil {
		return err
	}

	if err := co.validateQueueConnects(); err != nil {
		return err
	}

	return nil
}

func (co *ConnectOptions) validateQueueConnects() error {
	if co.MaxQueuedConnects < MinQueuedConnects {
		return fmt.Errorf("maxQueuedConnects must be at least %d", MinQueuedConnects)
	} else if co.MaxQueuedConnects > MaxQueuedConnects {
		return fmt.Errorf("maxQueuedConnects must be at most %d", MaxQueuedConnects)
	}
	return nil
}

func (co *ConnectOptions) validateOutstandingConnects() error {
	if co.MaxOutstandingConnects < MinOutstandingConnects {
		return fmt.Errorf("maxOutstandingConnects must be at least %d", MinOutstandingConnects)
	} else if co.MaxOutstandingConnects > MaxOutstandingConnects {
		return fmt.Errorf("maxOutstandingConnects must be at most %d", MaxOutstandingConnects)
	}

	return nil
}

func (co *ConnectOptions) validateConnectTimeout() error {
	if co.ConnectTimeout < MinConnectTimeout {
		return fmt.Errorf("connectTimeoutMs must be at least %d ms", MinConnectTimeout.Milliseconds())
	} else if co.ConnectTimeout > MaxConnectTimeout {
		return fmt.Errorf("connectTimeoutMs must be at most %d ms", MaxConnectTimeout.Milliseconds())
	}

	return nil
}
