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

import "encoding/json"

type ProfileResource struct {
	Name 	string 			`bson:"name" json:"name"`
	Get	[]ResourceOperation	`bson:"get" json:"get"`
	Set	[]ResourceOperation	`bson:"set" json:"set"`
}

// Custom marshaling to make empty strings null
func (pr ProfileResource) MarshalJSON()([]byte, error){
	test := struct{
		Name 	*string 			`json:"name"`
		Get	[]ResourceOperation	`json:"get"`
		Set	[]ResourceOperation	`json:"set"`
	}{}
	
	// Empty strings are null
	if pr.Name != "" {test.Name = &pr.Name}
	
	// Empty arrays are null
	if len(pr.Get) > 0 {test.Get = pr.Get}
	if len(pr.Set) > 0 {test.Set = pr.Set}
	
	return json.Marshal(test)
}

/*
 * To String function for DeviceService
 */
func (pr ProfileResource) String() string {
	out, err := json.Marshal(pr)
	if err != nil {
		return err.Error()
	}
	return string(out)
}