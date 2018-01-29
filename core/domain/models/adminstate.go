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
	"fmt"
)

// AdminState : unlocked or locked
type AdminState string

const (
	// Locked : device is locked
	// Unlocked : device is unlocked
	// locked : TODO rename all ref to Locked
	// unlocked : TODO rename all ref to Unlocked
	Locked   = "LOCKED"
	Unlocked = "UNLOCKED"
	locked   = "locked"
	unlocked = "unlocked"
)

/*
 *  Unmarshal the enum type
 */
func (as *AdminState) UnmarshalJSON(data []byte) error {
	// Extract the string from data.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AdminState should be a string, got %s", data)
	}

	got, err := map[string]AdminState{"LOCKED": Locked, "UNLOCKED": Unlocked}[s]
	if !err {
		return fmt.Errorf("invalid AdminState %q", s)
	}
	*as = got
	return nil
}
func IsAdminStateType(as string) bool {
	_, err := map[string]AdminState{"LOCKED": Locked, "UNLOCKED": Unlocked}[as]
	if !err {
		return false
	}
	return true
}
