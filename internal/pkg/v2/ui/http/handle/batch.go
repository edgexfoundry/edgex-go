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
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoError "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/batchdto"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"
)

// BatchRequest decodes a batch endpoint's request, processes it, and returns a set of results.
func BatchRequest(w http.ResponseWriter, r *http.Request, router *router.RouteMap) {
	defer func() {
		if r.Body != nil {
			_ = r.Body.Close()
		}
	}()

	requests := batchdto.EmptyRequestSlice()
	err := json.NewDecoder(r.Body).Decode(&requests)
	if err != nil {
		httpJSONResult(
			w,
			http.StatusBadRequest,
			dtoError.NewResponse(
				dtoBase.NewResponse("", err.Error(), application.StatusBatchUnmarshalFailure),
			),
		)
		return
	}

	responses := []interface{}{}
	for i := range requests {
		routable, exists := router.FindRoute(requests[i].Version, requests[i].Kind, requests[i].Action)

		var response interface{}
		if exists {
			request := routable.EmptyRequest()
			err := json.Unmarshal(*(requests[i].Content), request)
			if err == nil {
				response, _ = routable.Execute(request)
			} else {
				response = dtoError.NewResponse(
					dtoBase.NewResponse("", string(*requests[i].Content), application.StatusBatchUnmarshalFailure),
				)
			}
		} else {
			response = dtoError.NewResponse(
				dtoBase.NewResponse("", string(*requests[i].Content), application.StatusBatchNotRoutableRequestFailure),
			)
		}
		responses = append(responses, *batchdto.NewResponseFromRequest(&requests[i], response))
	}

	httpJSONResult(w, http.StatusMultiStatus, responses)
}
