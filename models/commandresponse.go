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
 * This file is the model for the CommandResponse object in EdgeX
 *
 *
 * CommandResponse struct
 */
type CommandResponse struct {
	Id             bson.ObjectId  `json:"id"`
	Name           string         `json:"name"`           // Unique name for identifying a device
	AdminState     AdminState     `json:"adminState"`     // Admin state (locked/unlocked)
	OperatingState OperatingState `json:"operatingState"` // Operating state (enabled/disabled)
	LastConnected  int64          `json:"lastConnected"`  // Time (milliseconds) that the device last provided any feedback or responded to any request
	LastReported   int64          `json:"lastReported"`   // Time (milliseconds) that the device reported data to the core microservice
	Labels         []string       `json:"labels"`         // Other labels applied to the device to help with searching
	Location       interface{}    `json:"location"`       // Device service specific location (interface{} is an empty interface so it can be anything)
	Commands       []Command      `json:"commands"`       // Associated Device Profile - Describes the device
}

// Custom marshaling to make empty strings null
func (cr CommandResponse) MarshalJSON() ([]byte, error) {
	res := struct {
		Id             *bson.ObjectId `json:"id"`
		Name           *string        `json:"name"`
		AdminState     AdminState     `json:"adminState"`
		OperatingState OperatingState `json:"operatingState"`
		LastConnected  int64          `json:"lastConnected"`
		LastReported   int64          `json:"lastReported"`
		Labels         []string       `json:"labels"`
		Location       interface{}    `json:"location"`
		Commands       []Command      `json:"commands"`
	}{
		AdminState:     cr.AdminState,
		OperatingState: cr.OperatingState,
		LastConnected:  cr.LastConnected,
		LastReported:   cr.LastReported,
		Labels:         cr.Labels,
		Location:       cr.Location,
		Commands:       cr.Commands,
	}

	if cr.Id != "" {
		res.Id = &cr.Id
	}

	// Empty strings are null
	if cr.Name != "" {
		res.Name = &cr.Name
	}

	return json.Marshal(res)
}

/*
 * String function for representing a device
 */
func (d CommandResponse) String() string {
	out, err := json.Marshal(d)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

func createUrl(basePath string, cmdId string) string {
	return basePath + cmdId
}

/*
 * Create a Command Response struct from a Device Struct
 */
func CommandResponseFromDevice(d Device, cmdURL string) CommandResponse {
	cmdResp := CommandResponse{
		Id:             d.Id,
		Name:           d.Name,
		AdminState:     d.AdminState,
		OperatingState: d.OperatingState,
		LastConnected:  d.LastConnected,
		LastReported:   d.LastReported,
		Labels:         d.Labels,
		Location:       d.Location,
		Commands:       d.Profile.Commands,
	}

	urlPath := "/api/v1/device/"
	urlCmdPath := "/command/"
	basePath := cmdURL + urlPath + d.Id.Hex() + urlCmdPath

	for _, c := range cmdResp.Commands {
		if c.Get != nil {
			c.Get.URL = createUrl(basePath, c.Id.Hex())
		}
		if c.Put != nil {
			c.Put.URL = createUrl(basePath, c.Id.Hex())
		}
	}

	return cmdResp
}
