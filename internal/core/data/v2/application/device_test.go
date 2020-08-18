//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	mocksV2 "github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/stretchr/testify/require"
)

func TestCheckDevice(t *testing.T) {
	testDeviceName := "Test Device"

	dic := mocksV2.NewMockDIC()
	// set MetaDataCheck Config Writable to true
	dic.Update(di.ServiceConstructorMap{
		dataContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					MetaDataCheck: true,
				},
			}
		},
	})
	err := checkDevice(testDeviceName, context.Background(), dic)
	require.NoError(t, err)

	mdc := mocksV2.MetadataDeviceClientFrom(dic.Get)
	// asserts the CheckForDevice method was called when MetaDataCheck Config Writable is true
	mdc.AssertCalled(t, "CheckForDevice", context.Background(), testDeviceName)

	dic2 := mocksV2.NewMockDIC()
	// set MetaDataCheck Config Writable to false
	dic2.Update(di.ServiceConstructorMap{
		dataContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					MetaDataCheck: false,
				},
			}
		},
	})
	err = checkDevice(testDeviceName, context.Background(), dic2)
	require.NoError(t, err)

	mdc = mocksV2.MetadataDeviceClientFrom(dic2.Get)
	// asserts the CheckForDevice method was NOT called when MetaDataCheck Config Writable is false
	mdc.AssertNotCalled(t, "CheckForDevice", context.Background(), testDeviceName)
}
