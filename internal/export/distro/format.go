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
	"strconv"
	"strings"
	"time"

	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/google/uuid"
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

func (jsonTr jsonFormatter) Format(event *contract.Event) []byte {

	b, err := json.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing JSON. Error: %s", err.Error()))
		return nil
	}
	return b
}

type xmlFormatter struct {
}

func (xmlTr xmlFormatter) Format(event *contract.Event) []byte {
	b, err := xml.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing XML. Error: %s", err.Error()))
		return nil
	}
	return b
}

type thingsboardJSONFormatter struct {
}

// ThingsBoard JSON formatter
// https://thingsboard.io/docs/reference/gateway-mqtt-api/#telemetry-upload-api
func (thingsboardjsonTr thingsboardJSONFormatter) Format(event *contract.Event) []byte {

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
		LoggingClient.Error(fmt.Sprintf("Error parsing ThingsBoard JSON. Error: %s", err.Error()))
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

// newAzureMessage creates a new Azure message and sets
// Body and default fields values.
func newAzureMessage() (*AzureMessage, error) {
	msg := &AzureMessage{
		Ack:        none,
		Properties: make(map[string]string),
		Created:    time.Now(),
	}

	id := uuid.New()
	msg.ID = id.String()

	correlationID := uuid.New()
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
func (af azureFormatter) Format(event *contract.Event) []byte {
	am, err := newAzureMessage()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error creating a new Azure message: %s", err))
		return []byte{}
	}
	am.ConnDevID = event.Device
	am.UserID = string(event.Origin)
	data, err := json.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing Event data: %s", err))
		return []byte{}
	}
	am.Body = data
	msg, err := json.Marshal(am)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing AzureMessage data: %s", err))
		return []byte{}
	}
	return msg
}

// converting event to AWS shadow message in bytes
type awsFormatter struct {
}

func (af awsFormatter) Format(event *contract.Event) []byte {
	reported := map[string]interface{}{}

	for _, reading := range event.Readings {
		value, err := strconv.ParseFloat(reading.Value, 64)

		if err != nil {
			strVal := reading.Value
			// not a valid numerical reading value, see if it's boolean
			if strings.Compare(strings.ToLower(strVal), "true") == 0 {
				reported[reading.Name] = true
			} else if strings.Compare(strings.ToLower(strVal), "false") == 0 {
				reported[reading.Name] = false
			} else {
				reported[reading.Name] = strVal
			}

			continue
		}

		reported[reading.Name] = value
	}

	currState := map[string]interface{}{
		"state": map[string]interface{}{
			"reported": reported,
		},
	}

	msg, err := json.Marshal(currState)

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error generating AWS shadow document: %s", err))
		return []byte{}
	}

	return msg
}

type noopFormatter struct {
}

func (noopFmt noopFormatter) Format(event *contract.Event) []byte {
	return []byte{}
}

// BIoTMessage represents Brightics IoT(Samsung SDS IoT platform)  messages.
type BIoTMessage struct {
	Version    string `json:"version"`
	MsgType    string `json:"msgType"`
	FuncType   string `json:"funcType"`
	SId        string `json:"sId"`
	TpId       string `json:"tpId"`
	TId        string `json:"tId"`
	MsgCode    string `json:"msgCode"`
	MsgId      string `json:"msgId"`
	MsgDate    int64  `json:"msgDate"`
	ResCode    string `json:"resCode"`
	ResMsg     string `json:"resMsg"`
	Severity   string `json:"severity"`
	Dataformat string `json:"dataformat"`
	EncType    string `json:"encType"`
	AuthToken  string `json:"authToken"`
	Data       []byte `json:"data"`
}

// newBIoTMessage creates a new Brightics IoT message and sets
// Body and default fields values.
func newBIoTMessage() (*BIoTMessage, error) {
	msg := &BIoTMessage{
		Severity: "1",
		MsgType:  "Q",
	}

	id := uuid.New()
	msg.MsgId = id.String()

	return msg, nil
}

// brighticsiotFormatter is used to convert Event to BIoT message and
// BIoT message to bytes.
type biotFormatter struct {
}

// Format method does all foramtting job.
func (af biotFormatter) Format(event *contract.Event) []byte {
	bm, err := newBIoTMessage()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error creating a new BIoT message: %s", err))
		return []byte{}
	}
	bm.TpId = event.Device
	bm.TId = string(event.Origin)
	rawdata, err := json.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error parsing Event data to BIoTMessage : %s", err))
		return []byte{}
	}
	bm.Data = rawdata
	msg, err := json.Marshal(bm)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error parsing BIoTMessage to data: %s", err))
		return []byte{}
	}
	return msg
}
