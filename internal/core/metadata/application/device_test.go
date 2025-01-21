//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateParentProfileAndAutoEvents(t *testing.T) {
	profile := "test-profile"
	notFountProfileName := "notFoundProfile"
	source1 := "source1"
	command1 := "command1"
	deviceProfile := models.DeviceProfile{
		Name:            profile,
		DeviceResources: []models.DeviceResource{{Name: source1}, {Name: "resource2"}},
		DeviceCommands:  []models.DeviceCommand{{Name: command1}, {Name: "command2"}},
	}

	dic := di.NewContainer(di.ServiceConstructorMap{})
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", profile).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", notFountProfileName).Return(models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "not found", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name          string
		device        models.Device
		errorExpected bool
	}{
		{"empty profile",
			models.Device{},
			false,
		},
		{"not found profile",
			models.Device{
				ProfileName: notFountProfileName,
			},
			true,
		},
		{"no auto events",
			models.Device{
				ProfileName: profile,
			},
			false,
		},
		{"resource exist",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1s"}},
			},
			false,
		},
		{"command exist",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1s"}, {SourceName: command1, Interval: "1s"}},
			},
			false,
		},
		{"resource not exist",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: "notFoundSource", Interval: "1s"}},
			},
			true,
		},
		{"interval format not valid",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1"}},
			},
			true,
		},
		{"no profile",
			models.Device{
				AutoEvents: []models.AutoEvent{{SourceName: source1, Interval: "1s"}},
			},
			false,
		},
		{"resource match regex",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: "res.*", Interval: "1s"}},
			},
			false,
		},
		{"is own parent",
			models.Device{
				ProfileName: profile,
				Parent:      "me",
				Name:        "me",
			},
			true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateParentProfileAndAutoEvent(dic, testCase.device)
			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestForceAddDevice(t *testing.T) {
	invalidDeviceName := "invalidDevice"
	validDeviceName := "validDevice"
	invalidDeviceName2 := "invalidDevice2"
	invalidDevice := models.Device{Name: invalidDeviceName}
	invalidDevice2 := models.Device{Name: invalidDeviceName2}
	returnedDevice := models.Device{Name: validDeviceName}

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
			}
		},
	})

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceByName", invalidDeviceName).Return(models.Device{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query", nil))
	dbClientMock.On("DeviceByName", validDeviceName).Return(returnedDevice, nil)
	dbClientMock.On("DeviceByName", invalidDeviceName2).Return(invalidDevice2, nil)
	dbClientMock.On("UpdateDevice", returnedDevice).Return(nil)
	dbClientMock.On("UpdateDevice", invalidDevice2).Return(errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to update", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name          string
		device        models.Device
		errorExpected bool
	}{
		{"invalid - DeviceByName error", invalidDevice, true},
		{"valid", returnedDevice, false},
		{"invalid - UpdateDevice error", invalidDevice2, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, _ := correlation.FromContextOrNew(context.Background())
			result, err := updateDevice(testCase.device, ctx, dic)
			if testCase.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, returnedDevice.Id, result)
			}
		})
	}
}
