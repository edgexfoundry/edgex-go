//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package data

// BaseReading contains data that was gathered from a device.
// Readings returned will all inherit from BaseReading but their concrete types will be either SimpleReading or BinaryReading,
// potentially interleaved in the APIv2 specification.
type BaseReading struct {
	Id          string   `json:"id,omitempty" codec:"id,omitempty"`
	Pushed      int64    `json:"pushed,omitempty" codec:"pushed,omitempty"`   // When the data was pushed out of EdgeX (0 - not pushed yet)
	Created     int64    `json:"created,omitempty" codec:"created,omitempty"` // When the reading was created
	Origin      int64    `json:"origin,omitempty" codec:"origin,omitempty"`
	Modified    int64    `json:"modified,omitempty" codec:"modified,omitempty"`
	Device      string   `json:"device,omitempty" codec:"device,omitempty"`
	Name        string   `json:"name,omitempty" codec:"name,omitempty"`
	Labels      []string `json:"labels,omitempty" codec:"labels,omitempty"` // Custom labels assigned to a reading, added in the APIv2 specification.
	isValidated bool     // internal member used for validation check
}

// An event reading for a binary data type
// BinaryReading object in the APIv2 specification.
type BinaryReading struct {
	BaseReading `json:",inline"`
	BinaryValue []byte `json:"binaryValue,omitempty" codec:"binaryValue,omitempty"` // Binary data payload
	MediaType   string `json:"mediaType,omitempty" codec:"mediaType,omitempty"`     // indicates what the content type of the binaryValue property is
}

// An event reading for a simple data type
// SimpleReading object in the APIv2 specification.
type SimpleReading struct {
	BaseReading   `json:",inline"`
	Value         string `json:"value,omitempty" codec:"value,omitempty"`                 // Device sensor data value
	ValueType     string `json:"valueType,omitempty" codec:"valueType,omitempty"`         // Indicates the datatype of the value property
	FloatEncoding string `json:"floatEncoding,omitempty" codec:"floatEncoding,omitempty"` // Indicates how a float value is encoded
}

// a abstract interface to be implemented by BinaryReading/SimpleReading
type Reading interface {
	implicit()
}

// Empty methods for BinaryReading and SimpleReading structs to implement the abstract Reading interface
func (BinaryReading) implicit() {}
func (SimpleReading) implicit() {}
