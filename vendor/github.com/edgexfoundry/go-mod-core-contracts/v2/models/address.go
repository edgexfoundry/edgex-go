//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type Address interface {
	GetBaseAddress() BaseAddress
}

// instantiateAddress instantiate the interface to the corresponding address type
func instantiateAddress(i interface{}) (address Address, err error) {
	a, err := json.Marshal(i)
	if err != nil {
		return address, errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to marshal address.", err)
	}
	return unmarshalAddress(a)
}

func unmarshalAddress(b []byte) (address Address, err error) {
	var alias struct {
		Type string
	}
	if err = json.Unmarshal(b, &alias); err != nil {
		return address, errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal address.", err)
	}
	switch alias.Type {
	case common.REST:
		var rest RESTAddress
		if err = json.Unmarshal(b, &rest); err != nil {
			return address, errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal REST address.", err)
		}
		address = rest
	case common.MQTT:
		var mqtt MQTTPubAddress
		if err = json.Unmarshal(b, &mqtt); err != nil {
			return address, errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal MQTT address.", err)
		}
		address = mqtt
	case common.EMAIL:
		var mail EmailAddress
		if err = json.Unmarshal(b, &mail); err != nil {
			return address, errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal Email address.", err)
		}
		address = mail
	default:
		return address, errors.NewCommonEdgeX(errors.KindContractInvalid, "Unsupported address type", err)
	}
	return address, nil
}

// BaseAddress is a base struct contains the common fields, such as type, host, port, and so on.
type BaseAddress struct {
	// Type is used to identify the Address type, i.e., REST or MQTT
	Type string

	// Common properties
	Host string
	Port int
}

// RESTAddress is a REST specific struct
type RESTAddress struct {
	BaseAddress
	Path       string
	HTTPMethod string
}

func (a RESTAddress) GetBaseAddress() BaseAddress { return a.BaseAddress }

// MQTTPubAddress is a MQTT specific struct
type MQTTPubAddress struct {
	BaseAddress
	Publisher      string
	Topic          string
	QoS            int
	KeepAlive      int
	Retained       bool
	AutoReconnect  bool
	ConnectTimeout int
}

func (a MQTTPubAddress) GetBaseAddress() BaseAddress { return a.BaseAddress }

// EmailAddress is an Email specific struct
type EmailAddress struct {
	BaseAddress
	Recipients []string
}

func (a EmailAddress) GetBaseAddress() BaseAddress { return a.BaseAddress }
