/*******************************************************************************
 * Copyright 2019 VMware Inc.
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

package scheduler

import (
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/intervalaction"
)

func restGetIntervalAction(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	op := intervalaction.NewAllExecutor(dbClient, Configuration.Service)
	intervalActions, err := op.Execute()

	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case errors.ErrLimitExceeded:
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)

		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(intervalActions, w, LoggingClient)
}
