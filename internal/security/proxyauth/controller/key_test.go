//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoMocks "github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/container"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestAuthController_AddKey(t *testing.T) {
	dic := mockDic()

	e := echo.New()
	controller := NewAuthController(dic)

	validNewKey := "validNewKey"
	validIssuer := "testIssuer"
	validKeyData := dtos.KeyData{
		Type:   common.VerificationKeyType,
		Issuer: validIssuer,
		Key:    validNewKey,
	}
	validKeyName := validKeyData.Issuer + "/" + validKeyData.Type
	validEncryptedKey := "encryptedValidNewKey"
	validReq := requests.AddKeyDataRequest{
		BaseRequest: commonDTO.BaseRequest{
			Versionable: commonDTO.NewVersionable(),
		},
		KeyData: validKeyData,
	}

	noIssuerReq := validReq
	noIssuerReq.KeyData.Issuer = ""

	invalidTypeReq := validReq
	invalidTypeReq.KeyData.Type = "invalidType"

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("KeyExists", validKeyName).Return(false, nil)
	dbClientMock.On("AddKey", validKeyName, validEncryptedKey).Return(nil)
	dbClientMock.On("KeyExists", validKeyName).Return(false, nil)

	cryptoMock := &cryptoMocks.Crypto{}
	cryptoMock.On("Encrypt", validKeyData.Key).Return(validEncryptedKey, nil)

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.CryptoInterfaceName: func(get di.Get) interface{} {
			return cryptoMock
		},
	})

	tests := []struct {
		name           string
		request        requests.AddKeyDataRequest
		expectedStatus int
	}{
		{"Valid - Successful add key", validReq, http.StatusCreated},
		{"Invalid - no issuer request", noIssuerReq, http.StatusBadRequest},
		{"Invalid - invalid type", invalidTypeReq, http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(test.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonBytes))
			req, err := http.NewRequest(http.MethodPost, common.ApiKeyRoute, reader)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := e.NewContext(req, recorder)

			edgexErr := controller.AddKey(ctx)
			require.NoError(t, edgexErr)
			require.Equal(t, test.expectedStatus, recorder.Code)

			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)
			require.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			require.Equal(t, test.expectedStatus, res.StatusCode, "BaseResponse status code not as expected")
		})
	}
}

func TestAuthController_VerificationKeyByIssuer(t *testing.T) {
	dic := mockDic()
	controller := NewAuthController(dic)

	validIssuer := "issuer1"
	validEncryptedKey := "encryptedKey"
	expectedKeyName := validIssuer + "/" + common.VerificationKeyType
	expectedKeyData := dtos.KeyData{Issuer: validIssuer, Type: common.VerificationKeyType, Key: "decryptedKey"}

	invalidIssuer := "invalidIssuer"
	invalidKeyName := invalidIssuer + "/" + common.VerificationKeyType

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ReadKeyContent", expectedKeyName).Return(validEncryptedKey, nil)
	dbClientMock.On("ReadKeyContent", invalidKeyName).Return("", errors.NewCommonEdgeX(errors.KindServerError, "read key error", nil))

	cryptoMock := &cryptoMocks.Crypto{}
	cryptoMock.On("Decrypt", validEncryptedKey).Return([]byte("decryptedKey"), nil)

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.CryptoInterfaceName: func(get di.Get) interface{} {
			return cryptoMock
		},
	})

	tests := []struct {
		name            string
		issuer          string
		expectedStatus  int
		expectedKeyData dtos.KeyData
	}{
		{"Valid - valid issuer", validIssuer, http.StatusOK, expectedKeyData},
		{"Invalid - no issuer request", "", http.StatusBadRequest, dtos.KeyData{}},
		{"Invalid - failed to read key by issuer", invalidIssuer, http.StatusInternalServerError, dtos.KeyData{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiVerificationKeyByIssuerRoute, http.NoBody)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Issuer)
			c.SetParamValues(test.issuer)

			edgexErr := controller.VerificationKeyByIssuer(c)
			require.NoError(t, edgexErr)
			require.Equal(t, test.expectedStatus, recorder.Code)
			if test.expectedStatus == http.StatusOK {
				var res responses.KeyDataResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				require.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				require.Equal(t, expectedKeyData, res.KeyData, "KeyData response not as expected")
			} else {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				require.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				require.Equal(t, test.expectedStatus, res.StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}
