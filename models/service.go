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

type Service struct {
	DescribedObject			`bson:",inline"`
	Id		bson.ObjectId	`bson:"_id,omitempty" json:"id"`
	Name 		string		`bson:"name" json:"name"`			// time in milliseconds that the device last provided any feedback or responded to any request
	LastConnected 	int64 		`bson:"lastConnected" json:"lastConnected"`	// time in milliseconds that the device last reported data to the core
	LastReported 	int64		`bson:"lastReported" json:"lastReported"`	// operational state - either enabled or disabled
	OperatingState 	OperatingState	`bson:"operatingState" json:"operatingState"`	// operational state - ether enabled or disableddc
	Labels 		[]string	`bson:"labels" json:"labels"`			// tags or other labels applied to the device service for search or other identification needs
	Addressable 	Addressable	`bson:"addressable" json:"addressable"`		// address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
}

// Custom Marshaling to make empty strings null
func (s Service) MarshalJSON()([]byte, error){
	test := struct{
		DescribedObject			`bson:",inline"`
		Id		*bson.ObjectId	`bson:"_id,omitempty" json:"id"`
		Name 		*string		`bson:"name" json:"name"`			// time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected 	int64 		`bson:"lastConnected" json:"lastConnected"`	// time in milliseconds that the device last reported data to the core
		LastReported 	int64		`bson:"lastReported" json:"lastReported"`	// operational state - either enabled or disabled
		OperatingState 	OperatingState	`bson:"operatingState" json:"operatingState"`	// operational state - ether enabled or disableddc
		Labels 		[]string	`bson:"labels" json:"labels"`			// tags or other labels applied to the device service for search or other identification needs
		Addressable 	Addressable	`bson:"addressable" json:"addressable"`		// address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
	}{
		DescribedObject : s.DescribedObject,
		LastConnected : s.LastConnected,
		LastReported : s.LastReported,
		OperatingState : s.OperatingState,
		Labels : s.Labels,
		Addressable : s.Addressable,
	}

	// Empty strings are null
	if s.Name != "" {test.Name = &s.Name}
	if s.Id != "" {test.Id = &s.Id}
	
	return json.Marshal(test)
}

/*
 * To String function for Service
 */
func (dp Service) String() string {
	out, err := json.Marshal(dp)
	if err != nil {
		return err.Error()
	}

	return string(out)
}
