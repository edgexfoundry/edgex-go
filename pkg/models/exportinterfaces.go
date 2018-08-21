//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package models

// Sender - Send interface
type Sender interface {
	Send(data []byte, event *Event) bool
}

// Formatter - Format interface
type Formatter interface {
	Format(event *Event) []byte
}

// Transformer - Transform interface
type Transformer interface {
	Transform(data []byte) []byte
}

// Filter - Filter interface
type Filterer interface {
	Filter(event *Event) (bool, *Event)
}
