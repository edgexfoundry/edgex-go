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
	"encoding/json"
	"testing"
)

// emptyResponse defines the signature of the functions provided to RecastDTOs to obtain empty responses.
type emptyResponse func() interface{}

// recastToDTO converts a JSON object (in actual) into either a use-case or error response based on matching the
// number and names of the top-level object properties to a corresponding empty use-case response.
func recastToDTO(t *testing.T, actual []byte, emptyErrorResponse, emptyUseCaseResponse emptyResponse) interface{} {
	propertiesFromGeneric := func(i interface{}) map[string]interface{} {
		return i.(map[string]interface{})
	}

	properties := func(i interface{}) map[string]interface{} {
		var value interface{}
		Unmarshal(t, Marshal(t, i), &value)
		return propertiesFromGeneric(value)
	}

	match := func(actual, candidate map[string]interface{}) bool {
		if len(actual) != len(candidate) {
			return false
		}

		for actualKey := range actual {
			if _, ok := candidate[actualKey]; !ok {
				return false
			}
		}

		return true
	}

	var response interface{}
	Unmarshal(t, actual, &response)
	responseProperties := propertiesFromGeneric(response)

	useCaseResponse := emptyUseCaseResponse()
	useCaseResponseProperties := properties(useCaseResponse)
	if match(responseProperties, useCaseResponseProperties) {
		Unmarshal(t, actual, useCaseResponse)
		return useCaseResponse
	}

	errorResponse := emptyErrorResponse()
	Unmarshal(t, actual, errorResponse)
	return errorResponse
}

// RecastDTOs accepts JSON as string (in actual) a two function pointers that return an empty error and empty use-case
// response DTO respectively.  This function is used to cast the JSON into a specific DTO version before marshalling
// the result back into a byte array for downstream processing.
func RecastDTOs(t *testing.T, actual []byte, emptyErrorResponse, emptyUseCaseResponse emptyResponse) []byte {
	responses := []*json.RawMessage{}
	err := json.Unmarshal(actual, &responses)
	if err == nil {
		var result []interface{}
		for responseIndex := range responses {
			result = append(result, recastToDTO(t, *responses[responseIndex], emptyErrorResponse, emptyUseCaseResponse))
		}
		return Marshal(t, result)
	}

	return Marshal(t, recastToDTO(t, actual, emptyErrorResponse, emptyUseCaseResponse))
}
