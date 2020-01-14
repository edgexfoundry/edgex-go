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

package base

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoError "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// Validate provides common validation for BaseRequest properties.
func Validate(r interface{}, requestID string) (interface{}, infrastructure.Status) {
	if len(requestID) == 0 {
		status := application.StatusRequestIdEmptyFailure
		return dtoError.NewResponse(dtoBase.NewResponse(requestID, r, status)), status
	}
	return nil, infrastructure.StatusSuccess
}
