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
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/intervalaction"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
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

func restAddIntervalAction(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	var intervalAction contract.IntervalAction
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&intervalAction)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("error decoding intervalAction" + err.Error())
		return
	}
	LoggingClient.Info("posting new intervalAction: " + intervalAction.String())

	op := intervalaction.NewAddExecutor(dbClient, scClient, intervalAction)
	newId, err := op.Execute()
	if err != nil {
		switch t := err.(type) {
		case errors.ErrIntervalActionNameInUse:
			http.Error(w, t.Error(), http.StatusBadRequest)
		case errors.ErrIntervalNotFound:
			http.Error(w, t.Error(), http.StatusBadRequest)
		default:
			http.Error(w, t.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(newId))
}
