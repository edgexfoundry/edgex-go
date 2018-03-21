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
 * @original author: Ryan Comer, Dell
 * @version: 0.5.0
 * @updated by:  Jim White, Dell Technologies, Feb 27, 2108
 * Changed func loadRestRoutes to LoadRestRoutes - keeping to Go naming standards (Fede C. initiated
 * during mono repo work)
 *******************************************************************************/
package data

import (
	"net/http"

	"github.com/gorilla/mux"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	// EVENTS
	// /api/v1/event
	b.HandleFunc("/event", eventHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	e := b.PathPrefix("/event").Subrouter()
	e.HandleFunc("/scrub", scrubHandler).Methods(http.MethodDelete)
	e.HandleFunc("/scruball", scrubAllHandler).Methods(http.MethodDelete)
	e.HandleFunc("/count", eventCountHandler).Methods(http.MethodGet)
	e.HandleFunc("/count/{deviceId}", eventCountByDeviceIdHandler).Methods(http.MethodGet)
	e.HandleFunc("/{id}", getEventByIdHandler).Methods(http.MethodGet)
	e.HandleFunc("/id/{id}", eventIdHandler).Methods(http.MethodDelete, http.MethodPut)
	e.HandleFunc("/device/{deviceId}/{limit:[0-9]+}", getEventByDeviceHandler).Methods(http.MethodGet)
	e.HandleFunc("/device/{deviceId}", deleteByDeviceIdHandler).Methods(http.MethodDelete)
	e.HandleFunc("/removeold/age/{age:[0-9]+}", eventByAgeHandler).Methods(http.MethodDelete)
	e.HandleFunc("/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", eventByCreationTimeHandler).Methods(http.MethodGet)
	e.HandleFunc("/device/{deviceId}/valuedescriptor/{valueDescriptor}/{limit:[0-9]+}", readingByDeviceFilteredValueDescriptor).Methods(http.MethodGet)

	// READINGS
	// /api/v1/reading
	b.HandleFunc("/reading", readingHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	rd := b.PathPrefix("/reading").Subrouter()
	rd.HandleFunc("/count", readingCountHandler).Methods(http.MethodGet)
	rd.HandleFunc("/id/{id}", deleteReadingByIdHandler).Methods(http.MethodDelete)
	rd.HandleFunc("/{id}", getReadingByIdHandler).Methods(http.MethodGet)
	rd.HandleFunc("/device/{deviceId}/{limit:[0-9]+}", readingByDeviceHandler).Methods(http.MethodGet)
	rd.HandleFunc("/name/{name}/{limit:[0-9]+}", readingbyValueDescriptorHandler).Methods(http.MethodGet)
	rd.HandleFunc("/uomlabel/{uomLabel}/{limit:[0-9]+}", readingByUomLabelHandler).Methods(http.MethodGet)
	rd.HandleFunc("/label/{label}/{limit:[0-9]+}", readingByLabelHandler).Methods(http.MethodGet)
	rd.HandleFunc("/type/{type}/{limit:[0-9]+}", readingByTypeHandler).Methods(http.MethodGet)
	rd.HandleFunc("/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", readingByCreationTimeHandler).Methods(http.MethodGet)
	rd.HandleFunc("/name/{name}/device/{device}/{limit:[0-9]+}", readingByValueDescriptorAndDeviceHandler).Methods(http.MethodGet)

	// VALUE DESCRIPTORS
	// /api/v1/valuedescriptor
	b.HandleFunc("/valuedescriptor", valueDescriptorHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	vd := b.PathPrefix("/valuedescriptor").Subrouter()
	vd.HandleFunc("/id/{id}", deleteValueDescriptorByIdHandler).Methods(http.MethodDelete)
	vd.HandleFunc("/name/{name}", valueDescriptorByNameHandler).Methods(http.MethodGet, http.MethodDelete)
	vd.HandleFunc("/{id}", valueDescriptorByIdHandler).Methods(http.MethodGet)
	vd.HandleFunc("/uomlabel/{uomLabel}", valueDescriptorByUomLabelHandler).Methods(http.MethodGet)
	vd.HandleFunc("/label/{label}", valueDescriptorByLabelHandler).Methods(http.MethodGet)
	vd.HandleFunc("/devicename/{device}", valueDescriptorByDeviceHandler).Methods(http.MethodGet)
	vd.HandleFunc("/deviceid/{id}", valueDescriptorByDeviceIdHandler).Methods(http.MethodGet)

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/ping", pingHandler)

	return r
}
