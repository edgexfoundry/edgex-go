//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package export

// Filter - Specifies the client filters on reading data
type Filter struct {
	DeviceIDs          []string `bson:"deviceIdentifiers,omitempty" json:"deviceIdentifiers,omitempty"`
	ValueDescriptorIDs []string `bson:"valueDescriptorIdentifiers,omitempty" json:"valueDescriptorIdentifiers,omitempty"`
}
