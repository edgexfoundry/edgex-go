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

type DeviceService struct {
	Service Service				`bson:",inline"`
	AdminState 	AdminState	`bson:"adminState" json:"adminState"` // Device Service Admin State
}

// Custom Marshaling to make empty strings null
func (ds DeviceService) MarshalJSON()([]byte, error){
	
	test := struct{
		DescribedObject	`json:",inline"`
		Id		*bson.ObjectId	`json:"id"`
		Name 		*string		`json:"name"`	// time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected 	int64	`json:"lastConnected"`	// time in milliseconds that the device last reported data to the core
		LastReported 	int64	`json:"lastReported"`	// operational state - either enabled or disabled
		OperatingState 	OperatingState	`json:"operatingState"`	// operational state - ether enabled or disableddc
		Labels 		[]string 	`json:"labels"`// tags or other labels applied to the device service for search or other identification needs
		Addressable 	Addressable		`json:"addressable"`	// address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service			
		AdminState 	AdminState 		`json:"adminState"`// Device Service Admin State
	}{
		DescribedObject : ds.Service.DescribedObject,
		LastConnected : ds.Service.LastConnected,
		LastReported : ds.Service.LastReported,
		OperatingState : ds.Service.OperatingState,
		Labels : ds.Service.Labels,
		Addressable : ds.Service.Addressable,
		AdminState : ds.AdminState,
	}

	if ds.Service.Id != "" {
		test.Id = &ds.Service.Id
	}
	
	// Empty strings are null
	if ds.Service.Name != "" {test.Name = &ds.Service.Name}
	
	return json.Marshal(test)
}

// Custom unmarshaling funcion
func (ds *DeviceService) UnmarshalJSON(data []byte) error{
	type Alias struct{
		DescribedObject	`json:",inline"`
		Id		bson.ObjectId	`json:"id"`
		Name 		*string		`json:"name"`	// time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected 	int64	`json:"lastConnected"`	// time in milliseconds that the device last reported data to the core
		LastReported 	int64	`json:"lastReported"`	// operational state - either enabled or disabled
		OperatingState 	OperatingState	`json:"operatingState"`	// operational state - ether enabled or disableddc
		Labels 		[]string 	`json:"labels"`// tags or other labels applied to the device service for search or other identification needs
		Addressable 	Addressable		`json:"addressable"`	// address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service			
		AdminState 	AdminState 		`json:"adminState"`// Device Service Admin State
	}
	a := Alias{}
	
	// Error with unmarshaling
	if err := json.Unmarshal(data, &a); err != nil{
		return err
	}
	
	// Set the fields
	ds.AdminState = a.AdminState
	ds.Service.DescribedObject = a.DescribedObject
	ds.Service.LastConnected = a.LastConnected
	ds.Service.LastReported = a.LastReported
	ds.Service.OperatingState = a.OperatingState
	ds.Service.Labels = a.Labels
	ds.Service.Addressable = a.Addressable
	ds.Service.Id = a.Id
	
	// Name can be nil
	if a.Name != nil{
		ds.Service.Name = *a.Name
	}
	
	return nil
}

/*
 * To String function for DeviceService
 */
func (ds DeviceService) String() string {
	out, err := json.Marshal(ds)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
