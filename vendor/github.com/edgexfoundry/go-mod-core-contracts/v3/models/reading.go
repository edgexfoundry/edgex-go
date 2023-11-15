//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type BaseReading struct {
	Id           string
	Origin       int64
	DeviceName   string
	ResourceName string
	ProfileName  string
	ValueType    string
	Units        string
	Tags         map[string]any
}

type BinaryReading struct {
	BaseReading `json:",inline"`
	BinaryValue []byte
	MediaType   string
}

type SimpleReading struct {
	BaseReading `json:",inline"`
	Value       string
}

type ObjectReading struct {
	BaseReading `json:",inline"`
	ObjectValue interface{}
}

// Reading is an abstract interface to be implemented by BinaryReading/SimpleReading
type Reading interface {
	GetBaseReading() BaseReading
}

// Implement GetBaseReading() method in order for BinaryReading and SimpleReading, ObjectReading structs to implement the
// abstract Reading interface and then be used as a Reading.
// Also, the Reading interface can access the BaseReading fields.
// This is Golang's way to implement inheritance.
func (b BinaryReading) GetBaseReading() BaseReading { return b.BaseReading }
func (s SimpleReading) GetBaseReading() BaseReading { return s.BaseReading }
func (o ObjectReading) GetBaseReading() BaseReading { return o.BaseReading }
