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

type DeviceReport struct {
	BaseObject			`bson:",inline"`
	Id		bson.ObjectId	`bson:"_id,omitempty" json:"id"`
	Name 		string 		`bson:"name" json:"name"`		// non-database identifier for a device report - must be unique
	Device 		string		`bson:"device" json:"device"`		// associated device name - should be a valid and unique device name
	Event 		string		`bson:"event" json:"event"`		// associated schedule event name - should be a valid and unique schedule event name
	Expected 	[]string	`bson:"expected" json:"expected"`	// array of value descriptor names describing the types of data captured in the report
}

// Custom marshaling to make empty strings null
func (dp DeviceReport)MarshalJSON()([]byte, error){
	test := struct{
		BaseObject
		Id		bson.ObjectId	`json:"id"`
		Name 		*string 		`json:"name"`		// non-database identifier for a device report - must be unique
		Device 		*string		`json:"device"`		// associated device name - should be a valid and unique device name
		Event 		*string		`json:"event"`		// associated schedule event name - should be a valid and unique schedule event name
		Expected 	[]string	`json:"expected"`	// array of value descriptor names describing the types of data captured in the report
	}{
		BaseObject : dp.BaseObject,
		Id : dp.Id,
		Expected : dp.Expected,
	}
	
	// Empty strings are null
	if dp.Name != "" {test.Name = &dp.Name}
	if dp.Device != "" {test.Device = &dp.Device}
	if dp.Event != "" {test.Event = &dp.Event}
	
	return json.Marshal(test)
}

/*
 * To String function for DeviceProfile
 */
func (dr DeviceReport) String() string {
	out, err := json.Marshal(dr)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
