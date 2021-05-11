//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	Sender:   "senderA",
	Category: "health-check",
	Severity: models.Normal,
	Content:  "test",
	Status:   models.New,
}

func newTestServer() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func addressData(t *testing.T, ts *httptest.Server) models.Address {
	tsUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)
	host := tsUrl.Hostname()
	port, err := strconv.Atoi(tsUrl.Port())
	require.NoError(t, err)

	return models.RESTAddress{
		BaseAddress: models.BaseAddress{Type: v2.REST, Host: host, Port: port},
		HTTPMethod:  http.MethodGet,
	}
}

func TestNormalSend(t *testing.T) {
	dic := mockDic()
	tests := []struct {
		name                       string
		expectedTransmissionStatus string
	}{
		{"sent successful", models.Sent},
		{"sent failed", models.Failed},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ts := newTestServer()
			ts.Start()
			defer ts.Close()
			address := addressData(t, ts)
			sub.Channels = []models.Address{address}
			trans := models.NewTransmission(sub.Name, address, notification.Id)

			if testCase.expectedTransmissionStatus != models.Sent {
				ts.Close()
			}

			trans, err := normalSend(dic, notification, trans)
			require.NoError(t, err)

			assert.EqualValues(t, testCase.expectedTransmissionStatus, trans.Status)
		})
	}
}

func TestCriticalSend(t *testing.T) {
	dic := mockDic()
	config := notificationContainer.ConfigurationFrom(dic.Get)
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("UpdateTransmission", mock.Anything).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name                       string
		expectedTransmissionStatus string
	}{
		{"critical sent successful", models.Sent},
		{"critical sent failed", models.Escalated},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ts := newTestServer()
			ts.Start()
			defer ts.Close()
			address := addressData(t, ts)
			sub.Channels = []models.Address{address}
			trans := models.NewTransmission(sub.Name, address, notification.Id)

			if testCase.expectedTransmissionStatus != models.Sent {
				ts.Close()
			}

			trans, err := criticalSend(dic, notification, sub, trans)
			require.NoError(t, err)

			assert.EqualValues(t, testCase.expectedTransmissionStatus, trans.Status)
			if testCase.expectedTransmissionStatus != models.Sent {
				assert.Equal(t, config.ResendLimit, trans.ResendCount)
				assert.Equal(t, config.ResendLimit, len(trans.Records))
			} else {
				assert.Equal(t, 1, trans.ResendCount)
				assert.Equal(t, 1, len(trans.Records))
			}
		})
	}
}
