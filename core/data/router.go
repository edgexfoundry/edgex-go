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
	b := r.PathPrefix("/api/v1").Subrouter()

	// EVENTS
	// /api/v1/event
	b.HandleFunc("/event", eventHandler).Methods("GET", "PUT", "POST")
	e := b.PathPrefix("/event").Subrouter()
	e.HandleFunc("/scrub", scrubHandler).Methods("DELETE")
	e.HandleFunc("/scruball", scrubAllHandler).Methods("DELETE")
	e.HandleFunc("/count", eventCountHandler).Methods("GET")
	e.HandleFunc("/count/{deviceId}", eventCountByDeviceIdHandler).Methods("GET")
	e.HandleFunc("/{id}", getEventByIdHandler).Methods("GET")
	e.HandleFunc("/id/{id}", eventIdHandler).Methods("DELETE", "PUT")
	e.HandleFunc("/device/{deviceId}/{limit:[0-9]+}", getEventByDeviceHandler).Methods("GET")
	e.HandleFunc("/device/{deviceId}", deleteByDeviceIdHandler).Methods("DELETE")
	e.HandleFunc("/removeold/age/{age:[0-9]+}", eventByAgeHandler).Methods("DELETE")
	e.HandleFunc("/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", eventByCreationTimeHandler).Methods("GET")
	e.HandleFunc("/device/{deviceId}/valuedescriptor/{valueDescriptor}/{limit:[0-9]+}", readingByDeviceFilteredValueDescriptor).Methods("GET")

	// READINGS
	// /api/v1/reading
	b.HandleFunc("/reading", readingHandler).Methods("GET", "PUT", "POST")
	rd := b.PathPrefix("/reading").Subrouter()
	rd.HandleFunc("/count", readingCountHandler).Methods("GET")
	rd.HandleFunc("/id/{id}", deleteReadingByIdHandler).Methods("DELETE")
	rd.HandleFunc("/{id}", getReadingByIdHandler).Methods("GET")
	rd.HandleFunc("/device/{deviceId}/{limit:[0-9]+}", readingByDeviceHandler).Methods("GET")
	rd.HandleFunc("/name/{name}/{limit:[0-9]+}", readingbyValueDescriptorHandler).Methods("GET")
	rd.HandleFunc("/uomlabel/{uomLabel}/{limit:[0-9]+}", readingByUomLabelHandler).Methods("GET")
	rd.HandleFunc("/label/{label}/{limit:[0-9]+}", readingByLabelHandler).Methods("GET")
	rd.HandleFunc("/type/{type}/{limit:[0-9]+}", readingByTypeHandler).Methods("GET")
	rd.HandleFunc("/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", readingByCreationTimeHandler).Methods("GET")
	rd.HandleFunc("/name/{name}/device/{device}/{limit:[0-9]+}", readingByValueDescriptorAndDeviceHandler).Methods("GET")

	// VALUE DESCRIPTORS
	// /api/v1/valuedescriptor
	b.HandleFunc("/valuedescriptor", valueDescriptorHandler).Methods("GET", "PUT", "POST")
	vd := b.PathPrefix("/valuedescriptor").Subrouter()
	vd.HandleFunc("/id/{id}", deleteValueDescriptorByIdHandler).Methods("DELETE")
	vd.HandleFunc("/name/{name}", valueDescriptorByNameHandler).Methods("GET", "DELETE")
	vd.HandleFunc("/{id}", valueDescriptorByIdHandler).Methods("GET")
	vd.HandleFunc("/uomlabel/{uomLabel}", valueDescriptorByUomLabelHandler).Methods("GET")
	vd.HandleFunc("/label/{label}", valueDescriptorByLabelHandler).Methods("GET")
	vd.HandleFunc("/devicename/{device}", valueDescriptorByDeviceHandler).Methods("GET")
	vd.HandleFunc("/deviceid/{id}", valueDescriptorByDeviceIdHandler).Methods("GET")

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/ping", pingHandler)

	return r
}
