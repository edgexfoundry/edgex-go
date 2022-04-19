//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
)

func TestOnlyOneFieldUpdated(t *testing.T) {
	testService := "test-service"
	testDescription := "test-description"
	testAdminState := models.Locked
	oneUpdated := dtos.UpdateDeviceService{
		Name:        &testService,
		Description: &testDescription,
	}
	twoUpdated := oneUpdated
	twoUpdated.AdminState = &testAdminState

	tests := []struct {
		name      string
		fieldName string
		model     interface{}
		expected  bool
	}{
		{"valid", "Description", oneUpdated, true},
		{"invalid - two fields are updated", "Description", twoUpdated, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, OnlyOneFieldUpdated(tt.fieldName, tt.model))
		})
	}
}
