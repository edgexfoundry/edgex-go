//
// Copyright (c) 2017
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"github.com/edgexfoundry/export-go"
)

const (
	defaultPort       = 48070
	defaultClientHost = "127.0.0.1"
	defaultDataHost   = "127.0.0.1"
)

// Sender - Send interface
type Sender interface {
	Send(data []byte)
}

// Formater - Format interface
type Formater interface {
	Format(event *export.Event) []byte
}

// Transformer - Transform interface
type Transformer interface {
	Transform(data []byte) []byte
}

// Filter - Filter interface
type Filterer interface {
	Filter(event *export.Event) (bool, *export.Event)
}

// RegistrationInfo - registration info
type registrationInfo struct {
	registration export.Registration
	format       Formater
	compression  Transformer
	encrypt      Transformer
	sender       Sender
	filter       []Filterer

	chRegistration chan *export.Registration
	chEvent        chan *export.Event

	deleteMe bool
}

type Config struct {
	Port       int
	ClientHost string
	DataHost   string
}

var cfg Config

func GetDefaultConfig() Config {
	return Config{
		Port:       defaultPort,
		ClientHost: defaultClientHost,
		DataHost:   defaultDataHost,
	}
}
