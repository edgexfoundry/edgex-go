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
package command

import (
	"net/http"

	mux "github.com/gorilla/mux"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()
	b.HandleFunc(PINGENDPOINT, ping)

	loadDeviceRoutes(b)
	return r
}

func loadDeviceRoutes(b *mux.Router) {

	b.HandleFunc("/device", restGetAllCommands).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	// /api/<version>/device
	d.HandleFunc("/{"+ID+"}", restGetCommandsByDeviceID).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", restGetDeviceCommandByCommandID).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", restPutDeviceCommandByCommandID).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restPutDeviceAdminState).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restPutDeviceOpState).Methods(http.MethodPut)

	// /api/<version>/device/name
	dn := d.PathPrefix("/" + NAME).Subrouter()

	dn.HandleFunc("/{"+NAME+"}", restGetCommandsByDeviceName).Methods(http.MethodGet)
	dn.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restPutDeviceAdminStateByDeviceName).Methods(http.MethodPut)
	dn.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restPutDeviceOpStateByDeviceName).Methods(http.MethodPut)
}

// Respond with PINGRESPONSE to see if the service is alive
func ping(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(CONTENTTYPE, TEXTPLAIN)
	w.Write([]byte(PINGRESPONSE))
}
