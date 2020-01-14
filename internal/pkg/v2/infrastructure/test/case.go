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
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// Endpoint defines the signature of a function that returns an endpoint.
type Endpoint func() string

// PreCondition defines the signature of the pre-condition function provided to the test case.
type PreCondition func(t *testing.T, router *mux.Router)

// CorrelationID defines the signature of a function that returns a correlationID.
type CorrelationID func() string

// Request defines the signature of the request function provided to the test case; function should return the body of
// the request to be sent for the test.
type Request func() []byte

// PostCondition defines the signature of the post-condition function provided to the test case.
type PostCondition func(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder)

// Case contains references to dependencies required by the test case.
type Case struct {
	name           string
	endpoint       Endpoint
	preCondition   PreCondition
	correlationID  CorrelationID
	request        Request
	postCondition  PostCondition
	expectedStatus int
}

// New is a factory function that returns a Case struct.
func New(
	name string,
	endpoint Endpoint,
	preCondition PreCondition,
	correlationID CorrelationID,
	request Request,
	postCondition PostCondition,
	expectedStatus int) *Case {

	return &Case{
		name:           name,
		endpoint:       endpoint,
		preCondition:   preCondition,
		correlationID:  correlationID,
		request:        request,
		postCondition:  postCondition,
		expectedStatus: expectedStatus,
	}
}

// NewWithoutCorrelationID is a factory function that returns a Case struct.
func NewWithoutCorrelationID(
	name string,
	endpoint Endpoint,
	preCondition PreCondition,
	request Request,
	postCondition PostCondition,
	expectedStatus int) *Case {

	return New(
		name,
		endpoint,
		preCondition,
		func() string {
			return FactoryRandomString()
		},
		request,
		postCondition,
		expectedStatus,
	)
}

// NewWithoutPreConditionOrCorrelationID is a factory function that returns a Case struct.
func NewWithoutPreConditionOrCorrelationID(
	name string,
	endpoint Endpoint,
	request Request,
	postCondition PostCondition,
	expectedStatus int) *Case {

	return NewWithoutCorrelationID(
		name,
		endpoint,
		func(t *testing.T, router *mux.Router) {},
		request,
		postCondition,
		expectedStatus,
	)
}

// Name returns the test case's name.
func (c *Case) Name() string {
	return c.name
}

// Endpoint returns the test case's endpoint.
func (c *Case) Endpoint() string {
	return c.endpoint()
}

// PreCondition executes the test case's pre-condition function.
func (c *Case) PreCondition(t *testing.T, router *mux.Router) {
	c.preCondition(t, router)
}

// Request returns the test case's correlationID.
func (c *Case) CorrelationID() string {
	return c.correlationID()
}

// Request returns the test case's request.
func (c *Case) Request() []byte {
	return c.request()
}

// PostCondition executes the test case's post-condition function.
func (c *Case) PostCondition(t *testing.T, router *mux.Router, w *httptest.ResponseRecorder) {
	c.postCondition(t, router, w)
}

// ExpectedStatus returns the test case's expected status.
func (c *Case) ExpectedStatus() int {
	return c.expectedStatus
}
