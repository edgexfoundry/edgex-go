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
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/command"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gorilla/mux"
)

func restGetAllCommands(
	w http.ResponseWriter,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	op := command.NewCommandLoadAll(Configuration.Service, dbClient)
	cmds, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.LimitExceeded, errorconcept.Default.InternalServerError)
		return
	}
	pkg.Encode(&cmds, w, loggingClient)
}

func restGetCommandById(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	cid, err := url.QueryUnescape(vars[ID])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	op := command.NewCommandById(dbClient, cid)
	cmd, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}
	pkg.Encode(cmd, w, loggingClient)
}

func restGetCommandsByName(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	op := command.NewCommandsByName(dbClient, n)
	cmds, err := op.Execute()
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}
	pkg.Encode(&cmds, w, loggingClient)
}

func restGetCommandsByDeviceId(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	did, err := url.QueryUnescape(vars[ID])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	op := command.NewDeviceIdExecutor(dbClient, did)
	commands, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}
	pkg.Encode(&commands, w, loggingClient)
}
