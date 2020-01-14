/*******************************************************************************
 * Copyright 2020 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package batch

import (
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoBaseV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoErrorV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	invalidVersion  = "invalidVersion"
	invalidKind     = "invalidKind"
	invalidAction   = "invalidAction"
	invalidStrategy = "invalidStrategy"
)

// UseCaseTest verifies batch endpoint returns expected result; common implementation intended to be executed by
// each service that includes batch support. Verifies empty use case request array returns empty use case response
// array.
func UseCaseTest(t *testing.T, router *mux.Router) {
	type testVariation struct {
		name             string
		request          []byte
		expectedResponse interface{}
		expectedStatus   int
	}

	invalidJSON := test.InvalidJSON()
	testVariations := []testVariation{
		{
			name:             test.TypeEmpty,
			request:          []byte("[]"),
			expectedResponse: []interface{}{},
			expectedStatus:   http.StatusMultiStatus,
		},
		{
			name:    test.Join(test.TypeCannotUnmarshal, "Transport Request"),
			request: test.Marshal(t, string(invalidJSON)),
			expectedResponse: dtoErrorV2dot0.NewResponse(
				dtoBaseV2dot0.NewResponse(
					"",
					"json: cannot unmarshal string into Go value of type []batchdto.request",
					application.StatusBatchUnmarshalFailure,
				),
			),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: test.Join(test.TypeCannotUnmarshal, "Use-case Request"),
			request: test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(
							application.Version2,
							application.KindTest,
							application.ActionCommand,
							batchdto.StrategySynchronous,
						),
						invalidJSON,
					),
				},
			),
			expectedResponse: []interface{}{
				batchdto.NewTestRequest(
					batchdto.NewCommon(
						application.Version2,
						application.KindTest,
						application.ActionCommand,
						batchdto.StrategySynchronous,
					),
					dtoErrorV2dot0.NewResponse(
						dtoBaseV2dot0.NewResponse(
							"",
							invalidJSON,
							application.StatusBatchUnmarshalFailure,
						),
					),
				),
			},
			expectedStatus: http.StatusMultiStatus,
		},
		{
			name: test.TypeNoRoute,
			request: test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(
							invalidVersion,
							invalidKind,
							invalidAction,
							invalidStrategy,
						),
						dtoBaseV2dot0.NewRequest(""),
					),
				},
			),
			expectedResponse: []interface{}{
				batchdto.NewTestRequest(
					batchdto.NewCommon(invalidVersion, invalidKind, invalidAction, invalidStrategy),
					dtoErrorV2dot0.NewResponse(
						dtoBaseV2dot0.NewResponse(
							"",
							dtoBaseV2dot0.NewRequest(""),
							application.StatusBatchNotRoutableRequestFailure,
						),
					),
				),
			},
			expectedStatus: http.StatusMultiStatus,
		},
	}

	for _, variation := range testVariations {
		t.Run(
			test.Name(batch.Method, variation.name),
			func(t *testing.T) {
				w, correlationID := test.SendRequestWithBody(t, router, batch.Method, batch.Endpoint, variation.request)

				test.AssertCorrelationID(t, w.Header(), correlationID)
				assert.Equal(t, variation.expectedStatus, w.Code)
				test.AssertContentTypeIsJSON(t, w.Header())
				test.AssertJSONBody(t, variation.expectedResponse, w.Body.Bytes())
			})
	}
}
