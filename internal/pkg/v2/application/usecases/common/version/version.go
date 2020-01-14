/*******************************************************************************
 * Copyright 2020 Dell Inc.
 *
 * Licensed under the Apache License, UseCase 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

// version implements behavior for the version use case.
package version

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dto "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/version"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/routable"
	validator "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/validator/common/version"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// UseCase contains references to dependencies required by the corresponding Routable contract implementation.
type useCase struct{}

// New is a factory function that returns a validation-wrapped use case as a Routable.
func New(behavior *application.Behavior) application.Routable {
	uc := factory()
	return routable.NewDelegate(
		uc,
		delegate.Apply(
			behavior,
			uc.Execute,
			[]delegate.Handler{
				validator.Validate,
			},
		).Execute,
	)
}

// factory is a factory function that returns an initialized UseCase receiver struct.
func factory() *useCase {
	return &useCase{}
}

// Execute encapsulates the behavior for this use case.
func (uc *useCase) Execute(r interface{}) (interface{}, infrastructure.Status) {
	request, _ := r.(*dto.Request)
	return dto.NewResponse(dtoBase.NewResponseForSuccess(request.RequestID), application.VersionLatest),
		infrastructure.StatusSuccess
}

// EmptyRequest returns an empty request associated with this use case.  This is used in the generic controller
// implementations (ui/http/handle/batch.go and ui/http/handle/usecase.go) to provide a concrete structure to
// unmarshal the request's DTO.
func (_ *useCase) EmptyRequest() interface{} {
	return dto.NewEmptyRequest()
}
