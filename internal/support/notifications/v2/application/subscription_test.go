//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"net/http"
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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	exampleUUID                = "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
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
	unsupportedChannelType         = "unsupportedChannelType"
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		notificationContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				ResendLimit:    2,
				ResendInterval: "1s",
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

func updateSubscriptionData() dtos.UpdateSubscription {
	return dtos.UpdateSubscription{
		Id:             &exampleUUID,
		Name:           &testSubscriptionName,
		Channels:       testSubscriptionChannels,
		Receiver:       &testSubscriptionReceiver,
		Categories:     testSubscriptionCategories,
		Labels:         testSubscriptionLabels,
		Description:    &testSubscriptionDescription,
		ResendLimit:    &testSubscriptionResendLimit,
		ResendInterval: &testSubscriptionResendInterval,
	}
}

func TestPatchSubscription(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	subscription := updateSubscriptionData()
	model := models.Subscription{
		Id:             *subscription.Id,
		Name:           *subscription.Name,
		Channels:       dtos.ToAddressModels(subscription.Channels),
		Receiver:       *subscription.Receiver,
		Categories:     subscription.Categories,
		Labels:         subscription.Labels,
		Description:    *subscription.Description,
		ResendLimit:    *subscription.ResendLimit,
		ResendInterval: *subscription.ResendInterval,
	}

	valid := updateSubscriptionData()
	dbClientMock.On("SubscriptionById", *valid.Id).Return(model, nil)
	dbClientMock.On("UpdateSubscription", model).Return(nil)

	emptyCategoriesAndLabels := updateSubscriptionData()
	emptyCategoriesAndLabels.Categories = []string{}
	emptyCategoriesAndLabels.Labels = []string{}

	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name              string
		subscription      dtos.UpdateSubscription
		errorExpected     bool
		expectedErrorKind errors.ErrKind
	}{
		{"valid", valid, false, ""},
		{"invalid, empty categories and labels", emptyCategoriesAndLabels, true, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := PatchSubscription(context.Background(), testCase.subscription, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.Equal(t, testCase.expectedErrorKind, errors.Kind(err))

			} else {
				require.NoError(t, err)
			}
		})
	}
}
