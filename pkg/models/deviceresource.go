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
 *******************************************************************************/

package models

import (
	"encoding/json"
)

type DeviceResource struct {
	Description string                 `json:"description"`
	Name        string                 `json:"name"`
	Tag         string                 `json:"tag"`
	Properties  ProfileProperty        `json:"properties" yaml:"properties"`
	Attributes  map[string]interface{} `json:"attributes" yaml:"attributes"`
}

// Custom marshaling to make empty strings null
func (do DeviceResource) MarshalJSON() ([]byte, error) {
	test := struct {
		Description *string                `json:"description"`
		Name        *string                `json:"name"`
		Tag         *string                `json:"tag"`
		Properties  ProfileProperty        `json:"properties"`
		Attributes  map[string]interface{} `json:"attributes"`
	}{
		Properties: do.Properties,
	}

	// Empty strings are null
	if do.Description != "" {
		test.Description = &do.Description
	}
	if do.Name != "" {
		test.Name = &do.Name
	}
	if do.Tag != "" {
		test.Tag = &do.Tag
	}

	// Empty maps are null
	if len(do.Attributes) > 0 {
		test.Attributes = do.Attributes
	}

	return json.Marshal(test)
}

/*
 * To String function for DeviceResource
 */
func (do DeviceResource) String() string {
	out, err := json.Marshal(do)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
