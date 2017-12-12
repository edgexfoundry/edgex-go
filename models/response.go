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
 * @microservice: core-domain-go library
 * @author: Ryan Comer & Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/

package models

import (
	"encoding/json"
	"strings"
	"reflect"
)
/*
 * Response for a Get or Put request to a service
 *
 *
 * Response Struct
 */
type Response struct {
	Code 		string			`bson:"code" json:"code" yaml:"code"`
	Description string			`bson:"description" json:"description" yaml:"description"`
	ExpectedValues 	[]string		`bson:"expectedValues" json:"expectedValues" yaml:"expectedValues"`
}

// Custom marshalling to make empty strings null
func (r Response) MarshalJSON()([]byte, error){
	test := struct{
		Code 		*string			`json:"code"`
		Description *string			`json:"description"`
		ExpectedValues 	[]string		`json:"expectedValues"`
	}{
		ExpectedValues : r.ExpectedValues,
	}
	
	// Empty strings are null
	if r.Code != "" {test.Code = &r.Code}
	if r.Description != "" {test.Description = &r.Description}
	
	return json.Marshal(test)
}

/*
 * To String function for Response Struct
 */
func (a Response) String() string {
	out, err := json.Marshal(a)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

func (r Response) Equals(r2 Response) bool {
	if strings.Compare(r.Code, r2.Code) != 0 {
		return false
	}
	if strings.Compare(r.Description, r2.Description) != 0 {
		return false
	}
	if len(r.ExpectedValues) != len (r2.ExpectedValues) {
		return false
	}
	if !reflect.DeepEqual(r.ExpectedValues, r2.ExpectedValues) {
		return false
	}
	return true

}