package io

import (
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	dto "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ExampleUUID            = "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	TestDeviceName         = "TestDevice"
	TestOriginTime         = 1600666185705354000
	TestDeviceResourceName = "TestDeviceResourceName"
	TestDeviceProfileName  = "TestDeviceProfileName"
	TestReadingValue       = "45"
)

var expectedEventId = uuid.New().String()

var testReading = dtos.BaseReading{
	Versionable:  common.NewVersionable(),
	DeviceName:   TestDeviceName,
	ResourceName: TestDeviceResourceName,
	ProfileName:  TestDeviceProfileName,
	Origin:       TestOriginTime,
	ValueType:    v2.ValueTypeUint8,
	SimpleReading: dtos.SimpleReading{
		Value: TestReadingValue,
	},
}

var testAddEvent = dto.AddEventRequest{
	BaseRequest: common.BaseRequest{
		Versionable: common.NewVersionable(),
		RequestId:   ExampleUUID,
	},
	Event: dtos.Event{
		Versionable: common.NewVersionable(),
		Id:          expectedEventId,
		DeviceName:  TestDeviceName,
		ProfileName: TestDeviceProfileName,
		SourceName:  TestDeviceResourceName,
		Origin:      TestOriginTime,
		Readings:    []dtos.BaseReading{testReading},
	},
}

func TestNewEventRequestReader(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		expectedType interface{}
	}{
		{"Get Json Reader", clients.ContentTypeJSON, jsonEventReader{}},
		{"Get Cbor Reader", clients.ContentTypeCBOR, cborEventReader{}},
		{"Get Json Reader when content-type is unknown", "Unknown-Type", jsonEventReader{}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reader := NewEventRequestReader(testCase.contentType)
			assert.IsType(t, testCase.expectedType, reader, "unexpected reader type")
		})
	}
}

func TestJsonSerialization(t *testing.T) {
	tests := []struct {
		name          string
		targetDTO     interface{}
		errorExpected bool
	}{
		{"Valid", testAddEvent, false},
		{"Invalid", "string1", true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonReader := NewEventRequestReader(clients.ContentTypeJSON)
			byteArray, err := json.Marshal(testCase.targetDTO)
			require.NoError(t, err, "error occurs during json marshalling")
			_, err = jsonReader.ReadAddEventRequest(byteArray)
			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCborSerialization(t *testing.T) {
	tests := []struct {
		name          string
		targetDTO     interface{}
		errorExpected bool
	}{
		{"Valid", testAddEvent, false},
		{"Invalid", "string1", true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cborReader := NewEventRequestReader(clients.ContentTypeCBOR)
			byteArray, err := cbor.Marshal(testCase.targetDTO)
			require.NoError(t, err, "error occurs during cbor marshalling")
			_, err = cborReader.ReadAddEventRequest(byteArray)
			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
