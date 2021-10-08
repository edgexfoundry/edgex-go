//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleUUID      = "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	testDeviceName   = "testDeviceName"
	testProfileName  = "testProfileName"
	testResourceName = "testResourceName"
)

func simpleReadingData() models.SimpleReading {
	return models.SimpleReading{
		BaseReading: models.BaseReading{
			Id:           exampleUUID,
			Origin:       1616728256236000000,
			DeviceName:   testDeviceName,
			ProfileName:  testProfileName,
			ResourceName: testResourceName,
			ValueType:    common.ValueTypeString,
		},
		Value: "123",
	}
}

func binaryReadingData() models.BinaryReading {
	return models.BinaryReading{
		BaseReading: models.BaseReading{
			Id:           exampleUUID,
			Origin:       1616728256236000001,
			DeviceName:   testDeviceName,
			ProfileName:  testProfileName,
			ResourceName: testResourceName,
			ValueType:    common.ValueTypeBinary,
		},
		BinaryValue: make([]byte, 0),
		MediaType:   "image",
	}
}

func objectReadingData() models.ObjectReading {
	return models.ObjectReading{
		BaseReading: models.BaseReading{
			Id:           exampleUUID,
			Origin:       1616728256236000001,
			DeviceName:   testDeviceName,
			ProfileName:  testProfileName,
			ResourceName: testResourceName,
			ValueType:    common.ValueTypeObject,
		},
		ObjectValue: map[string]interface{}{
			"f1": "ABC",
			"f2": float64(123),
		},
	}
}

func TestConvertObjectsToReadings(t *testing.T) {
	simpleReading := simpleReadingData()
	binaryReading := binaryReadingData()
	objectReading := objectReadingData()

	simpleReadingBytes, err := json.Marshal(simpleReading)
	require.NoError(t, err)
	binaryReadingBytes, err := json.Marshal(binaryReading)
	require.NoError(t, err)
	objectReadingBytes, err := json.Marshal(objectReading)
	require.NoError(t, err)

	readingsData := [][]byte{simpleReadingBytes, binaryReadingBytes, objectReadingBytes}
	expectedReadings := []models.Reading{
		simpleReading, binaryReading, objectReading,
	}

	events, err := convertObjectsToReadings(readingsData)
	require.NoError(t, err)
	assert.Equal(t, expectedReadings, events)
}
