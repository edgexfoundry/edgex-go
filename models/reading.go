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
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
)

/*
 * This file is for the Reading model in EdgeX
 * Holds data that was gathered from a device
 *
 *
 * Struct for the Reading object in EdgeX
 */
type Reading struct{
	Id bson.ObjectId 	`bson:"_id,omitempty"`
	Pushed int64	 	`bson:"pushed" json:"pushed"`		// When the data was pushed out of EdgeX (0 - not pushed yet)
	Created int64		`bson: "created" json:"created"`	// When the reading was created
	Origin int64		`bson:"origin" json:"origin"`
	Modified int64		`bson:"modified" json:"modified"`
	Device string		`bson:"device" json:"device"`
	Name string			`bson:"name" json:"name"`
	Value string		`bson:"value" json:"value"`		// Device sensor data value
}

// Custom marshaling to make empty strings null
func (r Reading) MarshalJSON()([]byte, error){
	test := struct{
		Id bson.ObjectId 	`json:"id"`
		Pushed int64	 	`json:"pushed"`		// When the data was pushed out of EdgeX (0 - not pushed yet)
		Created int64		`json:"created"`	// When the reading was created
		Origin int64		`json:"origin"`
		Modified int64		`json:"modified"`
		Device *string		`json:"device"`
		Name *string			`json:"name"`
		Value *string		`json:"value"`		// Device sensor data value
	}{
		Id : r.Id,
		Pushed : r.Pushed,
		Created : r.Created,
		Origin : r.Origin,
		Modified : r.Modified,
	}
	
	// Empty strings are null
	if r.Device != "" {test.Device = &r.Device}
	if r.Name != "" {test.Name = &r.Name}
	if r.Value != "" {test.Value = &r.Value}
	
	return json.Marshal(test)
}

/*
 * To String function for Reading Struct
 */
func (r Reading) String() string{
	out, err := json.Marshal(r)
	if err != nil{return err.Error()}
	
	return string(out)
}
