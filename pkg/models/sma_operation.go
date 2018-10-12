/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *******************************************************************************/

package models

import (
	"encoding/json"
)

/*
 * An Operation for SMA processing.
 *
 *
 * Operation struct
 */
type Operation struct {
	BaseObject        `bson:",inline"`
	Action   string   `bson:"action" json:"action,omitempty"`
	Params   []string `bson:"params" json:"params,omitempty"`
	Services []string `bson:"services,omitempty" json:"services,omitempty"`
}

// Custom marshaling to make empty strings null
func (o Operation) MarshalJSON() ([]byte, error) {
	test := struct {
		BaseObject
		Action   *string        `json:"action,omitempty"`
		Params   []string       `json:"params,omitempty"`
		Services []string       `json:"services,omitempty"`
	}{
		BaseObject: o.BaseObject,
	}

	if o.Action != "" {
		test.Action = &o.Action
	}
	return json.Marshal(test)
}

/*
 * To String function for Operation struct
 */
func (o Operation) String() string {
	out, err := json.Marshal(o)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
