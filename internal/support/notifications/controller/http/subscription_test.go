//
// Copyright (C) 2020-2021 IOTech Ltd
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

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ExampleUUID                = "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	testSubscriptionName       = "subscriptionName"
	testSubscriptionCategories = []string{"category1", "category2"}
	testSubscriptionLabels     = []string{"label"}
	testSubscriptionChannels   = []dtos.Address{
		dtos.NewEmailAddress([]string{"test@example.com"}),
		dtos.NewRESTAddress("host", 123, http.MethodPost),
	}
	testSubscriptionDescription    = "description"
	testSubscriptionReceiver       = "receiver"
	testSubscriptionResendLimit    = 5
	testSubscriptionResendInterval = "10s"
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 30,
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func addSubscriptionRequestData() requests.AddSubscriptionRequest {
	subscription := dtos.Subscription{
		Name:           testSubscriptionName,
		Categories:     testSubscriptionCategories,
		Labels:         testSubscriptionLabels,
		Channels:       testSubscriptionChannels,
		Description:    testSubscriptionDescription,
		Receiver:       testSubscriptionReceiver,
		ResendLimit:    testSubscriptionResendLimit,
		ResendInterval: testSubscriptionResendInterval,
		AdminState:     models.Unlocked,
	}
	return requests.NewAddSubscriptionRequest(subscription)
}

func updateSubscriptionRequestData() requests.UpdateSubscriptionRequest {
	testUUID := ExampleUUID
	testName := testSubscriptionName
	testDescription := testSubscriptionDescription
	subscription := dtos.UpdateSubscription{
		Id:             &testUUID,
		Name:           &testName,
		Channels:       testSubscriptionChannels,
		Receiver:       &testSubscriptionReceiver,
		Categories:     testSubscriptionCategories,
		Labels:         testSubscriptionLabels,
		Description:    &testDescription,
		ResendLimit:    &testSubscriptionResendLimit,
		ResendInterval: &testSubscriptionResendInterval,
	}
	return requests.NewUpdateSubscriptionRequest(subscription)
}

func TestAddSubscription(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	valid := addSubscriptionRequestData()
	model := dtos.ToSubscriptionModel(valid.Subscription)
	dbClientMock.On("AddSubscription", model).Return(model, nil)

	noName := addSubscriptionRequestData()
	noName.Subscription.Name = ""
	noRequestId := addSubscriptionRequestData()
	noRequestId.RequestId = ""

	duplicatedName := addSubscriptionRequestData()
	duplicatedName.Subscription.Name = "duplicatedName"
	model = dtos.ToSubscriptionModel(duplicatedName.Subscription)
	dbClientMock.On("AddSubscription", model).Return(model, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("subscription name %s already exists", model.Name), nil))

	unsupportedChannelType := addSubscriptionRequestData()
	unsupportedChannelType.Subscription.Channels = []dtos.Address{
		dtos.NewMQTTAddress("mqtt-broker", 1883, "publisher", "topic"),
	}
	invalidEmailAddress := addSubscriptionRequestData()
	invalidEmailAddress.Subscription.Channels = []dtos.Address{
		dtos.NewEmailAddress([]string{"test.example.com"}),
	}
	invalidHTTPMethod := addSubscriptionRequestData()
	invalidHTTPMethod.Subscription.Channels = []dtos.Address{
		dtos.NewRESTAddress("host", 123, "foo"),
	}

	noCategoriesAndLabels := addSubscriptionRequestData()
	noCategoriesAndLabels.Subscription.Categories = []string{}
	noCategoriesAndLabels.Subscription.Labels = []string{}

	noReceiver := addSubscriptionRequestData()
	noReceiver.Subscription.Receiver = ""

	invalidResendInterval := addSubscriptionRequestData()
	invalidResendInterval.Subscription.ResendInterval = "10"

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewSubscriptionController(dic)
	assert.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddSubscriptionRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.AddSubscriptionRequest{valid}, http.StatusCreated},
		{"Valid - no request Id", []requests.AddSubscriptionRequest{noRequestId}, http.StatusCreated},
		{"Invalid - no name", []requests.AddSubscriptionRequest{noName}, http.StatusBadRequest},
		{"Invalid - duplicated name", []requests.AddSubscriptionRequest{duplicatedName}, http.StatusConflict},
		{"Invalid - unsupported channel type", []requests.AddSubscriptionRequest{unsupportedChannelType}, http.StatusBadRequest},
		{"Invalid - invalid email address", []requests.AddSubscriptionRequest{invalidEmailAddress}, http.StatusBadRequest},
		{"Invalid - invalid HTTP method", []requests.AddSubscriptionRequest{invalidHTTPMethod}, http.StatusBadRequest},
		{"Invalid - no categories and labels", []requests.AddSubscriptionRequest{noCategoriesAndLabels}, http.StatusBadRequest},
		{"Invalid - no receiver", []requests.AddSubscriptionRequest{noReceiver}, http.StatusBadRequest},
		{"Invalid - resendInterval is not specified in ISO8601 format", []requests.AddSubscriptionRequest{invalidResendInterval}, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiSubscriptionRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddSubscription)
			handler.ServeHTTP(recorder, req)
			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, testCase.request[0].RequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}

func TestAllSubscriptions(t *testing.T) {
	subscription := dtos.ToSubscriptionModel(addSubscriptionRequestData().Subscription)
	subscriptions := []models.Subscription{subscription, subscription, subscription}
	expectedSubscriptionCount := uint32(len(subscriptions))

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionTotalCount").Return(expectedSubscriptionCount, nil)
	dbClientMock.On("AllSubscriptions", 0, 20).Return(subscriptions, nil)
	dbClientMock.On("AllSubscriptions", 1, 2).Return([]models.Subscription{subscriptions[1], subscriptions[2]}, nil)
	dbClientMock.On("AllSubscriptions", 4, 1).Return([]models.Subscription{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewSubscriptionController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		errorExpected      bool
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset and limit", "", "", false, 3, expectedSubscriptionCount, http.StatusOK},
		{"Valid - get subscriptions with offset and limit", "1", "2", false, 2, expectedSubscriptionCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", true, 0, expectedSubscriptionCount, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllSubscriptionRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllSubscriptions)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Subscriptions), "Subscription count is not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Subscription total count is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestSubscriptionByName(t *testing.T) {
	subscription := dtos.ToSubscriptionModel(addSubscriptionRequestData().Subscription)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionByName", subscription.Name).Return(subscription, nil)
	dbClientMock.On("SubscriptionByName", notFoundName).Return(models.Subscription{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "subscription doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewSubscriptionController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		subscriptionName   string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find subscription by name", subscription.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - subscription not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiSubscriptionByNameRoute, testCase.subscriptionName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.subscriptionName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionByName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.SubscriptionResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.subscriptionName, res.Subscription.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestSubscriptionsByCategory(t *testing.T) {
	testCategory := "category"
	expectedSubscriptionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionCountByCategory", testCategory).Return(expectedSubscriptionCount, nil)
	dbClientMock.On("SubscriptionsByCategory", 0, 20, testCategory).Return([]models.Subscription{}, nil)
	dbClientMock.On("SubscriptionsByCategory", 0, 1, testCategory).Return([]models.Subscription{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewSubscriptionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		category           string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset, and limit", "", "", testCategory, false, expectedSubscriptionCount, http.StatusOK},
		{"Valid - get subscriptions with offset, and limit", "0", "1", testCategory, false, expectedSubscriptionCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testCategory, true, expectedSubscriptionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testCategory, true, expectedSubscriptionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiSubscriptionByCategoryRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Category: testCase.category})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionsByCategory)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Subscription total count is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestSubscriptionsByLabel(t *testing.T) {
	testLabel := "label"
	expectedSubscriptionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionCountByLabel", testLabel).Return(expectedSubscriptionCount, nil)
	dbClientMock.On("SubscriptionsByLabel", 0, 20, testLabel).Return([]models.Subscription{}, nil)
	dbClientMock.On("SubscriptionsByLabel", 0, 1, testLabel).Return([]models.Subscription{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewSubscriptionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		label              string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset, and limit", "", "", testLabel, false, expectedSubscriptionCount, http.StatusOK},
		{"Valid - get subscriptions with offset, and limit", "0", "1", testLabel, false, expectedSubscriptionCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testLabel, true, expectedSubscriptionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testLabel, true, expectedSubscriptionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiSubscriptionByLabelRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Label: testCase.label})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionsByLabel)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Subscription total count is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestSubscriptionsByReceiver(t *testing.T) {
	testReceiver := "receiver"
	expectedSubscriptionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionCountByReceiver", testReceiver).Return(expectedSubscriptionCount, nil)
	dbClientMock.On("SubscriptionsByReceiver", 0, 20, testReceiver).Return([]models.Subscription{}, nil)
	dbClientMock.On("SubscriptionsByReceiver", 0, 1, testReceiver).Return([]models.Subscription{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewSubscriptionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		receiver           string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset, and limit", "", "", testReceiver, false, expectedSubscriptionCount, http.StatusOK},
		{"Valid - get subscriptions with offset, and limit", "0", "1", testReceiver, false, expectedSubscriptionCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testReceiver, true, expectedSubscriptionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testReceiver, true, expectedSubscriptionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiSubscriptionByReceiverRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Receiver: testCase.receiver})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionsByReceiver)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Subscription total count is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteSubscriptionByName(t *testing.T) {
	subscription := dtos.ToSubscriptionModel(addSubscriptionRequestData().Subscription)
	noName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteSubscriptionByName", subscription.Name).Return(nil)
	dbClientMock.On("DeleteSubscriptionByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "subscription doesn't exist in the database", nil))
	dbClientMock.On("SubscriptionByName", notFoundName).Return(subscription, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "subscription doesn't exist in the database", nil))
	dbClientMock.On("SubscriptionByName", subscription.Name).Return(subscription, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewSubscriptionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		subscriptionName   string
		expectedStatusCode int
	}{
		{"Valid - delete subscription by name", subscription.Name, http.StatusOK},
		{"Invalid - name parameter is empty", noName, http.StatusBadRequest},
		{"Invalid - subscription not found by name", notFoundName, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiSubscriptionByNameRoute, testCase.subscriptionName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.subscriptionName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteSubscriptionByName)
			handler.ServeHTTP(recorder, req)
			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
			if testCase.expectedStatusCode == http.StatusOK {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestPatchSubscription(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	testReq := updateSubscriptionRequestData()
	subscriptionModel := models.Subscription{
		Id:             *testReq.Subscription.Id,
		Name:           *testReq.Subscription.Name,
		Channels:       dtos.ToAddressModels(testReq.Subscription.Channels),
		Receiver:       *testReq.Subscription.Receiver,
		Categories:     testReq.Subscription.Categories,
		Labels:         testReq.Subscription.Labels,
		Description:    *testReq.Subscription.Description,
		ResendLimit:    *testReq.Subscription.ResendLimit,
		ResendInterval: *testReq.Subscription.ResendInterval,
	}

	valid := testReq
	dbClientMock.On("SubscriptionById", *valid.Subscription.Id).Return(subscriptionModel, nil)
	dbClientMock.On("UpdateSubscription", subscriptionModel).Return(nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.Subscription.Id = nil
	dbClientMock.On("SubscriptionByName", *validWithNoId.Subscription.Name).Return(subscriptionModel, nil)
	validWithNoName := testReq
	validWithNoName.Subscription.Name = nil

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.Subscription.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.Subscription.Id = &emptyString
	emptyId.Subscription.Name = nil
	emptyName := testReq
	emptyName.Subscription.Id = nil
	emptyName.Subscription.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.Subscription.Id = nil
	invalidNoIdAndName.Subscription.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.Subscription.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.Subscription.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("SubscriptionById", *invalidNotFoundId.Subscription.Id).Return(subscriptionModel, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.Subscription.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.Subscription.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("SubscriptionByName", *invalidNotFoundName.Subscription.Name).Return(subscriptionModel, notFoundNameError)

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewSubscriptionController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateSubscriptionRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateSubscriptionRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateSubscriptionRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateSubscriptionRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateSubscriptionRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateSubscriptionRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateSubscriptionRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateSubscriptionRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateSubscriptionRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateSubscriptionRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - no id and name", []requests.UpdateSubscriptionRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiSubscriptionRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchSubscription)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusMultiStatus {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, testCase.request[0].RequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedResponseCode, res[0].StatusCode, "BaseResponse status code not as expected")
				if testCase.expectedResponseCode == http.StatusOK {
					assert.Empty(t, res[0].Message, "Message should be empty when it is successful")
				} else {
					assert.NotEmpty(t, res[0].Message, "Response message doesn't contain the error message")
				}
			} else {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedResponseCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}
