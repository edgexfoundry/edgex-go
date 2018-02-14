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

const testAddrName = "TEST_ADDR.NAME"
const testProtocol = "HTTP"
const testMethod = "Get"
const testAddress = "localhost"
const testPort = 48089
const testAddressablePath = "/api/v1/device"
const testPublisher = "TEST_PUB"
const testUser = "edgexer"
const testPassword = "password"
const testTopic = "device_topic"

var TestAddressable = Addressable{BaseObject: TestBaseObject, Name: testAddrName, Protocol: testProtocol, HTTPMethod: testMethod, Address: testAddress, Port: testPort, Path: testAddressablePath, Publisher: testPublisher, User: testUser, Password: testPassword, Topic: testTopic}
var EmptyAddressable = Addressable{}

func TestAddressable_MarshalJSON(t *testing.T) {
	var resultTestBytes = []byte(TestAddressable.String())
	var resultEmptyBytes = []byte(EmptyAddressable.String())
	tests := []struct {
		name    string
		a       Addressable
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestAddressable, resultTestBytes, false},
		{"successful empty marshal", EmptyAddressable, resultEmptyBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Addressable.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Addressable.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddressable_String(t *testing.T) {
	tests := []struct {
		name string
		a    Addressable
		want string
	}{
		{"full addressable", TestAddressable, "{\"created\":" + strconv.FormatInt(TestAddressable.Created, 10) +
			",\"modified\":" + strconv.FormatInt(TestAddressable.Modified, 10) +
			",\"origin\":" + strconv.FormatInt(TestAddressable.Origin, 10) +
			",\"id\":null,\"name\":\"" + TestAddressable.Name +
			"\",\"protocol\":\"" + TestAddressable.Protocol +
			"\",\"method\":\"" + TestAddressable.HTTPMethod +
			"\",\"address\":\"" + TestAddressable.Address +
			"\",\"port\":" + strconv.Itoa(TestAddressable.Port) +
			",\"path\":\"" + TestAddressable.Path +
			"\",\"publisher\":\"" + TestAddressable.Publisher +
			"\",\"user\":\"" + TestAddressable.User +
			"\",\"password\":\"" + TestAddressable.Password +
			"\",\"topic\":\"" + TestAddressable.Topic +
			"\",\"baseURL\":\"" + TestAddressable.Protocol + "://" + TestAddressable.Address + ":" + strconv.Itoa(TestAddressable.Port) +
			"\",\"url\":\"" + TestAddressable.Protocol + "://" + TestAddressable.Address + ":" + strconv.Itoa(TestAddressable.Port) + TestAddressable.Path + "\"}"},
		{"empty", EmptyAddressable, "{\"created\":0,\"modified\":0,\"origin\":0,\"id\":null,\"name\":null,\"protocol\":null,\"method\":null,\"address\":null,\"port\":0,\"path\":null,\"publisher\":null,\"user\":null,\"password\":null,\"topic\":null,\"baseURL\":null,\"url\":null}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.String(); got != tt.want {
				t.Errorf("Addressable.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
