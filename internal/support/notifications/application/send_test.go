//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/application/channel"
	senderMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/application/channel/mocks"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testHost = "localhost"
	testPort = 8080
)

var sub = models.Subscription{
	Categories:  []string{"health-check"},
	Channels:    nil,
	Description: "test subscription",
	Receiver:    "test user",
	Name:        "TestSubscription",
	AdminState:  models.Unlocked,
}

var notification = models.Notification{
	Sender:      "senderA",
	Category:    "health-check",
	Severity:    models.Normal,
	Content:     "test",
	ContentType: common.ContentTypeJSON,
	Status:      models.New,
}

var testRestAddress = models.RESTAddress{
	BaseAddress: models.BaseAddress{Type: common.REST, Host: testHost, Port: testPort},
	HTTPMethod:  http.MethodGet,
	Path:        "path1",
}
var testRestAddress2 = models.RESTAddress{
	BaseAddress: models.BaseAddress{Type: common.REST, Host: testHost, Port: testPort},
	HTTPMethod:  http.MethodGet,
	Path:        "path2",
}
var testEmailAddress = models.EmailAddress{
	BaseAddress: models.BaseAddress{Type: common.EMAIL, Host: testHost, Port: testPort},
	Recipients:  []string{"test@gamil.com"},
}
var testEmailAddress2 = models.EmailAddress{
	BaseAddress: models.BaseAddress{Type: common.EMAIL, Host: testHost, Port: testPort},
	Recipients:  []string{"test2@gamil.com"},
}

func TestFirstSend(t *testing.T) {
	dic := mockDic()
	restSender := &senderMock.Sender{}
	restSender.On("Send", notification, testRestAddress).Return("", nil)
	restSender.On("Send", notification, testRestAddress2).Return("", errors.NewCommonEdgeX(errors.KindServerError, "fail to send the request", nil))
	emailSender := &senderMock.Sender{}
	emailSender.On("Send", notification, testEmailAddress).Return("", nil)
	emailSender.On("Send", notification, testEmailAddress2).Return("", errors.NewCommonEdgeX(errors.KindServerError, "fail to send the email", nil))
	dic.Update(di.ServiceConstructorMap{
		channel.RESTSenderName: func(get di.Get) interface{} {
			return restSender
		},
		channel.EmailSenderName: func(get di.Get) interface{} {
			return emailSender
		},
	})

	tests := []struct {
		name          string
		address       models.Address
		expectedError bool
	}{
		{"sent rest address successful", testRestAddress, false},
		{"sent email address successful", testEmailAddress, false},
		{"sent rest failed", testRestAddress2, true},
		{"sent email failed", testEmailAddress2, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			sub.Channels = []models.Address{testCase.address}
			trans := models.NewTransmission(sub.Name, testCase.address, notification.Id)

			trans = firstSend(dic, notification, trans)

			assert.Equal(t, 1, len(trans.Records))
			if testCase.expectedError {
				assert.EqualValues(t, models.Failed, trans.Status)
			} else {
				assert.EqualValues(t, models.Sent, trans.Status)
			}
		})
	}
}

func TestReSend(t *testing.T) {
	dic := mockDic()
	config := notificationContainer.ConfigurationFrom(dic.Get)
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("UpdateTransmission", mock.Anything).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	restSender := &senderMock.Sender{}
	restSender.On("Send", notification, testRestAddress).Return("", nil)
	restSender.On("Send", notification, testRestAddress2).Return("", errors.NewCommonEdgeX(errors.KindServerError, "fail to send the request", nil))
	emailSender := &senderMock.Sender{}
	emailSender.On("Send", notification, testEmailAddress).Return("", nil)
	emailSender.On("Send", notification, testEmailAddress2).Return("", errors.NewCommonEdgeX(errors.KindServerError, "fail to send the email", nil))
	dic.Update(di.ServiceConstructorMap{
		channel.RESTSenderName: func(get di.Get) interface{} {
			return restSender
		},
		channel.EmailSenderName: func(get di.Get) interface{} {
			return emailSender
		},
	})

	tests := []struct {
		name          string
		address       models.Address
		expectedError bool
	}{
		{"sent rest address successful", testRestAddress, false},
		{"sent email address successful", testEmailAddress, false},
		{"sent rest failed", testRestAddress2, true},
		{"sent email failed", testEmailAddress2, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			sub.Channels = []models.Address{testCase.address}
			trans := models.NewTransmission(sub.Name, testCase.address, notification.Id)

			trans, err := reSend(dic, notification, sub, trans)
			require.NoError(t, err)

			if testCase.expectedError {
				assert.EqualValues(t, models.Escalated, trans.Status)
				assert.Equal(t, config.Writable.ResendLimit, trans.ResendCount)
				assert.Equal(t, config.Writable.ResendLimit, len(trans.Records))
			} else {
				assert.EqualValues(t, models.Sent, trans.Status)
				assert.Equal(t, 1, trans.ResendCount)
				assert.Equal(t, 1, len(trans.Records))
			}
		})
	}
}
