//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package export

// Protocols
const (
	ProtoHTTP  = "HTTP"
	ProtoTCP   = "TCP"
	ProtoMAC   = "MAC"
	ProtoZMQ   = "ZMQ"
	ProtoOther = "OTHER"
)

// Methods
const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodPatch  = "PATCH"
	MethodDelete = "DELETE"
)

// Addressable - address for reaching the service
type Addressable struct {
	Name      string
	Method    string
	Protocol  string
	Address   string
	Port      int
	Path      string
	Publisher string
	User      string
	Password  string
	Topic     string
}
