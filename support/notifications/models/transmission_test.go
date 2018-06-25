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
	"reflect"
	"testing"
)

var TestEmptyTransmission = Transmission{}
var TestTransmission = Transmission{BaseObject: TestBaseObject, Notification: TestNotification, Receiver: "test receiver",
	Channel: Channel{Type: ChannelType(Email), MailAddresses: []string{"jpwhite_mn@yahoo.com", "james_white2@dell.com"}}, Status: TransmissionStatus(Sent), ResendCount: 0, Records: []TransmissionRecord{TestSentTransRecord}}

func TestTransmission_MarshalJSON(t *testing.T) {
	var emptyBytes = []byte(TestEmptyTransmission.String())
	var tranBytes = []byte(TestTransmission.String())
	tests := []struct {
		name    string
		trans   *Transmission
		want    []byte
		wantErr bool
	}{
		{"test empty transmission", &TestEmptyTransmission, emptyBytes, false},
		{"test transmission", &TestTransmission, tranBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.trans.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Transmission.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transmission.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransmission_String(t *testing.T) {
	tests := []struct {
		name  string
		trans *Transmission
		want  string
	}{
		{"test string of empty transmission", &TestEmptyTransmission, "{\"created\":0,\"modified\":0,\"id\":null,\"notification\":{\"created\":0,\"modified\":0,\"id\":null},\"channel\":{},\"resendcount\":0}"},
		{"test string of transmission", &TestTransmission, "{\"created\":123,\"modified\":123,\"id\":null,\"notification\":{\"created\":123,\"modified\":123,\"id\":null,\"slug\":\"test slug\",\"sender\":\"test sender\",\"category\":\"SECURITY\",\"severity\":\"CRITICAL\",\"content\":\"test content\",\"description\":\"test description\",\"status\":\"NEW\",\"labels\":[\"label1\",\"labe2\"]},\"receiver\":\"test receiver\",\"channel\":{\"channeltype\":\"EMAIL\",\"mailaddresses\":[\"jpwhite_mn@yahoo.com\",\"james_white2@dell.com\"]},\"status\":\"SENT\",\"resendcount\":0,\"records\":[{\"status\":\"SENT\",\"response\":\"ok\",\"sent\":123}]}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.trans.String(); got != tt.want {
				t.Errorf("Transmission.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
