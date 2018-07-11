//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type jsonFormatter struct {
}

// Azure IoT message feedback codes
type feedbackCode int

const (
	none feedbackCode = iota
	positive
	negative
	full
)

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

// Azure IoT Hub message
// https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-devguide-messages-construct
type connAuthMethod struct {
	Scope  string `json:"scope"`
	Type   string `json:"type"`
	Issuer string `json:"issuer"`
}

// AzureMessage represents Azure IoT Hub message.
type AzureMessage struct {
	ID             string            `json:"id"`
	SequenceNumber int64             `json:"sequenceNumber"`
	To             string            `json:"To"`
	Created        time.Time         `json:"CreationTimeUtc"`
	Expire         time.Time         `json:"ExpiryTimeUtc"`
	Enqueued       time.Time         `json:"EnqueuedTime"`
	CorrelationID  string            `json:"CorrelationId"`
	UserID         string            `json:"userId"`
	Ack            feedbackCode      `json:"ack"`
	ConnDevID      string            `json:"connectionDeviceId"`
	ConnDevGenID   string            `json:"connectionDeviceGenerationId"`
	ConnAuthMethod connAuthMethod    `json:"connectionAuthMethod,omitempty"`
	Body           []byte            `json:"body"`
	Properties     map[string]string `json:"properties"`
}

// NewAzureMessage creates a new Azure message and sets
// Body and default fields values.
func NewAzureMessage() (*AzureMessage, error) {
	msg := &AzureMessage{
		Ack:        none,
		Properties: make(map[string]string),
		Created:    time.Now(),
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	msg.ID = id.String()

	correlationID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	msg.CorrelationID = correlationID.String()

	return msg, nil
}

// AddProperty method ads property performing key check.
func (am *AzureMessage) AddProperty(key, value string) error {
	am.Properties[key] = value
	return nil
}

// azureFormatter is used to convert Event to Azure message and
// Azure message to bytes.
type azureFormatter struct {
}

// Format method does all foramtting job.
func (af azureFormatter) Format(event *models.Event) []byte {
	am, err := NewAzureMessage()
	if err != nil {
		logger.Error(fmt.Sprintf("error creating a new Azure message: %s", err))
		return []byte{}
	}
	am.ConnDevID = event.Device
	am.UserID = string(event.Origin)
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error(fmt.Sprintf("error parsing Event data: %s", err))
		return []byte{}
	}
	am.Body = data
	msg, err := json.Marshal(am)
	if err != nil {
		logger.Error(fmt.Sprintf("error parsing AzureMessage data: %s", err))
		return []byte{}
	}
	return msg
}

type noopFormatter struct {
}

func (noopFmt noopFormatter) Format(event *models.Event) []byte {
	return []byte{}
}
