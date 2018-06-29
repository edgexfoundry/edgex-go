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

import "testing"

func TestChannelType_UnmarshalJSON(t *testing.T) {
	var eChannelType = ChannelType(Email)
	var rChannelType = ChannelType(Rest)

	tests := []struct {
		name    string
		as      *ChannelType
		arg     []byte
		wantErr bool
	}{
		{"test unmarshal email", &eChannelType, []byte("\"EMAIL\""), false},
		{"test unmarshal rest", &rChannelType, []byte("\"REST\""), false},
		{"test unmarshal error", &eChannelType, []byte("\"foo\""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.as.UnmarshalJSON(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("ChannelType.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsChannelType(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{"test email", Email, true},
		{"test rest", Rest, true},
		{"test fail on not a channel type", "foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsChannelType(tt.arg); got != tt.want {
				t.Errorf("IsChannelType() = %v, want %v", got, tt.want)
			}
		})
	}
}
