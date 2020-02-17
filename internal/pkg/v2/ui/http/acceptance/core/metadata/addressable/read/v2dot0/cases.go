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
	dtoCreateV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/create"
	dtoV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/read"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/acceptance/core/metadata/addressable/v2dot0"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"
	controller "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/read"

	"github.com/gorilla/mux"
)

// factoryValidReadRequest returns a valid addressable read request.
func factoryValidReadRequest(requestID, ID string) *dtoV2dot0.Request {
	return dtoV2dot0.NewRequest(
		dtoBaseV2dot0.NewRequest(requestID),
		ID,
	)
}

// UseCaseTestCases returns a series of v2.0 test cases to test the ping use-case endpoint.
func UseCaseTestCases(_ *testing.T) []*test.Case {
	return []*test.Case{
		func() *test.Case {
			var id *string
			var createRequest *dtoCreateV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.New(
				test.Join(test.TypeValid, test.One),
				func() string {
					return controller.Endpoint(*id)
				},
				func(t *testing.T, router *mux.Router) {
					id, createRequest = v2dot0.CreateAddressableForTest(t, router)
				},
				func() string {
					return requestID
				},
				func() []byte {
					return []byte{}
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseValid(
						t,
						router,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
						[]string{requestID},
						map[string]*dtoCreateV2dot0.Request{
							requestID: createRequest,
						},
					)
				},
				http.StatusOK,
			)
		}(),
		func() *test.Case {
			id := infrastructure.NewIdentityString()
			var readRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.New(
				test.Join(v2dot0.TypeIDNotInPersistence, test.One),
				func() string {
					return controller.Endpoint(id)
				},
				func(t *testing.T, router *mux.Router) {
					readRequest = factoryValidReadRequest(requestID, id)
				},
				func() string {
					return requestID
				},
				func() []byte {
					return []byte{}
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					test.AssertJSONBody(
						t,
						dtoErrorV2dot0.NewResponse(
							dtoBaseV2dot0.NewResponse(
								requestID,
								readRequest,
								infrastructure.StatusPersistenceNotFound,
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
	}
}

// batchValidationOne implements a common test case for ensuring request DTOs are properly validated.
func batchValidationOne(
	t *testing.T,
	name string,
	kind string,
	action string,
	modifyRequest func(readRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var readRequest *dtoV2dot0.Request
	requestID := test.FactoryRandomString()

	return test.NewWithoutPreConditionOrCorrelationID(
		test.Join(test.One, name),
		func() string {
			return batch.Endpoint
		},
		func() []byte {
			readRequest = factoryValidReadRequest(requestID, infrastructure.NewIdentityString())
			modifyRequest(readRequest)
			return test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						readRequest,
					),
				},
			)
		},
		func(t *testing.T, _ *mux.Router, w *httptest.ResponseRecorder) {
			test.AssertJSONBody(
				t,
				[]interface{}{
					batchdto.NewResponse(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						dtoErrorV2dot0.NewResponse(dtoBaseV2dot0.NewResponse(requestID, readRequest, expectedStatus)),
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
	modifyRequest func(readRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var id1, id2 *string
	var createRequest1 *dtoCreateV2dot0.Request
	var readRequest1, readRequest2 *dtoV2dot0.Request
	requestIDOne := test.FactoryRandomString()
	requestIDTwo := test.FactoryRandomString()

	return test.NewWithoutCorrelationID(
		test.Join(test.Two, name),
		func() string {
			return batch.Endpoint
		},
		func(t *testing.T, router *mux.Router) {
			id1, createRequest1 = v2dot0.CreateAddressableForTest(t, router)
			id2, _ = v2dot0.CreateAddressableForTest(t, router)
		},
		func() []byte {
			readRequest1 = factoryValidReadRequest(requestIDOne, *id1)
			readRequest2 = factoryValidReadRequest(requestIDTwo, *id2)
			modifyRequest(readRequest2)
			return test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						readRequest1,
					),
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						readRequest2,
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
				map[string]*dtoCreateV2dot0.Request{
					requestIDOne: createRequest1,
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
			var id *string
			var createRequest *dtoCreateV2dot0.Request
			var readRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.One, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id, createRequest = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					readRequest = factoryValidReadRequest(requestID, *id)
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								readRequest,
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
						map[string]*dtoCreateV2dot0.Request{
							requestID: createRequest,
						},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			var id1, id2 *string
			var createRequest1, createRequest2 *dtoCreateV2dot0.Request
			var readRequest1, readRequest2 *dtoV2dot0.Request
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.Two, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id1, createRequest1 = v2dot0.CreateAddressableForTest(t, router)
					id2, createRequest2 = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					readRequest1 = factoryValidReadRequest(requestIDOne, *id1)
					readRequest2 = factoryValidReadRequest(requestIDTwo, *id2)
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								readRequest1,
							),
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								readRequest2,
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
						map[string]*dtoCreateV2dot0.Request{
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
			v2dot0.TypeMissingID,
			kind,
			action,
			func(readRequest *dtoV2dot0.Request) {
				readRequest.ID = ""
			},
			application.StatusAddressableMissingID,
		),
		batchValidationTwo(
			t,
			v2dot0.TypeMissingID,
			kind,
			action,
			func(readRequest *dtoV2dot0.Request) {
				readRequest.ID = ""
			},
			application.StatusAddressableMissingID,
		),
		batchValidationOne(
			t,
			v2dot0.TypeIDNotInPersistence,
			kind,
			action,
			func(readRequest *dtoV2dot0.Request) {
				readRequest.ID = infrastructure.NewIdentityString()
			},
			infrastructure.StatusPersistenceNotFound,
		),
		batchValidationTwo(
			t,
			v2dot0.TypeIDNotInPersistence,
			kind,
			action,
			func(readRequest *dtoV2dot0.Request) {
				readRequest.ID = infrastructure.NewIdentityString()
			},
			infrastructure.StatusPersistenceNotFound,
		),
	}
}
