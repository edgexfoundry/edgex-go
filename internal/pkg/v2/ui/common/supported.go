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
)

// Supported defines the common structure used to represent a list of supported behaviors; version specifies the
// version of the behavior (e.g. "2"), kind specifies the type (e.g. "ping"), action specifies a transport-agnostic
// verb (e.g. "read"), and Routable includes the arbitrary unit of behavior to execute and the arbitrary request DTO
// that behavior expects as input.
type Supported struct {
	*application.Behavior
	Routable application.Routable
}

// NewSupported is a factory function that returns a Supported struct.
func NewSupported(behavior *application.Behavior, routable application.Routable) Supported {
	return Supported{
		Behavior: behavior,
		Routable: routable,
	}
}
