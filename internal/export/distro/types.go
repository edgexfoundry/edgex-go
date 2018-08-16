//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	export "github.com/edgexfoundry/edgex-go/export"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Sender - Send interface
type Sender interface {
	Send(data []byte, event *models.Event) bool
}

// Formatter - Format interface
type Formatter interface {
	Format(event *models.Event) []byte
}

// Transformer - Transform interface
type Transformer interface {
	Transform(data []byte) []byte
}

// Filter - Filter interface
type Filterer interface {
	Filter(event *models.Event) (bool, *models.Event)
}

// RegistrationInfo - registration info
type registrationInfo struct {
	registration export.Registration
	format       Formatter
	compression  Transformer
	encrypt      Transformer
	sender       Sender
	filter       []Filterer

	chRegistration chan *export.Registration
	chEvent        chan *models.Event

	deleteMe bool
}
