/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package models

import (
	"encoding/json"
)

/*
 * This file is for the Reading model in EdgeX
 * Holds data that was gathered from a device
 */

// ValueType indicates the type of value being passed back
// from a ProtocolDriver instance.
type ValueType int

const (
        // No type was defined (default value)
        NoType  ValueType = iota
	// Bool indicates that the value is a bool,
	// stored in CommandValue's boolRes member.
	Bool
	// String indicates that the value is a string,
	// stored in CommandValue's stringRes member.
	String
	// Uint8 indicates that the value is a uint8 that
	// is stored in CommandValue's NumericRes member.
	Uint8
	// Uint16 indicates that the value is a uint16 that
	// is stored in CommandValue's NumericRes member.
	Uint16
	// Uint32 indicates that the value is a uint32 that
	// is stored in CommandValue's NumericRes member.
	Uint32
	// Uint64 indicates that the value is a uint64 that
	// is stored in CommandValue's NumericRes member.
	Uint64
	// Int8 indicates that the value is a int8 that
	// is stored in CommandValue's NumericRes member.
	Int8
	// Int16 indicates that the value is a int16 that
	// is stored in CommandValue's NumericRes member.
	Int16
	// Int32 indicates that the value is a int32 that
	// is stored in CommandValue's NumericRes member.
	Int32
	// Int64 indicates that the value is a int64 that
	// is stored in CommandValue's NumericRes member.
	Int64
	// Float32 indicates that the value is a float32 that
	// is stored in CommandValue's NumericRes member.
	Float32
	// Float64 indicates that the value is a float64 that
	// is stored in CommandValue's NumericRes member.
	Float64
)

/*
 * Struct for the Reading object in EdgeX
 */
type Reading struct {
	Id       string        `json:"id"`
	Pushed   int64         `json:"pushed"`  // When the data was pushed out of EdgeX (0 - not pushed yet)
	Created  int64         `json:"created"` // When the reading was created
	Origin   int64         `json:"origin"`
	Modified int64         `json:"modified"`
	Device   string        `json:"device"`
	Name     string        `json:"name"`
	Value    string        `json:"value"` // Device sensor data value
	Unit     string        `bson:"unit,omitempty" json:"unit"`
	Type     ValueType     `bson:"type,omitempty" json:"type"`
}

// Custom marshaling to make empty strings null
func (r Reading) MarshalJSON() ([]byte, error) {
	test := struct {
		Id       *string       `json:"id,omitempty"`
		Pushed   int64         `json:"pushed,omitempty"`  // When the data was pushed out of EdgeX (0 - not pushed yet)
		Created  int64         `json:"created,omitempty"` // When the reading was created
		Origin   int64         `json:"origin,omitempty"`
		Modified int64         `json:"modified,omitempty"`
		Device   *string       `json:"device,omitempty"`
		Name     *string       `json:"name,omitempty"`
		Value    *string       `json:"value,omitempty"` // Device sensor data value
		Unit     *string       `json:"unit,omitempty"`
		Type     *ValueType    `json:"type,omitempty"`
	}{
		Pushed:   r.Pushed,
		Created:  r.Created,
		Origin:   r.Origin,
		Modified: r.Modified,
	}

	// Empty strings are null
	if r.Id != "" {
		test.Id = &r.Id
	}
	if r.Device != "" {
		test.Device = &r.Device
	}
	if r.Name != "" {
		test.Name = &r.Name
	}
	if r.Value != "" {
		test.Value = &r.Value
	}
	if r.Unit != "" {
		test.Unit = &r.Unit
	}
        if r.Type != 0 {
	        test.Type = &r.Type

        }

	return json.Marshal(test)
}

/*
 * To String function for Reading Struct
 */
func (r Reading) String() string {
	out, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

