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
 *This file is the model for a device profile in EdgeX
 *
 *
 * Device profile struct
 */
type DeviceProfile struct {
	DescribedObject					`bson:",inline" yaml:",inline"`
	Id 		bson.ObjectId			`bson:"_id,omitempty" json:"id"`
	Name 		string 				`bson:"name" json:"name" yaml:"name"`				// Non-database identifier (must be unique)
	Manufacturer 	string 				`bson:"manufacturer" json:"manufacturer" yaml:"manufacturer"`	// Manufacturer of the device
	Model 		string 				`bson:"model" json:"model" yaml:"model"`			// Model of the device
	Labels 		[]string 			`bson:"labels" json:"labels" yaml:"labels,flow"`			// Labels used to search for groups of profiles
	Objects 	interface{}	 		`bson:"objects" json:"objects" yaml:"objects"`			// JSON data that the device service uses to communicate with devices with this profile
	DeviceResources []DeviceObject			`bson:"deviceResources" json:"deviceResources" yaml:"deviceResources"`
	Resources	[]ProfileResource		`bson:"resources" json:"resources" yaml:"resources"`
	Commands 	[]Command			`bson:"commands" json:"commands" yaml:"commands"`		// List of commands to Get/Put information for devices associated with this profile
}

// Custom marshaling so that empty strings and arrays are null
func (dp DeviceProfile) MarshalJSON()([]byte, error){
	test := struct{
		DescribedObject
		Id 		bson.ObjectId			`json:"id"`
		Name 		*string 				`json:"name"`				// Non-database identifier (must be unique)
		Manufacturer 	*string 				`json:"manufacturer"`	// Manufacturer of the device
		Model 		*string 				`json:"model"`			// Model of the device
		Labels 		[]string 			`json:"labels"`			// Labels used to search for groups of profiles
		Objects 	interface{}	 		`json:"objects"`			// JSON data that the device service uses to communicate with devices with this profile
		DeviceResources []DeviceObject			`json:"deviceResources"`
		Resources	[]ProfileResource		`json:"resources"`
		Commands 	[]Command			`json:"commands"`		// List of commands to Get/Put information for devices associated with this profile
	}{
		Id : dp.Id,
		Labels : dp.Labels,
		DescribedObject : dp.DescribedObject,
		Objects : dp.Objects,
	}
	
	// Empty strings are null
	if dp.Name != "" {test.Name = &dp.Name}
	if dp.Manufacturer != "" {test.Manufacturer = &dp.Manufacturer}
	if dp.Model != "" {test.Model = &dp.Model}
	
	// Empty arrays are null
	if len(dp.DeviceResources) > 0 {test.DeviceResources = dp.DeviceResources}
	if len(dp.Resources) > 0 {test.Resources = dp.Resources}
	if len(dp.Commands) > 0 {test.Commands = dp.Commands}
	
	return json.Marshal(test)
}

/*
 * To String function for DeviceProfile
 */
func (dp DeviceProfile) String() string {
	out, err := json.Marshal(dp)
	if err != nil {
		return err.Error()
	}
	return string(out)
}