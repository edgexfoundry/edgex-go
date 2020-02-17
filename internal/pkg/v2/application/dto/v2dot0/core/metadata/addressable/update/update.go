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

// update contains v2.0 update addressable Request and Response DTOs.
package update

import dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"

// Request defines the input for this use case.
type Request struct {
	dtoBase.Request `json:",inline"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	Protocol        string `json:"protocol"`
	Method          string `json:"method"`
	Address         string `json:"address"`
	Port            string `json:"port"`
	Path            string `json:"path"`
	Publisher       string `json:"publisher"`
	User            string `json:"user"`
	Password        string `json:"password"`
	Topic           string `json:"topic"`
}

// Response defines the output/result for this use case.
type Response struct {
	dtoBase.Response `json:",inline"`
	ID               string `json:"id"`
}

// NewRequest is a factory function that returns a Request for this use case.
func NewRequest(
	baseRequest *dtoBase.Request,
	ID string,
	name string,
	protocol string,
	method string,
	address string,
	port string,
	path string,
	publisher string,
	user string,
	password string,
	topic string) *Request {

	return &Request{
		Request:   *baseRequest,
		ID:        ID,
		Name:      name,
		Protocol:  protocol,
		Method:    method,
		Address:   address,
		Port:      port,
		Path:      path,
		Publisher: publisher,
		User:      user,
		Password:  password,
		Topic:     topic,
	}
}

// NewEmptyRequest returns an uninitialized request structure for this use case.
func NewEmptyRequest() interface{} {
	var request Request
	return &request
}

// NewResponse is a factory function that returns an initialized response struct.
func NewResponse(baseResponse *dtoBase.Response, id string) *Response {
	return &Response{
		Response: *baseResponse,
		ID:       id,
	}
}

// NewEmptyResponse returns an uninitialized response structure for this use case.
func NewEmptyResponse() interface{} {
	var response Response
	return &response
}
