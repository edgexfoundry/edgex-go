//
// Copyright (c) 2017
// Cavium
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"encoding/json"
	"encoding/xml"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"go.uber.org/zap"
)

type jsonFormatter struct {
}

func (jsonTr jsonFormatter) Format(event *models.Event) []byte {

	b, err := json.Marshal(event)
	if err != nil {
		logger.Error("Error parsing JSON", zap.Error(err))
		return nil
	}
	return b
}

type xmlFormatter struct {
}

func (xmlTr xmlFormatter) Format(event *models.Event) []byte {
	b, err := xml.Marshal(event)
	if err != nil {
		logger.Error("Error parsing XML", zap.Error(err))
		return nil
	}
	return b
}

type thingsboardJSONFormatter struct {
}

// ThingsBoard JSON formatter
// https://thingsboard.io/docs/reference/gateway-mqtt-api/#telemetry-upload-api
func (thingsboardjsonTr thingsboardJSONFormatter) Format(event *models.Event) []byte {

	type Device struct {
		Ts     int64             `json:"ts"`
		Values map[string]string `json:"values"`
	}

	values := make(map[string]string)
	for _, reading := range event.Readings {
		values[reading.Name] = reading.Value
	}

	var deviceValues []Device
	deviceValues = append(deviceValues, Device{Ts: event.Origin, Values: values})

	device := make(map[string][]Device)
	device[event.Device] = deviceValues

	b, err := json.Marshal(device)
	if err != nil {
		logger.Error("Error parsing ThingsBoard JSON", zap.Error(err))
		return nil
	}
	return b
}
