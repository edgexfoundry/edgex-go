//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"bytes"
	"encoding/json"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

func TestPing(t *testing.T) {
	// create test server with handler
	ts := httptest.NewServer(httpServer())
	defer ts.Close()

	response, err := http.Get(ts.URL + clients.ApiPingRoute)
	if err != nil {
		t.Errorf("Error getting ping: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}
}

func TestReplyNotifyRegistrations(t *testing.T) {
	nuNoName := models.NotifyUpdate{Operation: "add"}
	nuNoOp := models.NotifyUpdate{Name: "aaa"}
	nuInvalidOp := models.NotifyUpdate{Name: "aaa", Operation: "aadd"}
	nuValidAdd := models.NotifyUpdate{Name: "aaa", Operation: "add"}
	nuValidDelete := models.NotifyUpdate{Name: "aaa", Operation: "delete"}
	nuValidUpdate := models.NotifyUpdate{Name: "aaa", Operation: "update"}

	var tests = []struct {
		name   string
		nu     models.NotifyUpdate
		status int
	}{
		{"empty", models.NotifyUpdate{}, http.StatusBadRequest},
		{"noName", nuNoName, http.StatusBadRequest},
		{"noOperation", nuNoOp, http.StatusBadRequest},
		{"invalidOperation", nuInvalidOp, http.StatusBadRequest},
		{"validAdd", nuValidAdd, http.StatusOK},
		{"validDelete", nuValidDelete, http.StatusOK},
		{"validUpdate", nuValidUpdate, http.StatusOK},
	}
	// create test server with handler
	ts := httptest.NewServer(httpServer())
	defer ts.Close()

	url := ts.URL + clients.ApiNotifyRegistrationRoute

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			data, err := json.Marshal(tt.nu)
			if err != nil {
				t.Errorf("marshaling error %v", err)
			}
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
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
