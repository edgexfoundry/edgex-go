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
type sender interface {
	Send(data []byte, event *models.Event) bool
}

// Formatter - Format interface
type formatter interface {
	Format(event *models.Event) []byte
}

// Transformer - Transform interface
type transformer interface {
	Transform(data []byte) []byte
}

// Filter - Filter interface
type filterer interface {
	Filter(event *models.Event) (bool, *models.Event)
}
