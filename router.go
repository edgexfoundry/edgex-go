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
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

import (
	"github.com/gorilla/mux"
)

func loadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	baseUrl := "/api/v1"

	// EVENTS
	r.HandleFunc(baseUrl+"/event", eventHandler).Methods("GET", "PUT", "POST")
	r.HandleFunc(baseUrl+"/event/scrub", scrubHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/scruball", scrubAllHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/count", eventCountHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/count/{deviceId}", eventCountByDeviceIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/{id}", getEventByIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/id/{id}", eventIdHandler).Methods("DELETE", "PUT")
	r.HandleFunc(baseUrl+"/event/device/{deviceId}/{limit:[0-9]+}", getEventByDeviceHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/device/{deviceId}", deleteByDeviceIdHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/removeold/age/{age:[0-9]+}", eventByAgeHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", eventByCreationTimeHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/device/{deviceId}/valuedescriptor/{valueDescriptor}/{limit:[0-9]+}", readingByDeviceFilteredValueDescriptor).Methods("GET")

	// READINGS
	r.HandleFunc(baseUrl+"/reading", readingHandler).Methods("GET", "PUT", "POST")
	r.HandleFunc(baseUrl+"/reading/count", readingCountHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/id/{id}", deleteReadingByIdHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/reading/{id}", getReadingByIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/device/{deviceId}/{limit:[0-9]+}", readingByDeviceHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/name/{name}/{limit:[0-9]+}", readingbyValueDescriptorHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/uomlabel/{uomLabel}/{limit:[0-9]+}", readingByUomLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/label/{label}/{limit:[0-9]+}", readingByLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/type/{type}/{limit:[0-9]+}", readingByTypeHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", readingByCreationTimeHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/name/{name}/device/{device}/{limit:[0-9]+}", readingByValueDescriptorAndDeviceHandler).Methods("GET")

	// VALUE DESCRIPTORS
	r.HandleFunc(baseUrl+"/valuedescriptor", valueDescriptorHandler).Methods("GET", "PUT", "POST")
	r.HandleFunc(baseUrl+"/valuedescriptor/id/{id}", deleteValueDescriptorByIdHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/valuedescriptor/name/{name}", valueDescriptorByNameHandler).Methods("GET", "DELETE")
	r.HandleFunc(baseUrl+"/valuedescriptor/{id}", valueDescriptorByIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/uomlabel/{uomLabel}", valueDescriptorByUomLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/label/{label}", valueDescriptorByLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/devicename/{device}", valueDescriptorByDeviceHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/deviceid/{id}", valueDescriptorByDeviceIdHandler).Methods("GET")

	// Ping Resource
	r.HandleFunc(baseUrl+"/ping", pingHandler)

	return r
}
