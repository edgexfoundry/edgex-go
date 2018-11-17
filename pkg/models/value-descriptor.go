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

	"gopkg.in/mgo.v2/bson"
)

/*
 * Value Descriptor Struct
 */
type ValueDescriptor struct {
	Id           bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Created      int64         `bson:"created" json:"created"`
	Description  string        `bson:"description" json:"description"`
	Modified     int64         `bson:"modified" json:"modified"`
	Origin       int64         `bson:"origin" json:"origin"`
	Name         string        `bson:"name" json:"name"`
	Min          interface{}   `bson:"min,omitempty" json:"min"`
	Max          interface{}   `bson:"max,omitempty" json:"max"`
	DefaultValue interface{}   `bson:"defaultValue,omitempty" json:"defaultValue"`
	Type         string        `bson:"type" json:"type"`
	UomLabel     string        `bson:"uomLabel,omitempty" json:"uomLabel"`
	Formatting   string        `bson:"formatting,omitempty" json:"formatting"`
	Labels       []string      `bson:"labels,omitempty" json:"labels"`
}

// Custom marshaling to make empty strings null
func (v ValueDescriptor) MarshalJSON() ([]byte, error) {
	test := struct {
		Id           bson.ObjectId `json:"id" bson:"_id,omitempty"`
		Created      int64         `json:"created,omitempty"`
		Description  *string       `json:"description,omitempty"`
		Modified     int64         `json:"modified,omitempty"`
		Origin       int64         `json:"origin,omitempty"`
		Name         *string       `json:"name,omitempty"`
		Min          *interface{}  `json:"min,omitempty"`
		Max          *interface{}  `json:"max,omitempty"`
		DefaultValue *interface{}  `json:"defaultValue,omitempty"`
		Type         *string       `json:"type,omitempty"`
		UomLabel     *string       `json:"uomLabel,omitempty"`
		Formatting   *string       `json:"formatting,omitempty"`
		Labels       []string      `json:"labels,omitempty"`
	}{
		Id:       v.Id,
		Created:  v.Created,
		Modified: v.Modified,
		Origin:   v.Origin,
		Labels:   v.Labels,
	}

	// Empty strings are null
	if v.Name != "" {
		test.Name = &v.Name
	}
	if v.Description != "" {
		test.Description = &v.Description
	}
	if v.Min != "" {
		test.Min = &v.Min
	}
	if v.Max != "" {
		test.Max = &v.Max
	}
	if v.Min != "" {
		test.DefaultValue = &v.DefaultValue
	}
	if v.Type != "" {
		test.Type = &v.Type
	}
	if v.UomLabel != "" {
		test.UomLabel = &v.UomLabel
	}
	if v.Formatting != "" {
		test.Formatting = &v.Formatting
	}

	return json.Marshal(test)
}

/*
 * To String function for ValueDescriptor Struct
 */
func (a ValueDescriptor) String() string {
	out, err := json.Marshal(a)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
