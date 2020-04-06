//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package data

// Event represents a single measurable event read from a device
type Event struct {
	ID          string    `json:"id,omitempty" codec:"id,omitempty"`             // ID uniquely identifies an event, for example a UUID
	Pushed      int64     `json:"pushed,omitempty" codec:"pushed,omitempty"`     // Pushed is a timestamp indicating when the event was exported. If unexported, the value is zero.
	Device      string    `json:"device,omitempty" codec:"device,omitempty"`     // Device identifies the source of the event, can be a device name or id. Usually the device name.
	Created     int64     `json:"created,omitempty" codec:"created,omitempty"`   // Created is a timestamp indicating when the event was created.
	Modified    int64     `json:"modified,omitempty" codec:"modified,omitempty"` // Modified is a timestamp indicating when the event was last modified.
	Origin      int64     `json:"origin,omitempty" codec:"origin,omitempty"`     // Origin is a timestamp that can communicate the time of the original reading, prior to event creation
	Readings    []Reading `json:"readings,omitempty" codec:"readings,omitempty"` // Readings will contain zero to many entries for the associated readings of a given event.
	isValidated bool      // internal member used for validation check
}
