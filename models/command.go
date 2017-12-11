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
	"gopkg.in/mgo.v2/bson"
)

/*
 * This file is the model for a command in EdgeX
 *
 *
 * Command struct
 */
type Command struct {
	BaseObject			`bson:",inline" yaml:",inline"`
	Id	bson.ObjectId 		`bson:"_id,omitempty" json:"id"`
	Name 	string			`bson:"name" json:"name" yaml:"name"`	// Command name (unique on the profile)
	Get 	*Get			`bson:"get" json:"get" yaml:"get"`	// Get Command
	Put 	*Put			`bson:"put" json:"put" yaml:"put"`	// Put Command
}

// Custom marshaling for making empty strings null
func (c Command) MarshalJSON()([]byte, error){
	test := struct{
		BaseObject
		Id	*bson.ObjectId 		`json:"id"`
		Name 	*string			`json:"name"`	// Command name (unique on the profile)
		Get 	*Get			`json:"get"`	// Get Command
		Put 	*Put			`json:"put"`	// Put Command
	}{
		BaseObject : c.BaseObject,
		Get : c.Get,
		Put : c.Put,
	}

	if c.Id != "" {
		test.Id = &c.Id
	}
	
	// Make empty strings null
	if c.Name != ""{test.Name = &c.Name}
	
	return json.Marshal(test)
}

/*
 * String() function for formatting
 */
func (c Command) String() string {
	out, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// Append all the associated value descriptors to the list
// Associated by PUT command parameters and PUT/GET command return values
func (c *Command) AllAssociatedValueDescriptors (vdNames *map[string]string){
	// Check and add Get value descriptors
	if &(c.Get) != nil{
		c.Get.AllAssociatedValueDescriptors(vdNames)
	}
	
	// Check and add Put value descriptors
	if &(c.Put) != nil{
		c.Put.AllAssociatedValueDescriptors(vdNames)
	}
}
