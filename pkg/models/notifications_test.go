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

var TestEmptyNotification = Notification{}
var TestNotification = Notification{BaseObject: BaseObject{Created: 123, Modified: 123}, Category: NotificationsCategory("SECURITY"),
	Content: "test content", Description: "test description", Labels: []string{"label1", "labe2"}, Sender: "test sender",
	Severity: NotificationsSeverity("CRITICAL"), Slug: "test slug", Status: NotificationsStatus("NEW")}

func TestNotification_MarshalJSON(t *testing.T) {
	var testEmptyNotifBytes = []byte(TestEmptyNotification.String())
	var testNotifBytes = []byte(TestNotification.String())
	tests := []struct {
		name         string
		notification *Notification
		want         []byte
		wantErr      bool
	}{
		{"test marshal of empty notification", &TestEmptyNotification, testEmptyNotifBytes, false},
		{"test marshal of notification", &TestNotification, testNotifBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Notification{
				BaseObject:  tt.notification.BaseObject,
				ID:          tt.notification.ID,
				Slug:        tt.notification.Slug,
				Sender:      tt.notification.Sender,
				Category:    tt.notification.Category,
				Severity:    tt.notification.Severity,
				Content:     tt.notification.Content,
				Description: tt.notification.Description,
				Status:      tt.notification.Status,
				Labels:      tt.notification.Labels,
				ContentType: tt.notification.ContentType,
			}
			got, err := n.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Notification.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Notification.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotification_String(t *testing.T) {
	tests := []struct {
		name         string
		notification *Notification
		want         string
	}{
		{"test empty notification to string", &TestEmptyNotification, "{\"created\":0,\"modified\":0,\"id\":null}"},
		{"test notification to string", &TestNotification, "{\"created\":123,\"modified\":123,\"id\":null,\"slug\":\"test slug\",\"sender\":\"test sender\",\"category\":\"SECURITY\",\"severity\":\"CRITICAL\",\"content\":\"test content\",\"description\":\"test description\",\"status\":\"NEW\",\"labels\":[\"label1\",\"labe2\"]}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notification.String(); got != tt.want {
				t.Errorf("Notification.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
