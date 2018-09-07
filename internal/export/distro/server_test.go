//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	// create test server with handler
	ts := httptest.NewServer(httpServer())
	defer ts.Close()

	response, err := http.Get(ts.URL + "/api/v1" + "/ping")
	if err != nil {
		t.Errorf("Error getting ping: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}
}

func TestReplyNotifyRegistrations(t *testing.T) {
	var tests = []struct {
		name   string
		data   string
		status int
	}{
		{"empty", "", http.StatusBadRequest},
		{"empty", "{}", http.StatusBadRequest},
		{"noName", `{"operation": "add"}`, http.StatusBadRequest},
		{"noOperation", `{"name": "aaa"}`, http.StatusBadRequest},
		{"invalidOperation", `{"operation": "aadd", "name": "aaa"}`, http.StatusBadRequest},
		{"validOperation", `{"operation": "add", "name": "aaa"}`, http.StatusOK},
		{"validOperation", `{"operation": "delete", "name": "aaa"}`, http.StatusOK},
		{"validOperation", `{"operation": "update", "name": "aaa"}`, http.StatusOK},
	}
	// create test server with handler
	ts := httptest.NewServer(httpServer())
	defer ts.Close()

	url := ts.URL + apiV1NotifyRegistrations

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(tt.data))
			if err != nil {
				t.Errorf("Error creating http request: %v", err)
			}
			response, err := client.Do(req)
			if err != nil {
				t.Errorf("Error sending update: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			if tt.status == http.StatusOK {
				// Remove the inserted notification update from the channel
				<-registrationChanges
			}

		})
	}
}
