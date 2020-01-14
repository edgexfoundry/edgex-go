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

package test

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// UseCaseAcceptance implements the common acceptance test for use-cases via use-case endpoint.
func UseCaseAcceptance(
	t *testing.T,
	router *mux.Router,
	method string,
	versionVariations map[string][]*Case) {

	for version := range versionVariations {
		for _, variation := range versionVariations[version] {
			t.Run(
				Name(method, version, variation.Name()),
				func(t *testing.T) {
					variation.preCondition(t, router)

					w, correlationID := SendRequest(t,
						router,
						method,
						variation.Endpoint(),
						variation.CorrelationID(),
						variation.Request(),
					)

					AssertCorrelationID(t, w.Header(), correlationID)
					assert.Equal(t, variation.ExpectedStatus(), w.Code)
					AssertContentTypeIsJSON(t, w.Header())
					variation.PostCondition(t, router, w)
				},
			)
		}
	}
}

// BatchAcceptance implements the common acceptance test for use-cases via batch endpoint.
func BatchAcceptance(
	t *testing.T,
	router *mux.Router,
	method string,
	versionVariations map[string][]*Case) {

	for version := range versionVariations {
		for _, variation := range versionVariations[version] {
			t.Run(
				Name(version, method, variation.Name()),
				func(t *testing.T) {
					variation.preCondition(t, router)

					w, correlationID := SendRequest(
						t,
						router,
						method,
						variation.Endpoint(),
						variation.CorrelationID(),
						variation.Request(),
					)

					AssertCorrelationID(t, w.Header(), correlationID)
					assert.Equal(t, variation.ExpectedStatus(), w.Code)
					AssertContentTypeIsJSON(t, w.Header())
					variation.PostCondition(t, router, w)
				},
			)
		}
	}
}
