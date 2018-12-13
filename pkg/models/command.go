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

/*
 * This file is the model for a command in EdgeX
 *
 *
 * Command struct
 */
type Command struct {
	BaseObject `yaml:",inline"`
	Id         string `json:"id"`
	Name       string `json:"name" yaml:"name"` // Command name (unique on the profile)
	Get        *Get   `json:"get" yaml:"get"`   // Get Command
	Put        *Put   `json:"put" yaml:"put"`   // Put Command
}

// Custom marshaling for making empty strings null
func (c Command) MarshalJSON() ([]byte, error) {
	test := struct {
		BaseObject
		Id   *string `json:"id"`
		Name *string `json:"name"` // Command name (unique on the profile)
		Get  *Get    `json:"get"`  // Get Command
		Put  *Put    `json:"put"`  // Put Command
	}{
		BaseObject: c.BaseObject,
		Get:        c.Get,
		Put:        c.Put,
	}

	if c.Id != "" {
		test.Id = &c.Id
	}

	// Make empty strings null
	if c.Name != "" {
		test.Name = &c.Name
	}

	return json.Marshal(test)
}

/*
 * String() function for formatting
 */
func (c Command) String() string {
	out, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

func (c *Command) UnmarshalJSON(b []byte) error {
	type Alias Command
	alias := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(b, &alias); err != nil {
		return err
	}
	c = (*Command)(alias.Alias)
	if c.Get == nil {
		c.Get = &Get{}
	}
	if c.Put == nil {
		c.Put = &Put{}
	}
	return nil
}

// Append all the associated value descriptors to the list
// Associated by PUT command parameters and PUT/GET command return values
func (c *Command) AllAssociatedValueDescriptors(vdNames *map[string]string) {
	// Check and add Get value descriptors
	if &(c.Get) != nil {
		c.Get.AllAssociatedValueDescriptors(vdNames)
	}

	// Check and add Put value descriptors
	if &(c.Put) != nil {
		c.Put.AllAssociatedValueDescriptors(vdNames)
	}
}
