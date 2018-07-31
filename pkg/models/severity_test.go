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

func TestNotificationsSeverity_UnmarshalJSON(t *testing.T) {
	var crtical = NotificationsSeverity(Critical)
	var normal = NotificationsSeverity(Normal)

	tests := []struct {
		name    string
		as      *NotificationsSeverity
		arg     []byte
		wantErr bool
	}{
		{"test marshal of critical", &crtical, []byte("\"CRITICAL\""), false},
		{"test marshal of normal", &normal, []byte("\"NORMAL\""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.as.UnmarshalJSON(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("NotificationsSeverity.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsNotificationsSeverity(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{"test critical", Critical, true},
		{"test normal", Normal, true},
		{"test fail on non-notification severity", "foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotificationsSeverity(tt.arg); got != tt.want {
				t.Errorf("IsNotificationsSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}
