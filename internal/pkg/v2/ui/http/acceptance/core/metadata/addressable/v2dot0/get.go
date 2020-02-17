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
	dtoReadV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/read"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	controllerRead "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/read"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// GetAddressableById retrieves an addressable by its identity.
func GetAddressableById(t *testing.T, router *mux.Router, id *string) *dtoReadV2dot0.Response {
	w, _ := test.SendRequestWithoutBody(t, router, controllerRead.Method, controllerRead.Endpoint(*id))
	assert.Equal(t, w.Code, http.StatusOK)

	// Unpack the response to get the addressable.
	var readResponse dtoReadV2dot0.Response
	test.Unmarshal(
		t,
		test.RecastDTOs(
			t,
			w.Body.Bytes(),
			dtoErrorV2dot0.NewEmptyResponse,
			dtoReadV2dot0.NewEmptyResponse,
		),
		&readResponse,
	)
	return &readResponse
}

// GetAddressableById retrieves an addressable by its identity.
func GetNonExistentAddressableById(t *testing.T, router *mux.Router, id *string) *dtoErrorV2dot0.Response {
	w, _ := test.SendRequestWithoutBody(t, router, controllerRead.Method, controllerRead.Endpoint(*id))
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// Unpack the response to get the addressable.
	var errorResponse dtoErrorV2dot0.Response
	test.Unmarshal(
		t,
		test.RecastDTOs(
			t,
			w.Body.Bytes(),
			dtoErrorV2dot0.NewEmptyResponse,
			dtoReadV2dot0.NewEmptyResponse,
		),
		&errorResponse,
	)
	return &errorResponse
}
