/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package metadata

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/url"

	types "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/command"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
)

func restGetAllCommands(w http.ResponseWriter, _ *http.Request) {
	op := command.NewCommandLoadAll(Configuration.Service, dbClient)
	cmds, err := op.Execute()
	if err != nil {
		switch err.(type) {
		case *types.ErrLimitExceeded:
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(&cmds, w, LoggingClient)
}

func restGetCommandById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid, err := url.QueryUnescape(vars[ID])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	op := command.NewCommandById(dbClient, cid)
	cmd, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case *types.ErrItemNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(cmd, w, LoggingClient)
}

func restGetCommandsByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	op := command.NewCommandsByName(dbClient, n)
	cmds, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.Encode(&cmds, w, LoggingClient)
}

func restGetCommandsByDeviceId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did, err := url.QueryUnescape(vars[ID])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	op := command.NewDeviceIdExecutor(dbClient, did)
	commands, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case *types.ErrItemNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(&commands, w, LoggingClient)
}
