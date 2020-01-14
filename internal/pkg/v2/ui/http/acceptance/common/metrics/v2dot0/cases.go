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

package v2dot0

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoBaseV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoErrorV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	dtoV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/metrics"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"
	controller "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/metrics"

	"github.com/gorilla/mux"
)

// UseCaseTestCases returns a series of v2.0 test cases to test the ping use-case endpoint.
func UseCaseTestCases(t *testing.T) []*test.Case {
	return []*test.Case{
		func() *test.Case {
			requestID := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeValid, test.One),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					return test.Marshal(t, dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestID)))
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					assertValidV2dot0UseCaseMetricsResponse(
						t,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
						[]string{requestID},
					)
				},
				http.StatusOK,
			)
		}(),
		func() *test.Case {
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeValid, test.Two),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					return test.Marshal(
						t,
						[]interface{}{
							dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestIDOne)),
							dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestIDTwo)),
						},
					)
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					assertValidV2dot0UseCaseMetricsResponse(
						t,
						test.RecastDTOs(t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
						[]string{requestIDOne, requestIDTwo},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			invalidJSON := test.InvalidJSON()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeCannotUnmarshal, test.One),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					return test.Marshal(t, invalidJSON)
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					test.AssertJSONBody(
						t,
						dtoErrorV2dot0.NewResponse(
							dtoBaseV2dot0.NewResponse(
								"",
								invalidJSON,
								application.StatusUseCaseUnmarshalFailure),
						),
						w.Body.Bytes(),
					)
				},
				http.StatusBadRequest,
			)
		}(),
		func() *test.Case {
			requestID := test.FactoryRandomString()
			invalidJSON := test.InvalidJSON()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeValid, test.TypeCannotUnmarshal, test.Two),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					return test.Marshal(
						t,
						[]interface{}{
							dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestID)),
							invalidJSON,
						},
					)
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseMetricsResponseOneValidAndOneError(
						t,
						w.Body.Bytes(),
						requestID,
						application.StatusUseCaseUnmarshalFailure,
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			request := dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(""))

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeEmptyRequestId, test.One),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					return test.Marshal(t, request)
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					test.AssertJSONBody(
						t,
						dtoErrorV2dot0.NewResponse(
							dtoBaseV2dot0.NewResponse(
								"",
								request,
								application.StatusRequestIdEmptyFailure),
						),
						w.Body.Bytes(),
					)
				},
				http.StatusBadRequest,
			)
		}(),
		func() *test.Case {
			requestID := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeValid, test.TypeEmptyRequestId, test.Two),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					return test.Marshal(
						t,
						[]interface{}{
							dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestID)),
							dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest("")),
						},
					)
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseMetricsResponseOneValidAndOneError(
						t,
						w.Body.Bytes(),
						requestID,
						application.StatusRequestIdEmptyFailure,
					)
				},
				http.StatusMultiStatus,
			)
		}(),
	}
}

// BatchTestCases returns a series of v2.0 test cases to test ping use-cases requests via the batch endpoint.
func BatchTestCases(t *testing.T, kind, action string) []*test.Case {
	return []*test.Case{
		func() *test.Case {
			requestID := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.One, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func() []byte {
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewResponse(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestID)),
							),
						},
					)
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					assertValidV2dot0BatchMetricsResponse(
						t,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							batchdto.NewEmptyResponse,
						),
						application.Version2,
						kind,
						action,
						[]string{requestID},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.Two, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func() []byte {
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestIDOne)),
							),
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								dtoV2dot0.NewRequest(dtoBaseV2dot0.NewRequest(requestIDTwo)),
							),
						},
					)
				},
				func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
					assertValidV2dot0BatchMetricsResponse(
						t,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							batchdto.NewEmptyResponse,
						),
						application.Version2,
						kind,
						action,
						[]string{requestIDOne, requestIDTwo},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
	}
}
