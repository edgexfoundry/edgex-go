//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package direct

import (
	"errors"
	"net/http"
	"testing"

	"github.com/edgexfoundry/go-mod-registry/v2/registry"
	"github.com/edgexfoundry/go-mod-registry/v2/registry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		expectedErr     bool
		expectedHealthy bool
	}{
		{"valid - healthy", []string{"edgex-core-data", "edgex-core-metadata"}, rcMock, false, true},
		{"valid - unhealthy, service not running", []string{"edgex-core-command"}, rcMock, false, false},
		{"invalid - consul not running", []string{"edgex-core-data", "edgex-core-metadata"}, nil, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := GetHealth(tt.services, tt.rc)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				for _, v := range res {
					if tt.expectedHealthy {
						assert.Equal(t, v.StatusCode, http.StatusOK)
					} else {
						assert.NotEqual(t, v.StatusCode, http.StatusOK)
					}
				}
			}
		})
	}
}
