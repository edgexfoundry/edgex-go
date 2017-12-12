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

type Schedule struct {
	BaseObject			`bson:",inline"`
	Id		bson.ObjectId	`bson:"_id,omitempty" json:"id"`
	Name 		string		`bson:"name" json:"name"`		// non-database identifier for a shcedule (*must be quitue)
	Start 		string		`bson:"start" json:"start"`		// Start time in ISO 8601 format YYYYMMDD'T'HHmmss 	@JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyymmdd'T'HHmmss")
	End 		string		`bson:"end" json:"end"`			// Start time in ISO 8601 format YYYYMMDD'T'HHmmss 	@JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyymmdd'T'HHmmss")
	Frequency	string 		`bson:"frequency" json:"frequency"` 	// how frequently should the event occur
	Cron 		string		`bson:"cron" json:"cron"`		// cron styled regular expression indicating how often the action under schedule should occur.  Use either runOnce, frequency or cron and not all.
	RunOnce		bool		`bson:"runOnce" json:"runOnce"`		// boolean indicating that this schedules runs one time - at the time indicated by the start
}

// Custom marshaling to make empty strings null
func (s Schedule) MarshalJSON()([]byte, error){
	test := struct{
		BaseObject
		Id		bson.ObjectId	`json:"id"`
		Name 		*string		`json:"name"`		// non-database identifier for a shcedule (*must be quitue)
		Start 		*string		`json:"start"`		// Start time in ISO 8601 format YYYYMMDD'T'HHmmss 	@JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyymmdd'T'HHmmss")
		End 		*string		`json:"end"`			// Start time in ISO 8601 format YYYYMMDD'T'HHmmss 	@JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyymmdd'T'HHmmss")
		Frequency	*string 		`json:"frequency"` 	// how frequently should the event occur
		Cron 		*string		`json:"-"`		// cron styled regular expression indicating how often the action under schedule should occur.  Use either runOnce, frequency or cron and not all.
		RunOnce		bool		`json:"-"`		// boolean indicating that this schedules runs one time - at the time indicated by the start
	}{
		Id : s.Id,
		BaseObject : s.BaseObject,
		RunOnce : s.RunOnce,
	}
	
	// Empty strings are null
	if s.Name != "" {test.Name = &s.Name}
	if s.Start != "" {test.Start = &s.Start}
	if s.End != "" {test.End = &s.End}
	if s.Frequency != "" {test.Frequency = &s.Frequency}
	if s.Cron != "" {test.Cron = &s.Cron}
	
	return json.Marshal(test)
}

/*
 * To String function for Schedule
 */
func (dp Schedule) String() string {
	out, err := json.Marshal(dp)
	if err != nil {
		return err.Error()
	}
	return string(out)
}