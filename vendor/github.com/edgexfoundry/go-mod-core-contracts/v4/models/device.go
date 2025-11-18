//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "maps"

type Device struct {
	DBTimestamp
	Id             string
	Name           string
	Parent         string
	Description    string
	AdminState     AdminState
	OperatingState OperatingState
	Protocols      map[string]ProtocolProperties
	Labels         []string
	Location       interface{}
	ServiceName    string
	ProfileName    string
	AutoEvents     []AutoEvent
	Tags           map[string]any
	Properties     map[string]any
}

// ProtocolProperties contains the device connection information in key/value pair
type ProtocolProperties map[string]any

func (p ProtocolProperties) Clone() ProtocolProperties {
	cloned := make(map[string]any)
	maps.Copy(cloned, p)
	return cloned
}

// AdminState controls the range of values which constitute valid administrative states for a device
type AdminState string

// AssignAdminState provides a default value "UNLOCKED" to AdminState
func AssignAdminState(dtoAdminState string) AdminState {
	if dtoAdminState == "" {
		return AdminState(Unlocked)
	}
	return AdminState(dtoAdminState)
}

// OperatingState is an indication of the operations of the device.
type OperatingState string

func (device Device) Clone() Device {
	cloned := Device{
		DBTimestamp:    device.DBTimestamp,
		Id:             device.Id,
		Name:           device.Name,
		Parent:         device.Parent,
		Description:    device.Description,
		AdminState:     device.AdminState,
		OperatingState: device.OperatingState,
		Location:       device.Location,
		ServiceName:    device.ServiceName,
		ProfileName:    device.ProfileName,
	}
	if len(device.Protocols) > 0 {
		cloned.Protocols = make(map[string]ProtocolProperties)
		for k, v := range device.Protocols {
			cloned.Protocols[k] = v.Clone()
		}
	}
	if len(device.Labels) > 0 {
		cloned.Labels = make([]string, len(device.Labels))
		copy(cloned.Labels, device.Labels)
	}
	if len(device.AutoEvents) > 0 {
		cloned.AutoEvents = make([]AutoEvent, len(device.AutoEvents))
		copy(cloned.AutoEvents, device.AutoEvents)
	}
	if len(device.Tags) > 0 {
		cloned.Tags = make(map[string]any)
		maps.Copy(cloned.Tags, device.Tags)
	}
	if len(device.Properties) > 0 {
		cloned.Properties = make(map[string]any)
		maps.Copy(cloned.Properties, device.Properties)
	}
	return cloned
}
