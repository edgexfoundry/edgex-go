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
	"encoding/json"
	"fmt"
	"testing"

	dtoErrorV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	dtoV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/create"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/acceptance/core/metadata/addressable/v2dot0"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// assertValid validates addressable add response.
func assertValid(
	t *testing.T,
	router *mux.Router,
	response *dtoV2dot0.Response,
	requestIDs []string,
	createRequests map[string]*dtoV2dot0.Request) {

	inList := func() bool {
		for index := range requestIDs {
			if response.RequestID == requestIDs[index] {
				return true
			}
		}
		return false
	}

	if !inList() {
		assert.Fail(t, fmt.Sprintf("requestID %s not in list %v", response.RequestID, requestIDs))
	}

	test.AssertIsIdentity(t, response.ID)
	v2dot0.AssertCreated(t, createRequests[response.RequestID], v2dot0.GetAddressableById(t, router, &response.ID))
}

// assertV2dot0UseCaseOneValid validates addressable add response; we can't predict what will be
// returned so we consider valid response to have all non-zero fields.
func assertV2dot0UseCaseOneValid(
	t *testing.T,
	router *mux.Router,
	actual []byte,
	requestIDs []string,
	createRequests map[string]*dtoV2dot0.Request) {

	var responseDTO dtoV2dot0.Response

	// single response?
	err := json.Unmarshal(actual, &responseDTO)
	if err == nil {
		assertValid(t, router, &responseDTO, requestIDs, createRequests)
		return
	}

	// multiple  responses?
	var responseDTOs []dtoV2dot0.Response
	err = json.Unmarshal(actual, &responseDTOs)
	if err == nil {
		for i := range responseDTOs {
			assertValid(t, router, &responseDTOs[i], requestIDs, createRequests)
		}
		return
	}

	assert.Fail(t, "unable to validate response: %s", err.Error())
}

// assertV2dot0UseCaseWithOneValidAndOneError validates one successful result and one error result
// of a specific type.
func assertV2dot0UseCaseWithOneValidAndOneError(
	t *testing.T,
	router *mux.Router,
	actual []byte,
	validRequestID string,
	createRequests map[string]*dtoV2dot0.Request,
	status infrastructure.Status) {

	var responseDTOs []*json.RawMessage
	if err := json.Unmarshal(actual, &responseDTOs); err != nil {
		assert.Fail(t, "unable to unmarshal: %s", err.Error())
		return
	}

	assert.Equal(t, 2, len(responseDTOs))

	var validExists = false
	for i := range responseDTOs {
		var responseDTO dtoV2dot0.Response
		if err := json.Unmarshal(*responseDTOs[i], &responseDTO); err != nil {
			continue
		}
		if responseDTO.StatusCode != infrastructure.StatusSuccess {
			continue
		}

		validExists = true
		assertValid(t, router, &responseDTO, []string{validRequestID}, createRequests)
	}
	assert.True(t, validExists)

	var invalidExists = false
	for i := range responseDTOs {
		var responseDTO dtoErrorV2dot0.Response
		if err := json.Unmarshal(*responseDTOs[i], &responseDTO); err != nil {
			continue
		}
		if responseDTO.StatusCode == infrastructure.StatusSuccess {
			continue
		}

		invalidExists = true
		assert.Equal(t, status, responseDTO.StatusCode)
	}
	assert.True(t, invalidExists)

}

// assertV2dot0BatchWithOneValidAndOneError validates one successful result and one error result of a specific type.
func assertV2dot0BatchWithOneValidAndOneError(
	t *testing.T,
	router *mux.Router,
	actual []byte,
	validRequestID string,
	createRequests map[string]*dtoV2dot0.Request,
	status infrastructure.Status) {

	var responseDTOs []*json.RawMessage
	if err := json.Unmarshal(actual, &responseDTOs); err != nil {
		assert.Fail(t, "unable to unmarshal: %s", err.Error())
		return
	}

	assert.Equal(t, 2, len(responseDTOs))

	var validExists = false
	for i := range responseDTOs {
		var batchResponseDTO batchdto.TestResponse
		test.Unmarshal(t, *responseDTOs[i], &batchResponseDTO)

		var responseDTO dtoV2dot0.Response
		if err := json.Unmarshal(*batchResponseDTO.Content, &responseDTO); err != nil {
			continue
		}
		if responseDTO.StatusCode != infrastructure.StatusSuccess {
			continue
		}

		validExists = true
		assertValid(t, router, &responseDTO, []string{validRequestID}, createRequests)
	}
	assert.True(t, validExists)

	var invalidExists = false
	for i := range responseDTOs {
		var batchResponseDTO batchdto.TestResponse
		test.Unmarshal(t, *responseDTOs[i], &batchResponseDTO)

		var responseDTO dtoErrorV2dot0.Response
		if err := json.Unmarshal(*batchResponseDTO.Content, &responseDTO); err != nil {
			continue
		}
		if responseDTO.StatusCode == infrastructure.StatusSuccess {
			continue
		}

		invalidExists = true
		assert.Equal(t, status, responseDTO.StatusCode)
	}
	assert.True(t, invalidExists)
}

// assertV2dot0BatchValid validates addressable add response; we can't predict what will be
// returned so we consider valid response to have all non-zero fields.
func assertV2dot0BatchValid(
	t *testing.T,
	router *mux.Router,
	actual []byte,
	version string,
	kind string,
	action string,
	requestIDs []string,
	createRequests map[string]*dtoV2dot0.Request) {

	assertValidBatch := func(response *batchdto.TestResponse) error {
		assert.Equal(t, version, response.Version)
		assert.Equal(t, kind, response.Kind)
		assert.Equal(t, action, response.Action)

		var responseDTO dtoV2dot0.Response
		err := json.Unmarshal(*response.Content, &responseDTO)
		if err == nil {
			assertValid(t, router, &responseDTO, requestIDs, createRequests)
		}
		return err
	}

	// multiple batch responses?
	responseDTOs := batchdto.EmptyTestResponseSlice()
	err := json.Unmarshal(actual, &responseDTOs)
	if err == nil {
		for i := range responseDTOs {
			if err := assertValidBatch(&responseDTOs[i]); err != nil {
				assert.Fail(t, "unable to unmarshal responseDTO: %s", err.Error())
			}
		}
		return
	}

	assert.Fail(t, "unable to validate response: %s", err.Error())
}
