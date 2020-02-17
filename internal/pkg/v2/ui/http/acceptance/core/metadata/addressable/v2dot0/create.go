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
	"testing"

	dtoErrorV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	dtoCreateV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/create"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	controllerCreate "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/create"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// CreateAddressableForTest creates a new addressable and returns its identity.
func CreateAddressableForTest(t *testing.T, router *mux.Router) (*string, *dtoCreateV2dot0.Request) {
	// create addressable for test to update.
	createRequest := FactoryValidCreateRequest(test.FactoryRandomString())
	w, _ := test.SendRequestWithBody(
		t,
		router,
		controllerCreate.Method,
		controllerCreate.Endpoint,
		test.Marshal(t, createRequest))
	assert.Equal(t, w.Code, http.StatusOK)

	// Unpack the response to get the ID of the newly created addressable.
	var response dtoCreateV2dot0.Response
	test.Unmarshal(
		t,
		test.RecastDTOs(
			t,
			w.Body.Bytes(),
			dtoErrorV2dot0.NewEmptyResponse,
			dtoCreateV2dot0.NewEmptyResponse,
		),
		&response,
	)
	return &response.ID, createRequest
}
