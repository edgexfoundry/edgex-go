//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSecret(t *testing.T) {
	dic := mockDic()

	target := NewCommonController(dic, uuid.NewString())
	assert.NotNil(t, target)

	validRequest := commonDTO.NewSecretRequest(
		"mqtt",
		[]commonDTO.SecretDataKeyValue{
			{Key: "username", Value: "username"},
			{Key: "password", Value: "password"},
		},
	)

	NoPath := validRequest
	NoPath.Path = ""
	validNoRequestId := validRequest
	validNoRequestId.RequestId = ""
	badRequestId := validRequest
	badRequestId.RequestId = "bad requestId"
	noSecrets := validRequest
	noSecrets.SecretData = []commonDTO.SecretDataKeyValue{}
	missingSecretKey := validRequest
	missingSecretKey.SecretData = []commonDTO.SecretDataKeyValue{
		{Key: "", Value: "username"},
	}
	missingSecretValue := validRequest
	missingSecretValue.SecretData = []commonDTO.SecretDataKeyValue{
		{Key: "username", Value: ""},
	}

	mockProvider := &mocks.SecretProvider{}
	mockProvider.On("StoreSecret", validRequest.Path, map[string]string{"password": "password", "username": "username"}).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.SecretProviderName: func(get di.Get) interface{} {
			return mockProvider
		},
	})

	tests := []struct {
		Name               string
		Request            commonDTO.SecretRequest
		ErrorExpected      bool
		ExpectedStatusCode int
	}{
		{"Valid - no requestId", validNoRequestId, false, http.StatusCreated},
		{"Invalid - no path", NoPath, true, http.StatusBadRequest},
		{"Invalid - bad requestId", badRequestId, true, http.StatusBadRequest},
		{"Invalid - no secrets", noSecrets, true, http.StatusBadRequest},
		{"Invalid - missing secret key", missingSecretKey, true, http.StatusBadRequest},
		{"Invalid - missing secret value", missingSecretValue, true, http.StatusBadRequest},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiSecretRoute, reader)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(target.AddSecret)
			handler.ServeHTTP(recorder, req)

			actualResponse := commonDTO.BaseResponse{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion, "Api Version not as expected")
			assert.Equal(t, testCase.ExpectedStatusCode, actualResponse.StatusCode, "BaseResponse status code not as expected")

			if testCase.ErrorExpected {
				assert.NotEmpty(t, actualResponse.Message, "Message is empty")
			} else {
				assert.Empty(t, actualResponse.Message, "Message not empty, as expected")
			}
		})
	}
}

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationInterfaceName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				SecretStore: bootstrapConfig.SecretStoreInfo{
					Type:     "vault",
					Host:     "localhost",
					Port:     8200,
					Path:     "/v1/secret/edgex/device-simple/",
					Protocol: "http",
				},
			}
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}
