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

package test

import (
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoErrorV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	useCase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/usecases/common/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"
	controller "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/test"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// UseCaseTest verifies test endpoint returns expected result; common implementation intended to be executed by
// each service that includes test support.  Verifies concurrent execution of multiple use case requests.
func UseCaseTest(t *testing.T, router *mux.Router) {
	delayVariations := []struct {
		name      string
		delayInMS int
	}{
		{name: test.NoDelayDescription, delayInMS: 0},
		{name: test.StandardDelayInMSDescription, delayInMS: test.StandardDelayInMS},
	}

	for delayVariationIndex := range delayVariations {
		message := test.FactoryRandomString()
		requestVariations := []struct {
			name             string
			request          []byte
			expectedResponse interface{}
			expectedStatus   int
		}{
			{
				name: test.Join(delayVariations[delayVariationIndex].name, test.One, test.TypeValid),
				request: test.Marshal(
					t,
					useCase.NewRequest(message, delayVariations[delayVariationIndex].delayInMS),
				),
				expectedResponse: useCase.NewResponse(message),
				expectedStatus:   http.StatusOK,
			},
			{
				name: test.Join(delayVariations[delayVariationIndex].name, test.Two, test.TypeValid),
				request: test.Marshal(
					t,
					[]interface{}{
						useCase.NewRequest(message+test.One, delayVariations[delayVariationIndex].delayInMS),
						useCase.NewRequest(message+test.Two, delayVariations[delayVariationIndex].delayInMS),
					},
				),
				expectedResponse: []interface{}{
					useCase.NewResponse(message + test.One),
					useCase.NewResponse(message + test.Two),
				},
				expectedStatus: http.StatusMultiStatus,
			},
		}
		for _, variation := range requestVariations {
			t.Run(
				test.Name(controller.Method, variation.name),
				func(t *testing.T) {
					timer := test.NewTimer()
					w, correlationID := test.SendRequestWithBody(
						t,
						router,
						controller.Method,
						controller.Endpoint,
						variation.request,
					)
					timer.Stop()

					test.AssertCorrelationID(t, w.Header(), correlationID)
					assert.Equal(t, variation.expectedStatus, w.Code)
					test.AssertContentTypeIsJSON(t, w.Header())
					test.AssertJSONBody(
						t,
						variation.expectedResponse,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							useCase.NewEmptyResponse,
						),
					)
					test.AssertElapsedInsideStandardDeviation(
						t,
						timer,
						delayVariations[delayVariationIndex].delayInMS,
					)
				},
			)
		}
	}
}

// BatchTest verifies echo totalRequests sent to batch endpoint return expected results; common implementation
// intended to be executed by each service that includes echo support.  Verifies serial execution of multiple
// batch requests.
func BatchTest(t *testing.T, router *mux.Router) {
	delayVariations := []struct {
		name      string
		delayInMS int
	}{
		{name: test.NoDelayDescription, delayInMS: 0},
		{name: test.StandardDelayInMSDescription, delayInMS: test.StandardDelayInMS},
	}

	for delayVariationIndex := range delayVariations {
		message := test.FactoryRandomString()
		requestVariations := []struct {
			name             string
			totalRequests    int
			request          []byte
			expectedResponse interface{}
		}{
			{
				name: test.Name(
					batch.Method,
					delayVariations[delayVariationIndex].name,
					test.One,
					test.TypeValid,
				),
				totalRequests: 1,
				request: test.Marshal(
					t,
					[]interface{}{
						batchdto.NewTestRequest(
							batchdto.NewCommon(
								application.Version2,
								controller.Kind,
								controller.Action,
								batchdto.StrategySynchronous,
							),
							useCase.NewRequest(message, delayVariations[delayVariationIndex].delayInMS),
						)},
				),
				expectedResponse: []interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(
							application.Version2,
							controller.Kind,
							controller.Action,
							batchdto.StrategySynchronous,
						),
						useCase.NewResponse(message),
					),
				},
			},
			{
				name: test.Name(
					batch.Method,
					delayVariations[delayVariationIndex].name,
					test.Two,
					test.TypeValid,
				),
				totalRequests: 2,
				request: test.Marshal(
					t,
					[]interface{}{
						batchdto.NewTestRequest(
							batchdto.NewCommon(
								application.Version2,
								controller.Kind,
								controller.Action,
								batchdto.StrategySynchronous,
							),
							useCase.NewRequest(message+test.One, delayVariations[delayVariationIndex].delayInMS),
						),
						batchdto.NewTestRequest(
							batchdto.NewCommon(
								application.Version2,
								controller.Kind,
								controller.Action,
								batchdto.StrategySynchronous,
							),
							useCase.NewRequest(message+test.Two, delayVariations[delayVariationIndex].delayInMS),
						),
					},
				),
				expectedResponse: []interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(
							application.Version2,
							controller.Kind,
							controller.Action,
							batchdto.StrategySynchronous,
						),
						useCase.NewResponse(message+test.One),
					),
					batchdto.NewTestRequest(
						batchdto.NewCommon(
							application.Version2,
							controller.Kind,
							controller.Action,
							batchdto.StrategySynchronous,
						),
						useCase.NewResponse(message+test.Two),
					),
				},
			},
		}
		for _, variation := range requestVariations {
			t.Run(
				variation.name,
				func(t *testing.T) {
					timer := test.NewTimer()
					w, correlationID := test.SendRequestWithBody(t, router, batch.Method, batch.Endpoint, variation.request)
					timer.Stop()

					test.AssertCorrelationID(t, w.Header(), correlationID)
					assert.Equal(t, http.StatusMultiStatus, w.Code)
					test.AssertContentTypeIsJSON(t, w.Header())
					test.AssertJSONBody(
						t,
						variation.expectedResponse,
						test.RecastDTOs(t, w.Body.Bytes(), dtoErrorV2dot0.NewEmptyResponse, batchdto.NewEmptyResponse),
					)
					test.AssertElapsedInsideStandardDeviation(
						t,
						timer,
						delayVariations[delayVariationIndex].delayInMS*variation.totalRequests,
					)
				},
			)
		}
	}
}
