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

func TestNotificationsStatus_UnmarshalJSON(t *testing.T) {
	var new = NotificationsStatus(New)
	var processed = NotificationsStatus(Processed)
	var escalated = NotificationsStatus(Escalated)
	tests := []struct {
		name    string
		as      *NotificationsStatus
		arg     []byte
		wantErr bool
	}{
		{"test marshal of new", &new, []byte("\"NEW\""), false},
		{"test marshal of processed", &processed, []byte("\"PROCESSED\""), false},
		{"test marshal of escalated", &escalated, []byte("\"ESCALATED\""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.as.UnmarshalJSON(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("NotificationsStatus.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsNotificationsStatus(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{"test new", New, true},
		{"test processed", Processed, true},
		{"test escalated", Escalated, true},
		{"test faile on non-notif status", "foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotificationsStatus(tt.args); got != tt.want {
				t.Errorf("IsNotificationsStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
