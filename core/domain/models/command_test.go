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

var TestCommandName = "test command name"
var TestCommand = Command{BaseObject: TestBaseObject, Name: TestCommandName, Get: &TestGet, Put: &TestPut}

func TestCommand_MarshalJSON(t *testing.T) {
	var testCommandBytes = []byte(TestCommand.String())
	tests := []struct {
		name    string
		c       Command
		want    []byte
		wantErr bool
	}{
		{"successful marshalling", TestCommand, testCommandBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Command.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_String(t *testing.T) {
	tests := []struct {
		name string
		c    Command
		want string
	}{
		{"command to string", TestCommand,
			"{\"created\":" + strconv.FormatInt(TestCommand.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestCommand.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestCommand.Origin, 10) +
				",\"id\":null,\"name\":\"" + TestCommand.Name + "\"" +
				",\"get\":" + TestGet.String() +
				",\"put\":" + TestPut.String() + "}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("Command.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_AllAssociatedValueDescriptors(t *testing.T) {
	var testMap = make(map[string]string)
	type args struct {
		vdNames *map[string]string
	}
	tests := []struct {
		name string
		c    *Command
		args args
	}{
		{"get assoc val descs", &TestCommand, args{vdNames: &testMap}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.AllAssociatedValueDescriptors(tt.args.vdNames)
			if len(*tt.args.vdNames) != 2 {
				t.Error("Associated value descriptor size > than expected")
			}
		})
	}
}
