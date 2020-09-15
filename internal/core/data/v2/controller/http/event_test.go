//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"

	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedCorrelationId = uuid.New().String()

var testReading = dtos.BaseReading{
	DeviceName: TestDeviceName,
	Name:       TestDeviceResourceName,
	Origin:     TestOriginTime,
	ValueType:  dtos.ValueTypeUint8,
	SimpleReading: dtos.SimpleReading{
		Value: TestReadingValue,
	},
}

var testAddEvent = requests.AddEventRequest{
	BaseRequest: common.BaseRequest{
		RequestID: ExampleUUID,
	},
	Event: dtos.Event{
		DeviceName: TestDeviceName,
		Origin:     TestOriginTime,
		Readings:   []dtos.BaseReading{testReading},
	},
}

func TestAddEvent(t *testing.T) {
	expectedResponseCode := http.StatusMultiStatus
	expectedRequestId := "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	expectedMessage := "Add events successfully"

	dbClientMock := &dbMock.DBClient{}

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		dataContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					PersistData: false,
				},
			}
		},
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	ec := NewEventController(dic)
	assert.NotNil(t, ec)

	validRequest := testAddEvent

	noRequestId := validRequest
	noRequestId.RequestID = ""
	badRequestId := validRequest
	badRequestId.RequestID = "niv3sl"
	noEvent := validRequest
	noEvent.Event = dtos.Event{}
	noEventDevice := validRequest
	noEventDevice.Event.DeviceName = ""
	noEventOrigin := validRequest
	noEventOrigin.Event.Origin = 0

	noReading := validRequest
	noReading.Event.Readings = []dtos.BaseReading{}
	noReadingDevice := validRequest
	noReadingDevice.Event.Readings = []dtos.BaseReading{testReading}
	noReadingDevice.Event.Readings[0].DeviceName = ""
	noReadingName := validRequest
	noReadingName.Event.Readings = []dtos.BaseReading{testReading}
	noReadingName.Event.Readings[0].Name = ""
	noReadingOrigin := validRequest
	noReadingOrigin.Event.Readings = []dtos.BaseReading{testReading}
	noReadingOrigin.Event.Readings[0].Origin = 0
	noReadingValueType := validRequest
	noReadingValueType.Event.Readings = []dtos.BaseReading{testReading}
	noReadingValueType.Event.Readings[0].ValueType = ""
	invalidReadingInvalidValueType := validRequest
	invalidReadingInvalidValueType.Event.Readings = []dtos.BaseReading{testReading}
	invalidReadingInvalidValueType.Event.Readings[0].ValueType = "BadType"

	noSimpleValue := validRequest
	noSimpleValue.Event.Readings = []dtos.BaseReading{testReading}
	noSimpleValue.Event.Readings[0].Value = ""
	noSimpleFloatEnconding := validRequest
	noSimpleFloatEnconding.Event.Readings = []dtos.BaseReading{testReading}
	noSimpleFloatEnconding.Event.Readings[0].ValueType = dtos.ValueTypeFloat32
	noSimpleFloatEnconding.Event.Readings[0].FloatEncoding = ""
	noBinaryValue := validRequest
	noBinaryValue.Event.Readings = []dtos.BaseReading{{
		DeviceName: TestDeviceName,
		Name:       TestDeviceResourceName,
		Origin:     TestOriginTime,
		ValueType:  dtos.ValueTypeBinary,
		BinaryReading: dtos.BinaryReading{
			BinaryValue: []byte{},
			MediaType:   TestBinaryReadingMediaType,
		},
	}}
	noBinaryMediaType := validRequest
	noBinaryMediaType.Event.Readings = []dtos.BaseReading{{
		DeviceName: TestDeviceName,
		Name:       TestDeviceResourceName,
		Origin:     TestOriginTime,
		ValueType:  dtos.ValueTypeBinary,
		BinaryReading: dtos.BinaryReading{
			BinaryValue: []byte(TestReadingBinaryValue),
			MediaType:   "",
		},
	}}

	tests := []struct {
		Name               string
		Request            []requests.AddEventRequest
		ErrorExpected      bool
		ExpectedStatusCode int
	}{
		{"Valid - AddEventRequest", []requests.AddEventRequest{validRequest}, false, http.StatusCreated},
		{"Valid - No RequestId", []requests.AddEventRequest{noRequestId}, false, http.StatusCreated},
		{"Invalid - Bad RequestId", []requests.AddEventRequest{badRequestId}, true, http.StatusBadRequest},
		{"Invalid - No Event", []requests.AddEventRequest{noEvent}, true, http.StatusBadRequest},
		{"Invalid - No Event DeviceName", []requests.AddEventRequest{noEventDevice}, true, http.StatusBadRequest},
		{"Invalid - No Event Origin", []requests.AddEventRequest{noEventOrigin}, true, http.StatusBadRequest},
		{"Invalid - No Reading", []requests.AddEventRequest{noReading}, true, http.StatusBadRequest},
		{"Invalid - No Reading DeviceName", []requests.AddEventRequest{noReadingDevice}, true, http.StatusBadRequest},
		{"Invalid - No Reading Name", []requests.AddEventRequest{noReadingName}, true, http.StatusBadRequest},
		{"Invalid - No Reading Origin", []requests.AddEventRequest{noReadingOrigin}, true, http.StatusBadRequest},
		{"Invalid - No Reading ValueType", []requests.AddEventRequest{noReadingValueType}, true, http.StatusBadRequest},
		{"Invalid - Invalid Reading ValueType", []requests.AddEventRequest{invalidReadingInvalidValueType}, true, http.StatusBadRequest},
		{"Invalid - No SimpleReading Value", []requests.AddEventRequest{noSimpleValue}, true, http.StatusBadRequest},
		{"Invalid - No SimpleReading FloatEncoding", []requests.AddEventRequest{noSimpleFloatEnconding}, true, http.StatusBadRequest},
		{"Invalid - No BinaryReading BinaryValue", []requests.AddEventRequest{noBinaryValue}, true, http.StatusBadRequest},
		{"Invalid - No BinaryReading MediaType", []requests.AddEventRequest{noBinaryMediaType}, true, http.StatusBadRequest},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))

			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiEventRoute, reader)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ec.AddEvent)
			handler.ServeHTTP(recorder, req)

			var actualResponse []common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)

			if testCase.ErrorExpected {
				assert.NotEmpty(t, err, "Message is empty")
				assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				return // Test complete for error cases
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, expectedResponseCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, actualResponse[0].ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.ExpectedStatusCode, int(actualResponse[0].StatusCode), "BaseResponse status code not as expected")
			if actualResponse[0].RequestID != "" {
				assert.Equal(t, expectedRequestId, actualResponse[0].RequestID, "RequestID not as expected")
			}
			assert.Equal(t, expectedMessage, actualResponse[0].Message, "Message not as expected")
		})
	}
}
