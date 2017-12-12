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

type ScheduleEvent struct {
	BaseObject			`bson:",inline"`
	Id		bson.ObjectId	`bson:"_id,omitempty" json:"id"`
	Name        string        	`bson:"name" json:"name"`		// non-database unique identifier for a schedule event
	Schedule    string        	`bson:"schedule" json:"schedule"`	// Name to associated owning schedule
	Addressable Addressable   	`bson:"addressable" json:"addressable"`	// address {MQTT topic, HTTP address, serial bus, etc.} for the action (can be empty)
	Parameters  string        	`bson:"parameters" json:"parameters"`	// json body for parameters
	Service     string        	`bson:"service" json:"service"`		// json body for parameters
}

// Custom marshaling to make empty strings null
func (se ScheduleEvent) MarshalJSON()([]byte, error){
	test := struct{
		BaseObject
		Id		bson.ObjectId	`json:"id"`
		Name        *string        	`json:"name"`		// non-database unique identifier for a schedule event
		Schedule    *string        	`json:"schedule"`	// Name to associated owning schedule
		Addressable Addressable   	`json:"addressable"`	// address {MQTT topic, HTTP address, serial bus, etc.} for the action (can be empty)
		Parameters  *string        	`json:"parameters"`	// json body for parameters
		Service     *string        	`json:"service"`		// json body for parameters
	}{
		Id : se.Id,
		BaseObject : se.BaseObject,
		Addressable : se.Addressable,
	}
	
	// Empty strings are null
	if se.Name != "" {test.Name = &se.Name}
	if se.Schedule != "" {test.Schedule = &se.Schedule}
	if se.Parameters != "" {test.Parameters = &se.Parameters}
	//if se.Service != "" {test.Service = &se.Service}
	
	return json.Marshal(test)
}

/*
 * To String function for ScheduleEvent
 */
func (se ScheduleEvent) String() string {
	out, err := json.Marshal(se)
	if err != nil {
		return err.Error()
	}
	return string(out)
}