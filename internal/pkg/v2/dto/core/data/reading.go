//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package data

import dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/dto/common/base"

// Reading type inherited by SimpleReading or BinaryReading.
// BaseReading object in the APIv2 specification.
type BaseReading struct {
	Id       string   `json:"id,omitempty" codec:"id,omitempty"`
	Pushed   int64    `json:"pushed,omitempty" codec:"pushed,omitempty"`   // When the data was pushed out of EdgeX (0 - not pushed yet)
	Created  int64    `json:"created,omitempty" codec:"created,omitempty"` // When the reading was created
	Origin   int64    `json:"origin,omitempty" codec:"origin,omitempty"`
	Modified int64    `json:"modified,omitempty" codec:"modified,omitempty"`
	Device   string   `json:"device,omitempty" codec:"device,omitempty"`
	Name     string   `json:"name,omitempty" codec:"name,omitempty"`
	Labels   []string `json:"labels,omitempty" codec:"labels,omitempty"` // Custom labels assigned to a reading, added in the APIv2 specification.
}

// An event reading for a binary data type, inherit BaseReading
// BinaryReading object in the APIv2 specification.
type BinaryReading struct {
	BaseReading `json:",inline"`
	BinaryValue []byte `json:"binaryValue,omitempty" codec:"binaryValue,omitempty"` // Binary data payload
	MediaType   string `json:"mediaType,omitempty" codec:"mediaType,omitempty"`     // indicates what the content type of the binaryValue property is
}

// An event reading for a simple data type, inherit BaseReading
// SimpleReading object in the APIv2 specification.
type SimpleReading struct {
	BaseReading   `json:",inline"`
	Value         string `json:"value,omitempty" codec:"value,omitempty"`                 // Device sensor data value
	ValueType     string `json:"valueType,omitempty" codec:"valueType,omitempty"`         // Indicates the datatype of the value property
	FloatEncoding string `json:"floatEncoding,omitempty" codec:"floatEncoding,omitempty"` // Indicates how a float value is encoded
}

// a abstract interface to be implemented by BinaryReading or SimpleReading
type Reading interface {
	implicit()
}

// Empty methods for BinaryReading and SimpleReading structs to implement the abstract Reading interface
func (BinaryReading) implicit() {}
func (SimpleReading) implicit() {}

// ReadingCountResponse defines the Response Content for GET reading count DTO.  This object and its properties correspond to the
// ReadingCountResponse object in the APIv2 specification.
type ReadingCountResponse struct {
	dtoBase.Response `json:",inline"`
	Count            int
}

// ReadingResponse defines the Response Content for GET reading DTO.  This object and its properties correspond to the
// ReadingResponse object in the APIv2 specification.
type ReadingResponse struct {
	dtoBase.Response `json:",inline"`
	Reading          Reading
}
