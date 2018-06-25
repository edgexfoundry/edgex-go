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
	"testing"
)

var TestEChannel = Channel{Type: ChannelType(Email), MailAddresses: []string{"jpwhite_mn@yahoo.com", "james_white2@dell.com"}}
var TestRChannel = Channel{Type: ChannelType(Rest), Url: "http://www.someendpoint.com/notifications"}
var TestEmptyChannel = Channel{}

func TestChannel_String(t *testing.T) {

	tests := []struct {
		name string
		c    *Channel
		want string
	}{
		{"email channel to string", &TestEChannel, "{\"channeltype\":\"EMAIL\",\"mailaddresses\":[\"jpwhite_mn@yahoo.com\",\"james_white2@dell.com\"]}"},
		{"rest channel to string ", &TestRChannel, "{\"channeltype\":\"REST\",\"url\":\"http://www.someendpoint.com/notifications\"}"},
		{"empty channel to string", &TestEmptyChannel, "{}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("Channel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
