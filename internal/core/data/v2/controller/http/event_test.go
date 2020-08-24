//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedCorrelationId = uuid.New().String()
var expectedEventId = uuid.New().String()

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
		ID:         expectedEventId,
		DeviceName: TestDeviceName,
		Origin:     TestOriginTime,
		Readings:   []dtos.BaseReading{testReading},
	},
}

var persistedReading = models.SimpleReading{
	BaseReading: models.BaseReading{
		Id:         ExampleUUID,
		Created:    TestCreatedTime,
		Origin:     TestOriginTime,
		DeviceName: TestDeviceName,
		Name:       TestDeviceResourceName,
		ValueType:  dtos.ValueTypeUint8,
	},
	Value: TestReadingValue,
}

var persistedEvent = models.Event{
	Id:         expectedEventId,
	Pushed:     TestPushedTime,
	DeviceName: TestDeviceName,
	Created:    TestCreatedTime,
	Origin:     TestOriginTime,
	Readings:   []models.Reading{persistedReading},
}

func TestAddEvent(t *testing.T) {
	expectedResponseCode := http.StatusMultiStatus
	expectedRequestId := "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"

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

	validRequest := testAddEvent

	noRequestId := validRequest
	noRequestId.RequestID = ""
	badRequestId := validRequest
	badRequestId.RequestID = "niv3sl"
	noEvent := validRequest
	noEvent.Event = dtos.Event{}
	noEventID := validRequest
	noEventID.Event.ID = ""
	badEventID := validRequest
	badEventID.Event.ID = "DIWNI09320"
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
		{"Invalid - No Event Id", []requests.AddEventRequest{noEventID}, true, http.StatusBadRequest},
		{"Invalid - Bad Event Id", []requests.AddEventRequest{badEventID}, true, http.StatusBadRequest},
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
				assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				return // Test complete for error cases
			}

			require.NoError(t, err)
			assert.Equal(t, expectedResponseCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, actualResponse[0].ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.ExpectedStatusCode, int(actualResponse[0].StatusCode), "BaseResponse status code not as expected")
			if actualResponse[0].RequestID != "" {
				assert.Equal(t, expectedRequestId, actualResponse[0].RequestID, "RequestID not as expected")
			}
			assert.Empty(t, actualResponse[0].Message, "Message should be empty when it is successful")
		})
	}
}

func TestEventById(t *testing.T) {
	validEventId := expectedEventId
	emptyEventId := ""
	invalidEventId := "bad"
	notFoundEventId := NonexistentEventID

	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventById", notFoundEventId).Return(models.Event{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "event doesn't exist in the database", nil))
	dbClientMock.On("EventById", validEventId).Return(persistedEvent, nil)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	ec := NewEventController(dic)

	tests := []struct {
		Name               string
		EventId            string
		ErrorExpected      bool
		ExpectedStatusCode int
	}{
		{"Valid - Find Event by Id", validEventId, false, http.StatusOK},
		{"Invalid - Empty EventId", emptyEventId, true, http.StatusBadRequest},
		{"Invalid - EventId is not an UUID", invalidEventId, true, http.StatusBadRequest},
		{"Invalid - Event doesn't exist", notFoundEventId, true, http.StatusNotFound},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s/%s", contractsV2.ApiEventRoute, contractsV2.Id, testCase.EventId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Id: testCase.EventId})
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ec.EventById)
			handler.ServeHTTP(recorder, req)

			if testCase.ErrorExpected {
				var actualResponse common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, actualResponse.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.ExpectedStatusCode, int(actualResponse.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, actualResponse.Message, "Response message doesn't contain the error message")
			} else {
				var actualResponse responseDTO.EventResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, actualResponse.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.ExpectedStatusCode, int(actualResponse.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.EventId, actualResponse.Event.ID, "Event Id not as expected")
				assert.Empty(t, actualResponse.Message, "Message should be empty when it is successful")
			}
		})
	}
}
