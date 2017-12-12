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
)

type DeviceObject struct {
//	DescribedObject				`bson:",inline" yaml:",inline"`
//	Id		bson.ObjectId		`bson:"_id,omitempty" json:"id"`
	Description	string	`bson:"description" json:"description"`
	Name 		string			`bson:"name" json:"name"`
	Tag		string			`bson:"tag" json:"tag"`
//	Properties 	ProfileProperty 	`bson:"profileProperty" json:"profileProperty"`
	Properties 	ProfileProperty 	`bson:"properties" json:"properties" yaml:"properties"`
	Attributes  map[string]string   `bson:"attributes" json:"attributes" yaml:"attributes"`
//	Other 		string	`bson:"other" json:"other"`
//	Other 		map[string]string	`bson:"other" json:"other"`
}

// Custom marshaling to make empty strings null
func (do DeviceObject)MarshalJSON()([]byte, error){
	test := struct{
		Description	*string	`json:"description"`
		Name 		*string			`json:"name"`
		Tag		*string			`json:"tag"`
		Properties 	ProfileProperty 	`json:"properties"`
		Attributes  map[string]string   `json:"attributes"`
	}{
		Properties : do.Properties,
	}
	
	// Empty strings are null
	if do.Description != "" {test.Description = &do.Description}
	if do.Name != "" {test.Name = &do.Name}
	if do.Tag != "" {test.Tag = &do.Tag}
	
	// Empty maps are null
	if len(do.Attributes) > 0 {test.Attributes = do.Attributes}
	
	return json.Marshal(test)
}

/*
 * To String function for DeviceObject
 */
func (do DeviceObject) String() string {
	out, err := json.Marshal(do)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
