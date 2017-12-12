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

type Units struct {
	Type 		string 	`bson:"type" json:"type"`
	ReadWrite 	string 	`bson:"readWrite" json:"readWrite" yaml:"readWrite"`
	DefaultValue 	string 	`bson:"defaultValue" json:"defaultValue" yaml:"defaultValue"`
}

// Custom marshaling to make empty strings null
func (u Units) MarshalJSON()([]byte, error){
	test := struct{
		Type 		*string 	`json:"type"`
		ReadWrite		*string 	`json:"readWrite"`
		DefaultValue 	*string 	`json:"defaultValue"`
	}{}
	
	// Empty strings are null
	if u.Type != "" {test.Type = &u.Type}
	if u.ReadWrite != "" {test.ReadWrite = &u.ReadWrite}
	if u.DefaultValue != "" {test.DefaultValue = &u.DefaultValue}
	
	return json.Marshal(test)
}

/*
 * To String function for Units
 */
func (u Units) String() string {
	out, err := json.Marshal(u)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
