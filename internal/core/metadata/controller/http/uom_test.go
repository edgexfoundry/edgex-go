//
// Copyright (C) 2022-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/uom"
)

func TestUnitOfMeasureController_UnitsOfMeasure(t *testing.T) {
	testUoM := uom.UnitsOfMeasureImpl{
		Source: "test global source",
		Units: map[string]uom.Unit{
			"unit1": uom.Unit{
				Source: "test unit source",
				Values: []string{"v1", "v2", "v3"},
			},
		},
	}
	dic := mockDic()
	dic.Update(di.ServiceConstructorMap{
		container.UnitsOfMeasureInterfaceName: func(get di.Get) interface{} {
			return &testUoM
		},
	})

	controller := NewUnitOfMeasureController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name             string
		accept           string
		expectedResponse any
	}{
		{"valid - json response", common.ContentTypeJSON, responses.NewUnitsOfMeasureResponse("", "", http.StatusOK, testUoM)},
		{"valid - yaml response", common.ContentTypeYAML, testUoM},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiUnitsOfMeasureRoute, http.NoBody)
			req.Header.Set(common.Accept, testCase.accept)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.UnitsOfMeasure)
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "HTTP status code not as expected")
			if testCase.accept == common.ContentTypeJSON {
				expectedBytes, err := json.Marshal(testCase.expectedResponse)
				require.NoError(t, err)
				assert.JSONEq(t, recorder.Body.String(), string(expectedBytes), "JSON response not as expected")

				actualResponse := responses.UnitsOfMeasureResponse{}
				err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
				require.NoError(t, err)

				assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion, "Api Version not as expected")
				assert.Equal(t, http.StatusOK, actualResponse.StatusCode, "BaseResponse status code not as expected")
				assert.Empty(t, actualResponse.Message, "Message should be empty when it is successful")
			} else {
				actualResponse := uom.UnitsOfMeasureImpl{}
				err = yaml.Unmarshal(recorder.Body.Bytes(), &actualResponse)
				require.NoError(t, err)

				assert.Equal(t, actualResponse, testCase.expectedResponse, "YAML response not as expected")
			}
		})
	}
}
