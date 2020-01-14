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

package handle

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg/concurrent"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoError "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"
)

// UseCaseRequest decodes, processes, and returns a result for a use-case-specific endpoint's request.
func UseCaseRequest(w http.ResponseWriter, r *http.Request, version, kind, action string, router *router.RouteMap) {
	defer func() {
		if r.Body != nil {
			_ = r.Body.Close()
		}
	}()

	routable, exists := router.FindRoute(version, kind, action)
	if !exists {
		http.Error(w, "route not found", http.StatusNotFound)
		return
	}

	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err != io.EOF {
		httpJSONResult(
			w,
			http.StatusBadRequest,
			dtoError.NewResponse(dtoBase.NewResponse("", err.Error(), application.StatusUseCaseContentErrorFailure)),
		)
		return
	}

	// multiple requests?
	requests := []*json.RawMessage{}
	err = json.Unmarshal(body, &requests)
	if err == nil {
		// execute requests concurrently
		var closures []concurrent.Closure
		for i := range requests {
			closures = append(
				closures,
				func(jsonRequest *json.RawMessage) concurrent.Closure {
					return func() interface{} {
						request := routable.EmptyRequest()
						if err := json.Unmarshal(*jsonRequest, &request); err == nil {
							result, _ := routable.Execute(request)
							return result
						}
						return dtoError.NewResponse(
							dtoBase.NewResponse("", string(*jsonRequest), application.StatusUseCaseUnmarshalFailure),
						)
					}
				}(requests[i]),
			)
		}
		httpJSONResult(w, http.StatusMultiStatus, concurrent.ExecuteAndAggregateResults(closures))
		return
	}

	// single request?
	request := routable.EmptyRequest()
	err = json.Unmarshal(body, &request)
	if err == nil {
		response, status := routable.Execute(request)
		httpJSONResult(w, statusToHTTPStatusCode(status), response)
		return
	}

	httpJSONResult(
		w,
		http.StatusBadRequest,
		dtoError.NewResponse(dtoBase.NewResponse("", string(body), application.StatusUseCaseUnmarshalFailure)),
	)
}

// SingleUseCaseRequest processes and returns a result for a use-case-specific request created in a controller.
func SingleUseCaseRequest(
	w http.ResponseWriter,
	r *http.Request,
	request interface{},
	version,
	kind,
	action string,
	router *router.RouteMap) {

	defer func() {
		if r.Body != nil {
			_ = r.Body.Close()
		}
	}()

	routable, exists := router.FindRoute(version, kind, action)
	if !exists {
		http.Error(w, "route not found", http.StatusNotFound)
		return
	}

	response, status := routable.Execute(request)
	httpJSONResult(w, statusToHTTPStatusCode(status), response)
}

// statusToHTTPStatusCode translates application-level transport-agnostic status result into an HTTP status code.
func statusToHTTPStatusCode(status infrastructure.Status) int {
	if status == infrastructure.StatusSuccess {
		return http.StatusOK
	}
	return http.StatusBadRequest
}
