/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

import "encoding/json"

type LogEntry struct {
	Level         string        `bson:"logLevel" json:"logLevel"`
	Args          []interface{} `bson:"args" json:"args"`
	OriginService string        `bson:"originService" json:"originService"`
	Message       string        `bson:"message" json:"message"`
	Created       int64         `bson:"created" json:"created"`
}

func (l LogEntry) MarshalJSON() ([]byte, error) {
	test := struct {
		Level         *string       `json:"logLevel,omitempty"`
		Args          []interface{} `json:"args,omitempty"`
		OriginService *string       `json:"originService,omitempty"`
		Message       *string       `json:"message,omitempty"`
		Created       int64         `json:"created,omitempty"`
	}{
		Created: l.Created,
	}

	// Empty strings are null
	if l.Level != "" {
		test.Level = &l.Level
	}

	if l.OriginService != "" {
		test.OriginService = &l.OriginService
	}

	if l.Message != "" {
		test.Message = &l.Message
	}

	if len(l.Args) > 0 {
		test.Args = l.Args
	}

	return json.Marshal(test)
}
