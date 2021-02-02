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

	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

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
