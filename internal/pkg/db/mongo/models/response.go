/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package models

import (
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

/*
 * Response for a Get or Put request to a service
 *
 *
 * Response Struct
 */
type Response struct {
	Code           string   `bson:"code"`
	Description    string   `bson:"description"`
	ExpectedValues []string `bson:"expectedValues"`
}

func (r Response) ToContract() contract.Response {
	return contract.Response{
		Code:           r.Code,
		Description:    r.Description,
		ExpectedValues: r.ExpectedValues,
	}
}

func (r *Response) FromContract(from contract.Response) error {
	r.Code = from.Code
	r.Description = from.Description
	r.ExpectedValues = from.ExpectedValues
	return nil
}
