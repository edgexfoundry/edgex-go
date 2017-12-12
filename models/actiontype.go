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
)

type ActionType string

const (
	PROFILE	ActionType = "PROFILE"
	DEVICE = "DEVICE"
	SERVICE = "SERVICE"
	MANAGER = "MANAGER"
	SCHEDULE = "SCHEDULE"
	SCHEDULEEVENT = "SCHEDULEEVENT"
	ADDRESSABLE = "ADDRESSABLE"
	VALUEDESCRIPTOR = "VALUEDESCRIPTOR"
	PROVISIONWATCHER = "PROVISIONWATCHER"
)

//func (at *ActionType) UnmarshalJSON(data []byte) error {
//	// Extract the string from data.
//	var s string
//	if err := json.Unmarshal(data, &s); err != nil {
//		return fmt.Errorf("ActionType should be a string, got %s", data)
//	}
//
//	got, err := map[string]ActionType{"PROFILE": PROFILE, "DEVICE": DEVICE, "SERVICE": SERVICE, "MANAGER": MANAGER, "SCHEDULE": SCHEDULE, "SCHEDULEEVENT": SCHEDULEEVENT, "ADDRESSABLE": ADDRESSABLE, "VALUEDESCRIPTOR": VALUEDESCRIPTOR, "PROVISIONWATCHER": PROVISIONWATCHER}[s]
//	if !err {
//		return fmt.Errorf("invalid ActionType %q", s)
//	}
//	*at = got
//	return nil
//}

//func (at *ActionType) MarshalText() ([]byte, error) {
//
//}