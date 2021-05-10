//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"errors"
	"testing"

	"github.com/edgexfoundry/go-mod-registry/v2/registry"
	"github.com/edgexfoundry/go-mod-registry/v2/registry/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {

	rcMock := &mocks.Client{}
	rcMock.On("IsServiceAvailable", "edgex-core-data").Return(true, nil)
	rcMock.On("IsServiceAvailable", "edgex-core-metadata").Return(true, nil)
	rcMock.On("IsServiceAvailable", "edgex-core-command").Return(false, errors.New(""))

	tests := []struct {
		name            string
		services        []string
		rc              registry.Client
		expectedHealthy bool
	}{
		{"healthy", []string{"edgex-core-data", "edgex-core-metadata"}, rcMock, true},
		{"unhealthy - RegisterClient not running", []string{"edgex-core-data", "edgex-core-metadata"}, nil, false},
		{"unhealthy - service not running", []string{"edgex-core-command"}, rcMock, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := GetHealth(tt.services, tt.rc)
			for _, v := range res {
				if tt.expectedHealthy {
					assert.Equal(t, v, healthy)
				} else {
					assert.NotEqual(t, v, healthy)
				}
			}
		})
	}
}
