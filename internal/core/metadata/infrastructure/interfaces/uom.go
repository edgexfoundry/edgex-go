//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

// UnitsOfMeasure defines required functionality to perform units of measure
// validation in EdgeX
type UnitsOfMeasure interface {
	// Validate validates DeviceResource's unit against the list of
	// units of measure by core metadata.
	Validate(string) bool
}
