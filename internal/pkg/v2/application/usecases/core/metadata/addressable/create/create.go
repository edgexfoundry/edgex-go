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

package create

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dto "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/create"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/persistence/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/routable"
	validator "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/validator/core/metadata/addressable/create"
	model "github.com/edgexfoundry/edgex-go/internal/pkg/v2/domain/models/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// useCase contains references to dependencies required by the corresponding Routable contract implementation.
type useCase struct {
	persistence metadata.Addressable
}

// New is a factory function that returns a validation-wrapped use case as a Routable.
func New(behavior *application.Behavior, persistence metadata.Addressable) application.Routable {
	uc := factory(persistence)
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

// factory is a factory method that returns an initialized UseCase receiver struct.
func factory(persistence metadata.Addressable) *useCase {
	return &useCase{
		persistence: persistence,
	}
}

// Execute encapsulates the behavior for this use case.
func (uc *useCase) Execute(r interface{}) (interface{}, infrastructure.Status) {
	request, _ := r.(*dto.Request)
	identity := uc.persistence.NextIdentity()
	m := model.New(
		identity,
		request.Name,
		request.Protocol,
		request.Method,
		request.Address,
		request.Port,
		request.Path,
		request.Publisher,
		request.User,
		request.Password,
		request.Topic,
	)

	return dto.NewResponse(
			dtoBase.NewResponseForSuccess(request.RequestID),
			infrastructure.IdentityToString(identity),
		),
		uc.persistence.Save(*m)
}

// EmptyRequest returns an empty request associated with this use case.
func (_ *useCase) EmptyRequest() interface{} {
	return dto.NewEmptyRequest()
}
