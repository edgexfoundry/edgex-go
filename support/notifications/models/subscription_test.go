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

var TestEmptySubscription = Subscription{}
var TestSubscription = Subscription{BaseObject: TestBaseObject, Slug: "test slug", Receiver: "test receiver", Description: "test description",
	SubscribedCategories: []NotificationsCategory{NotificationsCategory(Swhealth)}, SubscribedLabels: []string{"test label"},
	Channels: []Channel{TestEChannel, TestRChannel}}

func TestSubscription_MarshalJSON(t *testing.T) {
	var testEmptyBytes = []byte(TestEmptySubscription.String())
	var testSubBytes = []byte(TestSubscription.String())

	tests := []struct {
		name    string
		sub     *Subscription
		want    []byte
		wantErr bool
	}{
		{"test empty subscription", &TestEmptySubscription, testEmptyBytes, false},
		{"test subscription", &TestSubscription, testSubBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sub.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscription.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscription.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscription_String(t *testing.T) {
	tests := []struct {
		name string
		sub  *Subscription
		want string
	}{
		{"test string empty subscription", &TestEmptySubscription, "{\"created\":0,\"modified\":0,\"id\":null}"},
		{"test subscription", &TestSubscription, "{\"created\":123,\"modified\":123,\"id\":null,\"slug\":\"test slug\",\"receiver\":\"test receiver\",\"description\":\"test description\",\"subscribedcategories\":[\"SW_HEALTH\"],\"subscribedlabels\":[\"test label\"],\"channels\":[{\"channeltype\":\"EMAIL\",\"mailaddresses\":[\"jpwhite_mn@yahoo.com\",\"james_white2@dell.com\"]},{\"channeltype\":\"REST\",\"url\":\"http://www.someendpoint.com/notifications\"}]}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sub.String(); got != tt.want {
				t.Errorf("Subscription.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
