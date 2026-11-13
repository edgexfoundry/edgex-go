//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//

package application

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/stretchr/testify/assert"
)

func TestValidateEntityName(t *testing.T) {
	tests := []struct {
		name          string
		entityName    string
		escapeEnabled bool
		errorExpected bool
	}{
		{"valid with escape enabled", "Invalid+Device#Name", true, false},
		{"valid normal name", "camera-device-01", false, false},
		{"valid unicode name", "哈囉世界-传感器", false, false},
		{"valid punctuation name", "camera(local)@site", false, false},
		{"mqtt wildcard plus", "Invalid+Device", false, true},
		{"mqtt wildcard hash", "Device#1", false, true},
		{"mqtt topic separator", "device/sub", false, true},
		{"nats wildcard star", "Device*1", false, true},
		{"nats wildcard gt", "Device>1", false, true},
		{"nats token separator", "Device.1", false, true},
		{"whitespace", "Device Name", false, true},
		{"tab", "Device\tName", false, true},
		{"empty", "", false, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateEntityName(testCase.entityName, testCase.escapeEnabled)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.Equal(t, errors.KindContractInvalid, errors.Kind(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
