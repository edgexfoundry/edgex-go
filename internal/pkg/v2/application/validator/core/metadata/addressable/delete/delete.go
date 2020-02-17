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

package delete

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dto "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/delete"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/validator"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/validator/common"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// Validate encapsulates the validation behavior for the corresponding request DTO.
func Validate(
	r interface{},
	behavior *application.Behavior,
	execute application.Executable) (response interface{}, status infrastructure.Status) {

	request, ok := r.(*dto.Request)
	response, status = common.Validate(
		r,
		ok,
		func() string {
			return request.RequestID
		},
		behavior,
		application.NewBehaviorSlice(application.Version2, application.KindAddressableDelete, application.ActionDelete),
	)
	if response != nil {
		return
	}

	response, status = validator.NotEmpty(request.RequestID, request.ID, application.StatusAddressableMissingID, r)
	if response != nil {
		return
	}

	return execute(r)
}
