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

package update

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoError "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	dto "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/update"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/persistence/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/routable"
	validator "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/validator/core/metadata/addressable/update"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/update"
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

	id := infrastructure.StringToIdentity(request.ID)
	m, status := uc.persistence.FindByID(id)
	if status != infrastructure.StatusSuccess {
		return dtoError.NewResponse(dtoBase.NewResponse(request.RequestID, r, status)), status
	}

	m.Name = update.String(m.Name, request.Name)
	m.Protocol = update.String(m.Protocol, request.Protocol)
	m.Method = update.String(m.Method, request.Method)
	m.Address = update.String(m.Address, request.Address)
	m.Port = update.String(m.Port, request.Port)
	m.Path = update.String(m.Path, request.Path)
	m.Publisher = update.String(m.Publisher, request.Publisher)
	m.User = update.String(m.User, request.User)
	m.Password = update.String(m.Password, request.Password)
	m.Topic = update.String(m.Topic, request.Topic)

	return dto.NewResponse(dtoBase.NewResponseForSuccess(request.RequestID), request.ID), uc.persistence.Save(*m)
}

// EmptyRequest returns an empty request associated with this use case.
func (_ *useCase) EmptyRequest() interface{} {
	return dto.NewEmptyRequest()
}
