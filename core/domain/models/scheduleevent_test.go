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

var TestScheduleEventName = "test schedule event"
var TestParams = "{key1:value1}"
var TestScheduleEvent = ScheduleEvent{BaseObject: TestBaseObject, Name: TestScheduleEventName, Schedule: TestSchedule.Name, Addressable: TestAddressable, Parameters: TestParams, Service: TestServiceName}

func TestScheduleEvent_MarshalJSON(t *testing.T) {
	var emptyScheduleEvent = ScheduleEvent{}
	var resultTestBytes = []byte(TestScheduleEvent.String())
	var resultEmptyTestBytes = []byte(emptyScheduleEvent.String())

	tests := []struct {
		name    string
		se      ScheduleEvent
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestScheduleEvent, resultTestBytes, false},
		{"successful empty marshal", emptyScheduleEvent, resultEmptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.se.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScheduleEvent.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScheduleEvent.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduleEvent_String(t *testing.T) {
	tests := []struct {
		name string
		se   ScheduleEvent
		want string
	}{
		{"schedule event to string", TestScheduleEvent,
			"{\"created\":" + strconv.FormatInt(TestBaseObject.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestBaseObject.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestBaseObject.Origin, 10) +
				",\"id\":\"\"" +
				",\"name\":\"" + TestScheduleEventName + "\"" +
				",\"schedule\":\"" + TestScheduleName + "\"" +
				",\"addressable\":" + TestAddressable.String() +
				",\"parameters\":\"" + TestParams + "\"" +
				",\"service\":\"" + TestServiceName + "\"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.se.String(); got != tt.want {
				t.Errorf("ScheduleEvent.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
