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

var TestScheduleID1 = "one"
var TestScheduleID2 = "two"
var TestScheduleName = "test schedule"
var TestStart = "" // defaults to now
var TestEnd = ""   // defaults to ZDT MAX
var TestTime2015 = "20150101T000000"
var TestFreq2S = "PT2S"
var TestCRON = "0 0 12 * * ?"
var TestRunOnce = true
var TestSchedule = Schedule{BaseObject: TestBaseObject, Name: TestScheduleName, Start: TestStart, End: TestEnd, Frequency: TestFreq2S, Cron: TestCRON, RunOnce: TestRunOnce}

func TestSchedule_MarshalJSON(t *testing.T) {
	var testScheduleBytes = []byte(TestSchedule.String())

	tests := []struct {
		name    string
		s       Schedule
		want    []byte
		wantErr bool
	}{
		{"successful marshalling", TestSchedule, testScheduleBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Schedule.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Schedule.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSchedule_String(t *testing.T) {
	tests := []struct {
		name string
		dp   Schedule
		want string
	}{
		{"schedule to string", TestSchedule,
			"{\"created\":" + strconv.FormatInt(TestSchedule.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestSchedule.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestSchedule.Origin, 10) +
				",\"id\":\"\"" +
				",\"name\":\"" + TestScheduleName + "\"" +
				",\"start\":null" +
				",\"end\":null" +
				",\"frequency\":\"" + TestSchedule.Frequency + "\"" +
				",\"cron\":\"" + TestSchedule.Cron + "\"" +
				",\"runOnce\":" + strconv.FormatBool(TestSchedule.RunOnce) + "}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dp.String(); got != tt.want {
				t.Errorf("Schedule.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
