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
	dtoV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/update"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/acceptance/core/metadata/addressable/v2dot0"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"
	controller "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/update"

	"github.com/gorilla/mux"
)

// factoryValidUpdateRequest returns a valid addressable update request.
func factoryValidUpdateRequest(requestID, ID string) *dtoV2dot0.Request {
	return dtoV2dot0.NewRequest(
		dtoBaseV2dot0.NewRequest(requestID),
		ID,
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
	)
}

// useCaseValidationOne implements a common test case for ensuring request DTOs are properly validated.
func useCaseValidationOne(
	t *testing.T,
	name string,
	modifyRequest func(updateRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var id *string
	var updateRequest *dtoV2dot0.Request
	requestID := test.FactoryRandomString()

	return test.NewWithoutCorrelationID(
		test.Join(name, test.One),
		func() string {
			return controller.Endpoint
		},
		func(t *testing.T, router *mux.Router) {
			id, _ = v2dot0.CreateAddressableForTest(t, router)
		},
		func() []byte {
			updateRequest = factoryValidUpdateRequest(requestID, *id)
			modifyRequest(updateRequest)
			return test.Marshal(t, updateRequest)
		},
		func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
			test.AssertJSONBody(
				t,
				dtoErrorV2dot0.NewResponse(dtoBaseV2dot0.NewResponse(requestID, updateRequest, expectedStatus)),
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
	modifyRequest func(updateRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var id1, id2 *string
	var updateRequest1, updateRequest2 *dtoV2dot0.Request
	requestIDOne := test.FactoryRandomString()
	requestIDTwo := test.FactoryRandomString()

	return test.NewWithoutCorrelationID(
		test.Join(name, test.Two),
		func() string {
			return controller.Endpoint
		},
		func(t *testing.T, router *mux.Router) {
			id1, _ = v2dot0.CreateAddressableForTest(t, router)
			id2, _ = v2dot0.CreateAddressableForTest(t, router)
		},
		func() []byte {
			updateRequest1 = factoryValidUpdateRequest(requestIDOne, *id1)
			updateRequest2 = factoryValidUpdateRequest(requestIDTwo, *id2)
			modifyRequest(updateRequest2)
			return test.Marshal(
				t,
				[]interface{}{
					updateRequest1,
					updateRequest2,
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
					requestIDOne: updateRequest1,
					requestIDTwo: updateRequest2,
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
			var id *string
			var createRequest *dtoCreateV2dot0.Request
			var updateRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.TypeValid, test.TypeUpdateOneProperty, test.One),
				func() string {
					return controller.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id, createRequest = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest = dtoV2dot0.NewRequest(
						dtoBaseV2dot0.NewRequest(requestID),
						*id,
						test.FactoryRandomString(), // replace name only
						createRequest.Protocol,
						createRequest.Method,
						createRequest.Address,
						createRequest.Port,
						createRequest.Path,
						createRequest.Publisher,
						createRequest.User,
						createRequest.Password,
						createRequest.Topic,
					)
					return test.Marshal(t, updateRequest)
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
							requestID: updateRequest,
						},
					)
				},
				http.StatusOK,
			)
		}(),
		func() *test.Case {
			var id1, id2 *string
			var createRequest1, createRequest2 *dtoCreateV2dot0.Request
			var updateRequest1, updateRequest2 *dtoV2dot0.Request
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.TypeValid, test.TypeUpdateOneProperty, test.Two),
				func() string {
					return controller.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id1, createRequest1 = v2dot0.CreateAddressableForTest(t, router)
					id2, createRequest2 = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest1 = dtoV2dot0.NewRequest(
						dtoBaseV2dot0.NewRequest(requestIDOne),
						*id1,
						test.FactoryRandomString(), // replace name only
						createRequest1.Protocol,
						createRequest1.Method,
						createRequest1.Address,
						createRequest1.Port,
						createRequest1.Path,
						createRequest1.Publisher,
						createRequest1.User,
						createRequest1.Password,
						createRequest1.Topic,
					)
					updateRequest2 = dtoV2dot0.NewRequest(
						dtoBaseV2dot0.NewRequest(requestIDTwo),
						*id2,
						test.FactoryRandomString(), // replace name only
						createRequest2.Protocol,
						createRequest2.Method,
						createRequest2.Address,
						createRequest2.Port,
						createRequest2.Path,
						createRequest2.Publisher,
						createRequest2.User,
						createRequest2.Password,
						createRequest2.Topic,
					)
					return test.Marshal(
						t,
						[]interface{}{
							updateRequest1,
							updateRequest2,
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
							requestIDOne: updateRequest1,
							requestIDTwo: updateRequest2,
						},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			var id *string
			var updateRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.TypeValid, test.TypeUpdateAllProperties, test.One),
				func() string {
					return controller.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id, _ = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest = factoryValidUpdateRequest(requestID, *id)
					return test.Marshal(t, *updateRequest)
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
							requestID: updateRequest,
						},
					)
				},
				http.StatusOK,
			)
		}(),
		func() *test.Case {
			var id1, id2 *string
			var updateRequest1, updateRequest2 *dtoV2dot0.Request
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.TypeValid, test.TypeUpdateAllProperties, test.Two),
				func() string {
					return controller.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id1, _ = v2dot0.CreateAddressableForTest(t, router)
					id2, _ = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest1 = factoryValidUpdateRequest(requestIDOne, *id1)
					updateRequest2 = factoryValidUpdateRequest(requestIDTwo, *id2)
					return test.Marshal(
						t,
						[]interface{}{
							updateRequest1,
							updateRequest2,
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
							requestIDOne: updateRequest1,
							requestIDTwo: updateRequest2,
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
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					test.AssertJSONBody(
						t,
						dtoErrorV2dot0.NewResponse(
							dtoBaseV2dot0.NewResponse(
								"",
								invalidJSON,
								application.StatusUseCaseUnmarshalFailure,
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
			var id *string
			var updateRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()
			invalidJSON := test.InvalidJSON()

			return test.NewWithoutCorrelationID(
				test.Join(test.TypeCannotUnmarshal, test.Two),
				func() string {
					return controller.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id, _ = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest = factoryValidUpdateRequest(requestID, *id)
					return test.Marshal(
						t,
						[]interface{}{
							updateRequest,
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
							requestID: updateRequest,
						},
						application.StatusUseCaseUnmarshalFailure,
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			var updateRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutPreConditionOrCorrelationID(
				test.Join(test.TypeUpdateNonExistent, test.One),
				func() string {
					return controller.Endpoint
				},
				func() []byte {
					updateRequest = factoryValidUpdateRequest(requestID, infrastructure.NewIdentityString())
					return test.Marshal(t, *updateRequest)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					test.AssertJSONBody(
						t,
						dtoErrorV2dot0.NewResponse(
							dtoBaseV2dot0.NewResponse(
								requestID,
								updateRequest,
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
		func() *test.Case {
			var id1 *string
			var updateRequest1, updateRequest2 *dtoV2dot0.Request
			requestIDOne := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.TypeUpdateNonExistent, test.Two),
				func() string {
					return controller.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id1, _ = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest1 = factoryValidUpdateRequest(requestIDOne, *id1)
					updateRequest2 = factoryValidUpdateRequest(
						test.FactoryRandomString(),
						infrastructure.NewIdentityString(),
					)
					return test.Marshal(
						t,
						[]interface{}{
							updateRequest1,
							updateRequest2,
						},
					)
				},
				func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
					assertV2dot0UseCaseWithOneValidAndOneError(
						t,
						router,
						test.RecastDTOs(
							t,
							w.Body.Bytes(),
							dtoErrorV2dot0.NewEmptyResponse,
							dtoV2dot0.NewEmptyResponse,
						),
						requestIDOne,
						map[string]*dtoV2dot0.Request{
							requestIDOne: updateRequest1,
						},
						infrastructure.StatusPersistenceNotFound,
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		useCaseValidationOne(
			t,
			v2dot0.TypeMissingID,
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = ""
			},
			application.StatusAddressableMissingID,
		),
		useCaseValidationTwo(
			t,
			v2dot0.TypeMissingID,
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = ""
			},
			application.StatusAddressableMissingID,
		),
		useCaseValidationOne(
			t,
			v2dot0.TypeIDNotInPersistence,
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = infrastructure.NewIdentityString()
			},
			infrastructure.StatusPersistenceNotFound,
		),
		useCaseValidationTwo(
			t,
			v2dot0.TypeIDNotInPersistence,
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = infrastructure.NewIdentityString()
			},
			infrastructure.StatusPersistenceNotFound,
		),
	}
}

// batchValidationOne implements a common test case for ensuring request DTOs are properly validated.
func batchValidationOne(
	t *testing.T,
	name string,
	kind string,
	action string,
	modifyRequest func(updateRequest *dtoV2dot0.Request),
	expectedStatus infrastructure.Status) *test.Case {

	var id *string
	var updateRequest *dtoV2dot0.Request
	requestID := test.FactoryRandomString()

	return test.NewWithoutCorrelationID(
		test.Join(test.One, name),
		func() string {
			return batch.Endpoint
		},
		func(t *testing.T, router *mux.Router) {
			id, _ = v2dot0.CreateAddressableForTest(t, router)
		},
		func() []byte {
			updateRequest = factoryValidUpdateRequest(requestID, *id)
			modifyRequest(updateRequest)
			return test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						updateRequest,
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
						dtoErrorV2dot0.NewResponse(dtoBaseV2dot0.NewResponse(requestID, updateRequest, expectedStatus)),
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

	var id1, id2 *string
	var updateRequest1, updateRequest2 *dtoV2dot0.Request
	requestIDOne := test.FactoryRandomString()
	requestIDTwo := test.FactoryRandomString()

	return test.NewWithoutCorrelationID(
		test.Join(test.Two, name),
		func() string {
			return batch.Endpoint
		},
		func(t *testing.T, router *mux.Router) {
			id1, _ = v2dot0.CreateAddressableForTest(t, router)
			id2, _ = v2dot0.CreateAddressableForTest(t, router)
		},
		func() []byte {
			updateRequest1 = factoryValidUpdateRequest(requestIDOne, *id1)
			updateRequest2 = factoryValidUpdateRequest(requestIDTwo, *id2)
			modifyRequest(updateRequest2)
			return test.Marshal(
				t,
				[]interface{}{
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						updateRequest1,
					),
					batchdto.NewTestRequest(
						batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
						updateRequest2,
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
					requestIDOne: updateRequest1,
					requestIDTwo: updateRequest2,
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
			var updateRequest *dtoV2dot0.Request
			requestID := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.One, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id, _ = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest = factoryValidUpdateRequest(requestID, *id)
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								updateRequest,
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
							requestID: updateRequest,
						},
					)
				},
				http.StatusMultiStatus,
			)
		}(),
		func() *test.Case {
			var id1, id2 *string
			var updateRequest1, updateRequest2 *dtoV2dot0.Request
			requestIDOne := test.FactoryRandomString()
			requestIDTwo := test.FactoryRandomString()

			return test.NewWithoutCorrelationID(
				test.Join(test.Two, test.TypeValid),
				func() string {
					return batch.Endpoint
				},
				func(t *testing.T, router *mux.Router) {
					id1, _ = v2dot0.CreateAddressableForTest(t, router)
					id2, _ = v2dot0.CreateAddressableForTest(t, router)
				},
				func() []byte {
					updateRequest1 = factoryValidUpdateRequest(requestIDOne, *id1)
					updateRequest2 = factoryValidUpdateRequest(requestIDTwo, *id2)
					return test.Marshal(
						t,
						[]interface{}{
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								updateRequest1,
							),
							batchdto.NewTestRequest(
								batchdto.NewCommon(application.Version2, kind, action, batchdto.StrategySynchronous),
								updateRequest2,
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
							requestIDOne: updateRequest1,
							requestIDTwo: updateRequest2,
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
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = ""
			},
			application.StatusAddressableMissingID,
		),
		batchValidationTwo(
			t,
			v2dot0.TypeMissingID,
			kind,
			action,
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = ""
			},
			application.StatusAddressableMissingID,
		),
		batchValidationOne(
			t,
			v2dot0.TypeIDNotInPersistence,
			kind,
			action,
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = infrastructure.NewIdentityString()
			},
			infrastructure.StatusPersistenceNotFound,
		),
		batchValidationTwo(
			t,
			v2dot0.TypeIDNotInPersistence,
			kind,
			action,
			func(updateRequest *dtoV2dot0.Request) {
				updateRequest.ID = infrastructure.NewIdentityString()
			},
			infrastructure.StatusPersistenceNotFound,
		),
	}
}
