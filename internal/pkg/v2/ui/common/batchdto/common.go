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

// batchdto defines the DTO envelope that wraps a use-case DTO for batch requests.  Unlike use-case DTOs (which are
// defined in the application layer, the batch DTO envelope is defined in the ui layer.  It's syntactic sugar that
// may vary based on transport implementation; every use-case request (whether received via use-case-specific
// endpoint or batch endpoint) is resolved to a call to an application layer's use-case implementation.
package batchdto

const (
	// StrategySynchronous indicates a response should be produced synchronously.
	StrategySynchronous = "sync"

	// StrategyAsynchronousPush indicates a response should be produced asynchronously and pushed to the requestor.
	// It is not currently supported.
	StrategyAsynchronousPush = "async-push"

	// StrategyAsynchronousPull indicates a token should be provided as part of an immediate response, the actual
	// response should be produced asynchronously, and the requestor will use the token to retrieve the response.
	// It is not currently supported.
	StrategyAsynchronousPull = "async-poll"
)

// Common defines the common content for request, response, and test DTO structs.
type Common struct {
	Version  string `json:"version"`
	Kind     string `json:"type"`
	Action   string `json:"action"`
	Strategy string `json:"strategy"`
}

// NewCommon is a factory function that returns a Common struct.
func NewCommon(version, kind, action, strategy string) *Common {
	return &Common{
		Version:  version,
		Kind:     kind,
		Action:   action,
		Strategy: strategy,
	}
}

// NewResponseCommonFromRequestCommon is a factory function that returns an initialized Common struct; fields are taken
// from provided request.
func NewResponseCommonFromRequestCommon(request *Common) *Common {
	return NewCommon(request.Version, request.Kind, request.Action, request.Strategy)
}
