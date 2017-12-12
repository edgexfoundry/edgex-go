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

// OperatingState Constant String
type OperatingState string

/*
	Enabled  : ENABLED
	Disabled : DISABLED
	enabled  : TODO rename all ref to Enabled
	disabled : TODO rename all ref to Disabled
*/
const (
	Enabled  = "ENABLED"
	Disabled = "DISABLED"
	enabled  = "enabled"
	disabled = "disabled"
)

// UnmarshalJSON : Struct into json
func (os *OperatingState) UnmarshalJSON(data []byte) error {
	// Extract the string from data.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("OperatingState should be a string, got %s", data)
	}

	got, err := map[string]OperatingState{"ENABLED": Enabled, "DISABLED": Disabled}[s]
	if !err {
		return fmt.Errorf("invalid OperatingState %q", s)
	}
	*os = got
	return nil
}

// IsOperatingStateType : return if ostype
func IsOperatingStateType(os string) bool {
	_, err := map[string]OperatingState{"ENABLED": Enabled, "DISABLED": Disabled}[os]
	if !err {
		return false
	}
	return true
}
