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
	dtoV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/create"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/acceptance/core/metadata/addressable/v2dot0"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"
	controller "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/create"

	"github.com/gorilla/mux"
)

// useCaseValidationOne implements a common test case for ensuring request DTOs are properly validated.
func useCaseValidationOne(
	t *testing.T,
	name string,
	modifyRequest func(createRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var createRequest *dtoV2dot0.Request
	requestID := test.FactoryRandomString()

	return test.NewWithoutPreConditionOrCorrelationID(
		test.Join(name, test.One),
		func() string {
			return controller.Endpoint
		},
		func() []byte {
			createRequest = v2dot0.FactoryValidCreateRequest(requestID)
			modifyRequest(createRequest)
			return test.Marshal(t, createRequest)
		},
		func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
			test.AssertJSONBody(
				t,
				dtoErrorV2dot0.NewResponse(dtoBaseV2dot0.NewResponse(requestID, createRequest, expectedStatus)),
				test.RecastDTOs(
					t,
					w.Body.Bytes(),
					dtoErrorV2dot0.NewEmptyResponse,
					dtoV2dot0.NewEmptyResponse,
				),
			)
		},
		http.StatusBadRequest,
	)
}

// useCaseValidationTwo implements a common test case for ensuring request DTOs are properly validated.
func useCaseValidationTwo(
	t *testing.T,
	name string,
	modifyRequest func(createRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var createRequest1, createRequest2 *dtoV2dot0.Request
	requestIDOne := test.FactoryRandomString()
	requestIDTwo := test.FactoryRandomString()

	return test.NewWithoutPreConditionOrCorrelationID(
		test.Join(name, test.Two),
		func() string {
			return controller.Endpoint
		},
		func() []byte {
			createRequest1 = v2dot0.FactoryValidCreateRequest(requestIDOne)
			createRequest2 = v2dot0.FactoryValidCreateRequest(requestIDTwo)
			modifyRequest(createRequest2)
			return test.Marshal(
				t,
				[]interface{}{
					createRequest1,
					createRequest2,
				},
			)
		},
		func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
			assertV2dot0UseCaseWithOneValidAndOneError(
				t,
				router,
				w.Body.Bytes(),
				requestIDOne,
				map[string]*dtoV2dot0.Request{
					requestIDOne: createRequest1,
					requestIDTwo: createRequest2,
				},
				expectedStatus,
			)
		},
		http.StatusMultiStatus,
	)
}

// UseCaseTestCases returns a series of v2.0 test cases to test the ping use-case endpoint.
func UseCaseTestCases(t *testing.T) []*test.Case {
	return []*test.Case{
		func() *test.Case {
			var createRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeValid, test.One),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					createRequest = v2dot0.FactoryValidCreateRequest(requestID)
					return test.Marshal(t, createRequest)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseOneValid(
						t,
						router,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
						[]string{requestID},
						map[string]*dtoV2dot0.Request{
							requestID: createRequest,
						},
					)
				},
				http.StatusOK,
			)
		}(),
		func() *test.Case {
			var createRequest1, createRequest2 *dtoV2dot0.Request
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeValid, test.Two),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					createRequest1 = v2dot0.FactoryValidCreateRequest(requestIDOne)
					createRequest2 = v2dot0.FactoryValidCreateRequest(requestIDTwo)
					return test.Marshal(
						t,
						[]interface{}{
							createRequest1,
							createRequest2,
						},
					)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseOneValid(
						t,
						router,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
						[]string{requestIDOne, requestIDTwo},
						map[string]*dtoV2dot0.Request{
							requestIDOne: createRequest1,
							requestIDTwo: createRequest2,
						},
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
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
					)
				},
				http.StatusBadRequest,
			)
		}(),
		func() *test.Case {
			var createRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()
			invalidJSON := test.InvalidJSON()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeCannotUnmarshal, test.Two),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					createRequest = v2dot0.FactoryValidCreateRequest(requestID)
					return test.Marshal(
						t,
						[]interface{}{
							createRequest,
							invalidJSON,
						},
					)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseWithOneValidAndOneError(
						t,
						router,
						w.Body.Bytes(),
						requestID,
						map[string]*dtoV2dot0.Request{
							requestID: createRequest,
						},
						application.StatusUseCaseUnmarshalFailure,
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			request := v2dot0.FactoryValidCreateRequest("")
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
								application.StatusRequestIdEmptyFailure,
							),
						),
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
					)
				},
				http.StatusBadRequest,
			)
		}(),
		func() *test.Case {
			var createRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeEmptyRequestId, test.Two),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					createRequest = v2dot0.FactoryValidCreateRequest(requestID)
					return test.Marshal(
						t,
						[]interface{}{
							createRequest,
							v2dot0.FactoryValidCreateRequest(""),
						},
					)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseWithOneValidAndOneError(
						t,
						router,
						w.Body.Bytes(),
						requestID,
						map[string]*dtoV2dot0.Request{
							requestID: createRequest,
						},
						application.StatusRequestIdEmptyFailure,
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		useCaseValidationOne(
			t,
			v2dot0.TypeMissingName,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Name = ""
			},
			application.StatusAddressableMissingName,
		),
		useCaseValidationTwo(
			t,
			v2dot0.TypeMissingName,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Name = ""
			},
			application.StatusAddressableMissingName,
		),
		useCaseValidationOne(
			t,
			v2dot0.TypeMissingProtocol,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Protocol = ""
			},
			application.StatusAddressableMissingProtocol,
		),
		useCaseValidationTwo(
			t,
			v2dot0.TypeMissingProtocol,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Protocol = ""
			},
			application.StatusAddressableMissingProtocol,
		),
		useCaseValidationOne(
			t,
			v2dot0.TypeMissingAddress,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Address = ""
			},
			application.StatusAddressableMissingAddress,
		),
		useCaseValidationTwo(
			t,
			v2dot0.TypeMissingAddress,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Address = ""
			},
			application.StatusAddressableMissingAddress,
		),
	}
}

// batchValidationOne implements a common test case for ensuring request DTOs are properly validated.
func batchValidationOne(
	t *testing.T,
	name string,
	kind string,
	action string,
	modifyRequest func(createRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var createRequest *dtoV2dot0.Request
	requestID := test.FactoryRandomString()

	return test.NewWithoutPreConditionOrCorrelationID(
		test.Join(test.One, name),
		func() string {
			return batch.Endpoint
		},
		func() []byte {
			createRequest = v2dot0.FactoryValidCreateRequest(requestID)
			modifyRequest(createRequest)
			return test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						createRequest,
					),
				},
			)
		},
		func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
			test.AssertJSONBody(
				t,
				[]interface{}{
					batchdto.NewResponse(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						dtoErrorV2dot0.NewResponse(dtoBaseV2dot0.NewResponse(requestID, createRequest, expectedStatus)),
					),
				},
				test.RecastDTOs(
					t,
					w.Body.Bytes(),
					dtoErrorV2dot0.NewEmptyResponse,
					batchdto.NewEmptyResponse,
				),
			)
		},
		http.StatusMultiStatus,
	)
}

// batchValidationTwo implements a common test case for ensuring request DTOs are properly validated.
func batchValidationTwo(
	t *testing.T,
	name string,
	kind string,
	action string,
	modifyRequest func(createRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var createRequest1, createRequest2 *dtoV2dot0.Request
	requestIDOne := test.FactoryRandomString()
	requestIDTwo := test.FactoryRandomString()

	return test.NewWithoutPreConditionOrCorrelationID(
		test.Join(test.Two, name),
		func() string {
			return batch.Endpoint
		},
		func() []byte {
			createRequest1 = v2dot0.FactoryValidCreateRequest(requestIDOne)
			createRequest2 = v2dot0.FactoryValidCreateRequest(requestIDTwo)
			modifyRequest(createRequest2)
			return test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						createRequest1,
					),
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						createRequest2,
					),
				},
			)
		},
		func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
			assertV2dot0BatchWithOneValidAndOneError(
				t,
				router,
				w.Body.Bytes(),
				requestIDOne,
				map[string]*dtoV2dot0.Request{
					requestIDOne: createRequest1,
					requestIDTwo: createRequest2,
				},
				expectedStatus,
			)
		},
		http.StatusMultiStatus,
	)
}

// BatchTestCases returns a series of v2.0 test cases to test ping use-cases requests via the batch endpoint.
func BatchTestCases(t *testing.T, kind, action string) []*test.Case {
	return []*test.Case{
		func() *test.Case {
			var createRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.One, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func() []byte {
					createRequest = v2dot0.FactoryValidCreateRequest(requestID)
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								createRequest,
							),
						},
					)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0BatchValid(
						t,
						router,
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
						map[string]*dtoV2dot0.Request{
							requestID: createRequest,
						},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			var createRequest1, createRequest2 *dtoV2dot0.Request
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.Two, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func() []byte {
					createRequest1 = v2dot0.FactoryValidCreateRequest(requestIDOne)
					createRequest2 = v2dot0.FactoryValidCreateRequest(requestIDTwo)
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								createRequest1,
							),
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								createRequest2,
							),
						},
					)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0BatchValid(
						t,
						router,
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
						map[string]*dtoV2dot0.Request{
							requestIDOne: createRequest1,
							requestIDTwo: createRequest2,
						},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		batchValidationOne(
			t,
			v2dot0.TypeMissingName,
			kind,
			action,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Name = ""
			},
			application.StatusAddressableMissingName,
		),
		batchValidationTwo(
			t,
			v2dot0.TypeMissingName,
			kind,
			action,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Name = ""
			},
			application.StatusAddressableMissingName,
		),
		batchValidationOne(
			t,
			v2dot0.TypeMissingProtocol,
			kind,
			action,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Protocol = ""
			},
			application.StatusAddressableMissingProtocol,
		),
		batchValidationTwo(
			t,
			v2dot0.TypeMissingProtocol,
			kind,
			action,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Protocol = ""
			},
			application.StatusAddressableMissingProtocol,
		),
		batchValidationOne(
			t,
			v2dot0.TypeMissingAddress,
			kind,
			action,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Address = ""
			},
			application.StatusAddressableMissingAddress,
		),
		batchValidationTwo(
			t,
			v2dot0.TypeMissingAddress,
			kind,
			action,
			func(createRequest *dtoV2dot0.Request) {
				createRequest.Address = ""
			},
			application.StatusAddressableMissingAddress,
		),
	}
}
