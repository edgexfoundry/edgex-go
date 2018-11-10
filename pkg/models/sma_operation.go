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
	Action   string   `bson:"action" json:"action,omitempty"`
	Services []string `bson:"services,omitempty" json:"services,omitempty"`
}

//Implements unmarshaling of JSON string to Operation type instance
func (o *Operation) UnmarshalJSON(data []byte) error {
	test := struct {
		Action   *string  `json:"action"`
		Services []string `json:"services"`
	}{}

	//Verify that incoming string will unmarshal successfully
	if err := json.Unmarshal(data, &test); err != nil {
		return err
	}

	//If so, copy the fields
	if test.Action != nil {
		o.Action = *test.Action
	}

	o.Services = []string{}
	if len(test.Services) > 0 {
		o.Services = test.Services
	}
	return nil
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
