//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package export

// Event - packet of Readings
type Event struct {
	ID       string    `json:"id,omitempty"`
	Pushed   int64     `json:"pushed"`
	Device   string    `json:"device,omitempty"`
	Readings []Reading `json:"readings,omitempty"`
	Created  int64     `json:"created"`
	Modified int64     `json:"modified"`
	Origin   int64     `json:"origin"`
}

// Reading - Sensor measurement
type Reading struct {
	ID       string `json:"id,omitempty"`
	Pushed   int64  `json:"pushed"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Device   string `json:"device,omitempty"`
	Created  int64  `json:"created"`
	Modified int64  `json:"modified"`
	Origin   int64  `json:"origin"`
}
