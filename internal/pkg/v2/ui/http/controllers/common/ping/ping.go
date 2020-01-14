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

package ping

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	useCase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/usecases/common/ping"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/handle"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"

	"github.com/gorilla/mux"
)

const (
	version  = application.Version2
	Kind     = application.KindPing
	Action   = application.ActionCommand
	Method   = controllers.ActionCommand
	Endpoint = controllers.BaseURL + "/ping"
)

// controller contains references to dependencies required by the corresponding Controller contract implementation.
type controller struct{}

// Factory is a factory function that returns an initialized controller receiver struct.
func New() *controller {
	return &controller{}
}

// Add wires up zero or more routes in the provided mux.Router.
func (c *controller) Add(muxRouter *mux.Router, router *router.RouteMap) {
	muxRouter.HandleFunc(
		Endpoint,
		func(w http.ResponseWriter, r *http.Request) {
			handle.UseCaseRequest(w, r, version, Kind, Action, router)
		}).Methods(Method)
}

// Supported returns a slice of Supported (a list of supported behaviors).
func (c *controller) Supported() []common.Supported {
	behavior := application.NewBehavior(version, Kind, Action)
	return []common.Supported{
		common.NewSupported(behavior, useCase.New(behavior)),
	}
}
