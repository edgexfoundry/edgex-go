//
// Copyright (c) 2018
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import "github.com/edgexfoundry/edgex-go/pkg/models"

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
