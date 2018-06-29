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

type TransmissionRecord struct {
	Status   TransmissionStatus `bson:"status" json:"status"`
	Response string             `bson:"response" json:"response"`
	Sent     int64              `bson:"sent" json:"sent"`
}

// Custom marshaling to make empty strings null
func (t TransmissionRecord) MarshalJSON() ([]byte, error) {
	test := struct {
		Status   TransmissionStatus `json:"status"`
		Response *string            `json:"response"`
		Sent     int64              `json:"sent"`
	}{
		Status: t.Status,
		Sent:   t.Sent,
	}
	// Empty strings are null
	if t.Response != "" {
		test.Response = &t.Response
	}
	return json.Marshal(test)
}

/*
 * To String function for TransmissionRecord Struct
 */
func (t TransmissionRecord) String() string {
	out, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
