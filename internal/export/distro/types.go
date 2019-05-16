//
// Copyright (c) 2018
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"context"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Sender - Send interface
type sender interface {
	Send(data []byte, ctx context.Context) bool
}

// Formatter - Format interface
type formatter interface {
	Format(event *contract.Event) []byte
}

// Transformer - Transform interface
type transformer interface {
	Transform(data []byte) []byte
}

// Filter - Filter interface
type filterer interface {
	Filter(event *contract.Event) (bool, *contract.Event)
}
