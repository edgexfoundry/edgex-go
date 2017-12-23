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
 * @author: Jim White, Dell
 * @version: 0.5.0
 *******************************************************************************/

package models

import (
	"reflect"
	"strconv"
	"testing"
)

var TestEvent = Event{Pushed: 123, Created: 123, Origin: 123, Modified: 123, Readings: []Reading{TestReading}}

func TestEvent_MarshalJSON(t *testing.T) {
	var emptyEvent = Event{}
	var resultTestBytes = []byte(TestEvent.String())
	var resultEmptyTestBytes = []byte(emptyEvent.String())

	tests := []struct {
		name    string
		e       Event
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestEvent, resultTestBytes, false},
		{"successful empty marshal", emptyEvent, resultEmptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Event.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_String(t *testing.T) {
	tests := []struct {
		name string
		e    Event
		want string
	}{
		{"event to string", TestEvent,
			"{\"id\":\"\"" +
				",\"pushed\":" + strconv.FormatInt(TestEvent.Pushed, 10) +
				",\"device\":null" +
				",\"created\":" + strconv.FormatInt(TestEvent.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestEvent.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestEvent.Origin, 10) +
				",\"schedule\":null" +
				",\"event\":null" +
				",\"readings\":[" + TestReading.String() + "]" +
				"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("Event.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
