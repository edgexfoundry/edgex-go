//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateAutoEvents(t *testing.T) {
	profile := "test-profile"
	source1 := "source1"
	command1 := "command1"
	deviceProfile := models.DeviceProfile{
		Name:            profile,
		DeviceResources: []models.DeviceResource{{Name: source1}, {Name: "resource2"}},
		DeviceCommands:  []models.DeviceCommand{{Name: command1}, {Name: "command2"}},
	}

	dic := di.NewContainer(di.ServiceConstructorMap{})
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", mock.Anything).Return(deviceProfile, nil)
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
			true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateAutoEvent(dic, testCase.device)
			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
