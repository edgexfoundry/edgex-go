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

package common

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoError "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/validator/common/base"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/validator/common/behavior"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// getRequestID defines the signature of a closure provided to Validate that is called to obtain the requestID from
// the caller's type-asserted struct.
type getRequestID func() string

// Validate handles common generic validation after initial type assertion; includes base validation and behavior
// validation.
func Validate(
	r interface{},
	assertionOK bool,
	getRequestID getRequestID,
	b *application.Behavior,
	supportedBehaviors []*application.Behavior) (interface{}, infrastructure.Status) {

	if !assertionOK {
		status := application.StatusTypeAssertionFailure
		return dtoError.NewResponse(dtoBase.NewResponse("", r, status)), status
	}

	requestID := getRequestID()
	response, status := base.Validate(r, requestID)
	if response != nil {
		return response, status
	}

	response, status = behavior.Validate(r, requestID, b, supportedBehaviors)
	if response != nil {
		return response, status
	}

	return nil, infrastructure.StatusSuccess
}
