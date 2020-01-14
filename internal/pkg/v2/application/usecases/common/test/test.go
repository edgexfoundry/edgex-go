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

// test implements behavior for the test use case.
package test

import (
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// UseCase contains references to dependencies required by the corresponding Routable contract implementation.
type UseCase struct{}

// New is a factory function that returns an initialized UseCase receiver struct.
func New() *UseCase {
	return &UseCase{}
}

// Execute encapsulates the behavior for this use case.
func (_ *UseCase) Execute(r interface{}) (interface{}, infrastructure.Status) {
	request, _ := r.(*request)
	if request.DelayInMS != 0 {
		time.Sleep(time.Duration(request.DelayInMS) * time.Millisecond)
	}
	return NewResponse(request.Message), infrastructure.StatusSuccess
}

// EmptyRequest returns an empty request associated with this use case.  This is used in the generic controller
// implementations (ui/http/handle/batch.go and ui/http/handle/usecase.go) to provide a concrete structure to
// unmarshal the request's DTO.
func (_ *UseCase) EmptyRequest() interface{} {
	return NewEmptyRequest()
}
