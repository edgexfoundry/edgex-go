package io

import (
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	dto "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestDeviceName         = "TestDevice"
	TestDeviceResourceName = "TestDeviceResourceName"
	TestDeviceProfileName  = "TestDeviceProfileName"
	TestReadingValue       = uint8(45)
)

func buildTestAddEvent() dto.AddEventRequest {
	testEvent := dtos.NewEvent(TestDeviceProfileName, TestDeviceName, TestDeviceResourceName)
	testEvent.AddSimpleReading(TestDeviceResourceName, common.ValueTypeUint8, TestReadingValue)
	return dto.NewAddEventRequest(testEvent)
}

func TestNewEventRequestReader(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		expectedType interface{}
	}{
		{"Get Json Reader", common.ContentTypeJSON, jsonEventReader{}},
		{"Get Cbor Reader", common.ContentTypeCBOR, cborEventReader{}},
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
	var testAddEvent = buildTestAddEvent()
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
			jsonReader := NewEventRequestReader(common.ContentTypeJSON)
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
	var testAddEvent = buildTestAddEvent()
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
			cborReader := NewEventRequestReader(common.ContentTypeCBOR)
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
