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

package routable

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// delegate contains references to Routable.
type delegate struct {
	routable application.Routable
	execute  application.Executable
}

// NewDelegate is a factory function that returns a routable; fulfills the ui.Routable contract.
func NewDelegate(routable application.Routable, execute application.Executable) *delegate {
	return &delegate{
		routable: routable,
		execute:  execute,
	}
}

// Execute delegates to the application.Execute instance provided to the delegate's factory function.
func (d *delegate) Execute(request interface{}) (response interface{}, status infrastructure.Status) {
	return d.execute(request)
}

// EmptyRequest returns an empty request associated with this use case.  This is used in the generic controller
// implementations (ui/http/handle/batch.go and ui/http/handle/usecase.go) to provide a concrete structure to
// unmarshal the request's DTO.
func (d *delegate) EmptyRequest() interface{} {
	return d.routable.EmptyRequest()
}
