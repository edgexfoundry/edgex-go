//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/constants"
	"github.com/edgexfoundry/edgex-go/internal/core/keeper/container"
	"github.com/edgexfoundry/edgex-go/internal/core/keeper/infrastructure/interfaces/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	messageClientMocks "github.com/edgexfoundry/go-mod-messaging/v4/messaging/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/labstack/echo/v4"
)

var flattenValue = map[string]interface{}{
	"Writable": map[string]interface{}{
		"PersistData": false,
		"LogLevel":    "INFO",
		"Area": map[string]interface{}{
			"Width":  7.89,
			"Height": 5.2833124,
		},
	},
	"Labels": []interface{}{"a", "b", "c"},
}

func encodeToBase64Str(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func buildTestKVRequest() requests.UpdateKeysRequest {
	return requests.UpdateKeysRequest{
		BaseRequest: commonDTO.BaseRequest{
			Versionable: commonDTO.NewVersionable(),
			RequestId:   "",
		},
		Value: flattenValue,
	}
}

var (
	notFoundKey = "not-found-key"
	notFoundErr = errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("query key %s not found", notFoundKey), nil)
)

func TestKeys_KeyValue(t *testing.T) {
	key := "test-key"
	logLevel := "INFO"
	kvModel := models.KVS{
		Key: key,
		StoredData: models.StoredData{
			Value: encodeToBase64Str(logLevel),
		},
	}

	rawKey := "test-raw-key"
	rawKV := models.KVS{
		Key: rawKey,
		StoredData: models.StoredData{
			Value: logLevel,
		},
	}

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("KeeperKeys", key, false, false).Return([]models.KVResponse{&kvModel}, nil)
	dbClientMock.On("KeeperKeys", rawKey, false, true).Return([]models.KVResponse{&rawKV}, nil)
	dbClientMock.On("KeeperKeys", notFoundKey, false, false).Return(nil, notFoundErr)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewKVController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		key                string
		plaintext          string
		expectedStatusCode int
		errorExpected      bool
	}{
		{"Valid - GetKey", key, "false", http.StatusOK, false},
		{"Valid - GetKey with plaintext is true", rawKey, "true", http.StatusOK, false},
		{"Invalid - key contains invalid character", "invalidChar:", "false", http.StatusBadRequest, true},
		{"Invalid - key not found", notFoundKey, "false", http.StatusNotFound, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiKVSByKeyRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(constants.Plaintext, testCase.plaintext)

			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(constants.Key)
			c.SetParamValues(testCase.key)
			err = controller.Keys(c)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res struct {
					commonDTO.BaseResponse
					Response []models.KVS
				}
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.key, res.Response[0].Key, "Key from response not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestKeys_KeyOnly(t *testing.T) {
	key := "test-key"
	keyOnlyModel := models.KeyOnly(key)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("KeeperKeys", key, true, false).Return([]models.KVResponse{&keyOnlyModel}, nil)
	dbClientMock.On("KeeperKeys", notFoundKey, true, false).Return(nil, notFoundErr)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewKVController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		key                string
		expectedStatusCode int
		errorExpected      bool
	}{
		{"Valid - GetKey", key, http.StatusOK, false},
		{"Invalid - key contains invalid character", "invalidChar:", http.StatusBadRequest, true},
		{"Invalid - key not found", notFoundKey, http.StatusNotFound, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiKVSByKeyRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(constants.KeyOnly, "true")

			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(constants.Key)
			c.SetParamValues(testCase.key)
			err = controller.Keys(c)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res struct {
					commonDTO.BaseResponse
					Response []models.KeyOnly
				}
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.key, string(res.Response[0]), "Key from response not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestAddKeys(t *testing.T) {
	key := "flattenKey"
	kvRequest := buildTestKVRequest()
	kvModel := requests.UpdateKeysReqToKVModels(kvRequest, key)
	flattenResp := []models.KeyOnly{
		models.KeyOnly(key + "/Writable/PersistData"),
		models.KeyOnly(key + "/Writable/LogLevel"),
		models.KeyOnly(key + "/Writable/Area/Width"),
		models.KeyOnly(key + "/Writable/Area/Height"),
		models.KeyOnly(key + "/Labels"),
	}

	nonFlatten := "nonFlattenKey"
	nonFlattenReq := buildTestKVRequest()
	nonFlattenReq.Value = "test"
	nonFlattenModel := requests.UpdateKeysReqToKVModels(nonFlattenReq, nonFlatten)
	nonFlattenResp := []models.KeyOnly{models.KeyOnly(nonFlatten)}

	nonFlattenEmptyKey := "nonFlattenEmptyKey"
	nonFlattenEmptyReq := buildTestKVRequest()
	nonFlattenEmptyReq.Value = ""
	nonFlattenEmptyModel := requests.UpdateKeysReqToKVModels(nonFlattenEmptyReq, nonFlattenEmptyKey)
	nonFlattenEmptyResp := []models.KeyOnly{models.KeyOnly(nonFlattenEmptyKey)}

	invalidReq := map[string]interface{}{"someInvalidReq": 12345}

	nullValueReq := buildTestKVRequest()
	nullValueReq.Value = nil

	childKeysExistsReq := buildTestKVRequest()
	childKeysExistsKey := "parentKey"
	childKeysExistsReq.Value = "test123"
	childKeysExistsModel := requests.UpdateKeysReqToKVModels(childKeysExistsReq, childKeysExistsKey)

	msgClientMock := &messageClientMocks.MessageClient{}
	msgClientMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddKeeperKeys", kvModel, true).Return(flattenResp, nil)
	dbClientMock.On("AddKeeperKeys", nonFlattenModel, false).Return(nonFlattenResp, nil)
	dbClientMock.On("AddKeeperKeys", nonFlattenEmptyModel, false).Return(nonFlattenEmptyResp, nil)
	dbClientMock.On("AddKeeperKeys", childKeysExistsModel, true).Return(nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "update key failed since child key(s) already exist", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
			return msgClientMock
		},
	})

	controller := NewKVController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name                 string
		request              interface{}
		key                  string
		flatten              string
		expectedRespKeyCount int
		expectedStatusCode   int
		errorExpected        bool
	}{
		{"Valid - AddKeyRequest", kvRequest, key, "true", 5, http.StatusOK, false},
		{"Valid - AddKeyRequest - no flatten", nonFlattenReq, nonFlatten, "false", 1, http.StatusOK, false},
		{"Valid - with value is an empty string - no flatten", nonFlattenEmptyReq, nonFlattenEmptyKey, "false", 1, http.StatusOK, false},
		{"Invalid - no value field in the requests payload", invalidReq, "invalidKey", "false", 0, http.StatusBadRequest, true},
		{"Invalid - key contains invalid character - no flatten", kvRequest, "invalidChar:", "false", 0, http.StatusBadRequest, true},
		{"Invalid - with value is null - no flatten", nullValueReq, "nullValue", "false", 0, http.StatusBadRequest, true},
		{"Invalid - with child keys exist", childKeysExistsReq, childKeysExistsKey, "true", 0, http.StatusBadRequest, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPut, common.ApiKVSByKeyRoute, reader)
			query := req.URL.Query()
			query.Add(constants.Flatten, testCase.flatten)

			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(constants.Key)
			c.SetParamValues(testCase.key)
			err = controller.AddKeys(c)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res struct {
					commonDTO.BaseResponse
					Response []models.KeyOnly
				}
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedRespKeyCount, len(res.Response), "Update key count from response not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteKeys(t *testing.T) {
	key := "test-key"
	prefixExistsKey := "prefix-key"
	keyResp := []models.KeyOnly{models.KeyOnly(key)}

	msgClientMock := &messageClientMocks.MessageClient{}
	msgClientMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeleteKeeperKeys", key, true).Return(keyResp, nil)
	dbClientMock.On("DeleteKeeperKeys", key, false).Return(keyResp, nil)
	dbClientMock.On("DeleteKeeperKeys", prefixExistsKey, false).
		Return(nil, errors.NewCommonEdgeX(errors.KindStatusConflict, "keys having the same prefix prefix-key exist and cannot be deleted", nil))
	dbClientMock.On("DeleteKeeperKeys", notFoundKey, false).Return(nil, notFoundErr)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
			return msgClientMock
		},
	})

	controller := NewKVController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		key                string
		prefixMatch        string
		expectedStatusCode int
		errorExpected      bool
	}{
		{"Valid - DeleteKeys with prefixMatch is true", key, "true", http.StatusOK, false},
		{"Valid - DeleteKeys with prefixMatch is false", key, "false", http.StatusOK, false},
		{"Invalid - DeleteKeys with prefixMatch is false and prefix exists", prefixExistsKey, "false", http.StatusConflict, true},
		{"Invalid - key contains invalid character", "invalidChar:", "false", http.StatusBadRequest, true},
		{"Invalid - key not found", notFoundKey, "false", http.StatusNotFound, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodDelete, constants.ApiKVRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(constants.PrefixMatch, testCase.prefixMatch)

			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(constants.Key)
			c.SetParamValues(testCase.key)
			err = controller.DeleteKeys(c)
			require.NoError(t, err)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res struct {
					commonDTO.BaseResponse
					Response []models.KeyOnly
				}
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.key, string(res.Response[0]), "Key from response not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
