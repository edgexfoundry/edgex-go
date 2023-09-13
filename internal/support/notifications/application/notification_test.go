//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces/mocks"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func TestPurgeNotification(t *testing.T) {
	configuration := &config.ConfigurationStruct{
		Retention: config.NotificationRetention{
			Enabled:  true,
			Interval: "1s",
			MaxCap:   5,
			MinCap:   3,
		},
	}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	tests := []struct {
		name              string
		notificationCount uint32
	}{
		{"invoke notification purging", configuration.Retention.MaxCap},
		{"not invoke notification purging", configuration.Retention.MinCap},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			dbClientMock := &dbMock.DBClient{}
			notification := models.Notification{}
			dbClientMock.On("LatestNotificationByOffset", configuration.Retention.MinCap).Return(notification, nil)
			dbClientMock.On("NotificationTotalCount").Return(testCase.notificationCount, nil)
			dbClientMock.On("CleanupNotificationsByAge", mock.Anything).Return(nil)
			dic.Update(di.ServiceConstructorMap{
				container.DBClientInterfaceName: func(get di.Get) interface{} {
					return dbClientMock
				},
			})
			err := purgeNotification(dic)
			require.NoError(t, err)
			if testCase.notificationCount >= configuration.Retention.MaxCap {
				dbClientMock.AssertCalled(t, "CleanupNotificationsByAge", mock.Anything)
			} else {
				dbClientMock.AssertNotCalled(t, "CleanupNotificationsByAge", mock.Anything)
			}
		})
	}
}
