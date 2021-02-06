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
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ExampleUUID                = "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	testSubscriptionName       = "subscriptionName"
	testSubscriptionCategories = []string{models.SoftwareHealth}
	testSubscriptionLabels     = []string{"label"}
	testSubscriptionChannels   = []dtos.Channel{
		{Type: models.Email, EmailAddresses: []string{"test@example.com"}},
	}
	testSubscriptionDescription    = "description"
	testSubscriptionReceiver       = "receiver"
	testSubscriptionResendLimit    = int64(5)
	testSubscriptionResendInterval = "10s"
	unsupportedChannelType         = "unsupportedChannelType"
	unsupportedCategory            = "unsupportedCategory"
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		notificationContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 30,
				},
			}
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func addSubscriptionRequestData() requests.AddSubscriptionRequest {
	var testAddSubscriptionReq = requests.AddSubscriptionRequest{
		BaseRequest: common.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: common.NewVersionable(),
		},
		Subscription: dtos.Subscription{
			Versionable:    common.NewVersionable(),
			Id:             ExampleUUID,
			Name:           testSubscriptionName,
			Categories:     testSubscriptionCategories,
			Labels:         testSubscriptionLabels,
			Channels:       testSubscriptionChannels,
			Description:    testSubscriptionDescription,
			Receiver:       testSubscriptionReceiver,
			ResendLimit:    testSubscriptionResendLimit,
			ResendInterval: testSubscriptionResendInterval,
		},
	}

	return testAddSubscriptionReq
}

func TestAddSubscription(t *testing.T) {
	expectedRequestId := ExampleUUID
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

	invalidChannelType := addSubscriptionRequestData()
	invalidChannelType.Subscription.Channels = []dtos.Channel{
		{Type: unsupportedChannelType, EmailAddresses: []string{"test@example.com"}},
	}
	invalidEmailAddress := addSubscriptionRequestData()
	invalidEmailAddress.Subscription.Channels = []dtos.Channel{
		{Type: models.Email, EmailAddresses: []string{"test.example.com"}},
	}
	invalidUrl := addSubscriptionRequestData()
	invalidUrl.Subscription.Channels = []dtos.Channel{
		{Type: models.Rest, Url: "http127.0.0.1"},
	}

	noCategoriesAndLabels := addSubscriptionRequestData()
	noCategoriesAndLabels.Subscription.Categories = []string{}
	noCategoriesAndLabels.Subscription.Labels = []string{}
	invalidCategory := addSubscriptionRequestData()
	invalidCategory.Subscription.Categories = []string{unsupportedCategory}

	noReceiver := addSubscriptionRequestData()
	noReceiver.Subscription.Receiver = ""

	invalidResendInterval := addSubscriptionRequestData()
	invalidResendInterval.Subscription.ResendInterval = "10"

	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		{"Invalid - invalid channel type", []requests.AddSubscriptionRequest{invalidChannelType}, http.StatusBadRequest},
		{"Invalid - invalid email address", []requests.AddSubscriptionRequest{invalidEmailAddress}, http.StatusBadRequest},
		{"Invalid - invalid url", []requests.AddSubscriptionRequest{invalidUrl}, http.StatusBadRequest},
		{"Invalid - no categories and labels", []requests.AddSubscriptionRequest{noCategoriesAndLabels}, http.StatusBadRequest},
		{"Invalid - invalid category", []requests.AddSubscriptionRequest{invalidCategory}, http.StatusBadRequest},
		{"Invalid - no receiver", []requests.AddSubscriptionRequest{noReceiver}, http.StatusBadRequest},
		{"Invalid - resendInterval is not specified in ISO8601 format", []requests.AddSubscriptionRequest{invalidResendInterval}, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, v2.ApiSubscriptionRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddSubscription)
			handler.ServeHTTP(recorder, req)
			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}

func TestAllSubscriptions(t *testing.T) {
	subscription := dtos.ToSubscriptionModel(addSubscriptionRequestData().Subscription)
	subscriptions := []models.Subscription{subscription, subscription, subscription}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllSubscriptions", 0, 20).Return(subscriptions, nil)
	dbClientMock.On("AllSubscriptions", 1, 2).Return([]models.Subscription{subscriptions[1], subscriptions[2]}, nil)
	dbClientMock.On("AllSubscriptions", 4, 1).Return([]models.Subscription{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset and limit", "", "", false, 3, http.StatusOK},
		{"Valid - get subscriptions with offset and limit", "1", "2", false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", true, 0, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllSubscriptionRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllSubscriptions)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Subscriptions), "Subscription count is not as expected")
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
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s", v2.ApiSubscriptionByNameRoute, testCase.subscriptionName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.subscriptionName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionByName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.SubscriptionResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
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
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionsByCategory", 0, 20, testCategory).Return([]models.Subscription{}, nil)
	dbClientMock.On("SubscriptionsByCategory", 0, 1, testCategory).Return([]models.Subscription{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset, and limit", "", "", testCategory, false, http.StatusOK},
		{"Valid - get subscriptions with offset, and limit", "0", "1", testCategory, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testCategory, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testCategory, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiSubscriptionByCategoryRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Category: testCase.category})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionsByCategory)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestSubscriptionsByLabel(t *testing.T) {
	testLabel := "label"
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionsByLabel", 0, 20, testLabel).Return([]models.Subscription{}, nil)
	dbClientMock.On("SubscriptionsByLabel", 0, 1, testLabel).Return([]models.Subscription{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset, and limit", "", "", testLabel, false, http.StatusOK},
		{"Valid - get subscriptions with offset, and limit", "0", "1", testLabel, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testLabel, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testLabel, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiSubscriptionByLabelRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Label: testCase.label})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionsByLabel)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestSubscriptionsByReceiver(t *testing.T) {
	testReceiver := "receiver"
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionsByReceiver", 0, 20, testReceiver).Return([]models.Subscription{}, nil)
	dbClientMock.On("SubscriptionsByReceiver", 0, 1, testReceiver).Return([]models.Subscription{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedStatusCode int
	}{
		{"Valid - get subscriptions without offset, and limit", "", "", testReceiver, false, http.StatusOK},
		{"Valid - get subscriptions with offset, and limit", "0", "1", testReceiver, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testReceiver, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testReceiver, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiSubscriptionByReceiverRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Receiver: testCase.receiver})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.SubscriptionsByReceiver)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiSubscriptionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
