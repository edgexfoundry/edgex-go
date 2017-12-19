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
 *
 * @microservice: core-command-go service
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main


import (
	"net/http"

	mux "github.com/gorilla/mux"
)

func loadRestRoutes() http.Handler {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()
	// TODO b.HandleFunc("/ping", ping)

	loadDeviceRoutes(b)
	return r
}

func loadDeviceRoutes(b *mux.Router) {

	b.HandleFunc("/device", restGetAllCommands).Methods(GET)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	// /api/<version>/device
	d.HandleFunc("/{"+ID+"}", restGetCommandsByDeviceID).Methods(GET)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{CID}", restGetDeviceCommandByCommandID).Methods(GET)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{CID}", restPutDeviceCommandByCommandID).Methods(PUT)
	d.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restPutDeviceAdminState).Methods(PUT)
	d.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restPutDeviceOpState).Methods(PUT)

	// /api/<version>/device/name
	dn := d.PathPrefix("/" + NAME).Subrouter()

	dn.HandleFunc("/{"+NAME+"}", restGetCommandsByDeviceName).Methods(GET)
	dn.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restPutDeviceAdminStateByDeviceName).Methods(PUT)
	dn.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restPutDeviceOpStateByDeviceName).Methods(PUT)
}
