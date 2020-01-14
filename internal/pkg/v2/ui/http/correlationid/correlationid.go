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

package correlationid

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/google/uuid"
)

const HTTPHeader = "X-Correlation-Id"

// handler is a mux.Router middleware handler that echos the correlation id header.
func handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := Get(r)
		if len(value) == 0 {
			value = uuid.New().String()
		}
		w.Header().Set(HTTPHeader, value)

		next.ServeHTTP(w, r)
	})
}

// WireUp adds middleware to mux.Router instance to capture/create/return correlation id for every HTTP call.
func WireUp(muxRouter *mux.Router) {
	muxRouter.Use(handler)

	muxRouter.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}),
		).ServeHTTP(w, r)
	})

	muxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(http.NotFoundHandler()).ServeHTTP(w, r)
	})
}

// Get returns correlation id value from request.
func Get(r *http.Request) string {
	return r.Header.Get(HTTPHeader)
}
