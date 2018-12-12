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

import (
	"encoding/json"
	"strconv"
	"strings"
)

type IntervalAction struct {
	ID         string `json:"id"`
	Created    int64  `json:"created"`
	Modified   int64  `json:"modified"`
	Origin     int64  `json:"origin"`
	Name       string `json:"name"`
	Interval   string `json:"interval"`
	Parameters string `json:"parameters"`
	Target     string `json:"target"`
	Protocol   string `json:"protocol"`
	HTTPMethod string `json:"httpMethod"`
	Address    string `json:"address"`
	Port       int    `json:"port"`
	Path       string `json:"path"`
	Publisher  string `json:"publisher"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Topic      string `json:"topic"`
}

func (ia IntervalAction) MarshalJSON() ([]byte, error) {
	test := struct {
		ID         *string `json:"id,omitempty"`
		Created    int64   `json:"created,omitempty"`
		Modified   int64   `json:"modified,omitempty"`
		Origin     int64   `json:"origin,omitempty"`
		Name       *string `json:"name,omitempty"`
		Interval   *string `json:"interval,omitempty"`
		Parameters *string `json:"parameters,omitempty"`
		Target     *string `json:"target,omitempty"`
		Protocol   *string `json:"protocol,omitempty"`
		HTTPMethod *string `json:"httpMethod,omitempty"`
		Address    *string `json:"address,omitempty"`
		Port       int     `json:"port,omitempty"`
		Path       *string `json:"path,omitempty"`
		Publisher  *string `json:"publisher,omitempty"`
		User       *string `json:"user,omitempty"`
		Password   *string `json:"password,omitempty"`
		Topic      *string `json:"topic,omitempty"`
	}{
		Created:  ia.Created,
		Modified: ia.Modified,
		Origin:   ia.Origin,
		Port:     ia.Port,
	}

	// Empty strings are null
	if ia.ID != "" {
		test.ID = &ia.ID
	}
	if ia.Name != "" {
		test.Name = &ia.Name
	}
	if ia.Interval != "" {
		test.Interval = &ia.Interval
	}
	if ia.Parameters != "" {
		test.Parameters = &ia.Parameters
	}
	if ia.Target != "" {
		test.Target = &ia.Target
	}
	if ia.Protocol != "" {
		test.Protocol = &ia.Protocol
	}
	if ia.HTTPMethod != "" {
		test.HTTPMethod = &ia.HTTPMethod
	}
	if ia.Address != "" {
		test.Address = &ia.Address
	}
	if ia.Publisher != "" {
		test.Publisher = &ia.Publisher
	}
	if ia.User != "" {
		test.User = &ia.User
	}
	if ia.Password != "" {
		test.Password = &ia.Password
	}
	if ia.Topic != "" {
		test.Topic = &ia.Topic
	}

	return json.Marshal(test)
}

func (ia IntervalAction) String() string {
	out, err := json.Marshal(ia)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

func (ia IntervalAction) GetBaseURL() string {
	protocol := strings.ToLower(ia.Protocol)
	address := ia.Address
	port := strconv.Itoa(ia.Port)
	baseUrl := protocol + "://" + address + ":" + port
	return baseUrl
}
