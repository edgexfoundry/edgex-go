//
// Copyright (C) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "maps"

type DiscoveredDevice struct {
	ProfileName string
	AdminState  AdminState
	AutoEvents  []AutoEvent
	Properties  map[string]any
}

func (d DiscoveredDevice) Clone() DiscoveredDevice {
	cloned := DiscoveredDevice{
		ProfileName: d.ProfileName,
		AdminState:  d.AdminState,
	}
	if len(d.AutoEvents) > 0 {
		cloned.AutoEvents = make([]AutoEvent, len(d.AutoEvents))
		copy(cloned.AutoEvents, d.AutoEvents)
	}
	if len(d.Properties) > 0 {
		cloned.Properties = make(map[string]any)
		maps.Copy(cloned.Properties, d.Properties)
	}
	return cloned
}
