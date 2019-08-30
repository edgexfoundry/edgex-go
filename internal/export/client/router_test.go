//
// Copyright (c) 2018 Cavium
// Copyright (c) 2019 Dell Technologies, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

//const regJson = `{"origin":1471806386919,"name":"OSIClient","addressable":{"origin":1471806386919,"name":"OSIMQTTBroker","protocol":"TCP","address":"m10.cloudmqtt.com","port":15421,"publisher":"EdgeXExportPublisher","user":"hukfgtoh","password":"uP6hJLYW6Ji4","topic":"EdgeXDataTopic"},"format":"JSON","filter":{"deviceIdentifiers":["livingroomthermosat", "hallwaythermostat"],"valueDescriptorIdentifiers":["temperature", "humidity"]},"encryption":{"encryptionAlgorithm":"AES","encryptionKey":"123","initializingVector":"123"},"compression":"GZIP","enable":true, "destination": "REST_ENDPOINT"}`

var testTimestamps = models.Timestamps{Origin: 1471806386919}
var testAddressable = models.Addressable{Timestamps: testTimestamps, Name: "OSIMQTTBroker", Protocol: "TCP",
	Address: "m10.cloudmqtt.com", Port: 15421, Publisher: "EdgeXExportPublisher", User: "hukfgtoh", Password: "uP6hJLYW6Ji4",
	Topic: "EdgeXDataTopic"}
var testFilter = models.Filter{DeviceIDs: []string{"livingroomthermosat", "hallwaythermostat"}, ValueDescriptorIDs: []string{"temperature", "humidity"}}
var testEncryption = models.EncryptionDetails{Algo: "AES", Key: "123", InitVector: "123"}
var testRegistration = models.Registration{Origin: testTimestamps.Origin, Name: "OSIClient", Addressable: testAddressable,
	Format: models.FormatJSON, Filter: testFilter, Encryption: testEncryption, Compression: models.CompGzip, Enable: true,
	Destination: models.DestRest}

type distroMockClient struct{}

func (d *distroMockClient) NotifyRegistrations(models.NotifyUpdate, context.Context) error {
	return nil
}

func prepareTest(t *testing.T) *httptest.Server {
	LoggingClient = logger.NewClient(clients.ExportClientServiceKey, false, "./logs/edgex-export-client-test.log", models.InfoLog)

	dbClient = &MemDB{}
	dc = &distroMockClient{}
	return httptest.NewServer(LoadRestRoutes())
}

func createRegistration(t *testing.T, serverUrl string) string {
	request, err := json.Marshal(testRegistration)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}
	response, err := http.Post(serverUrl+clients.ApiRegistrationRoute, clients.ContentTypeJSON, bytes.NewBuffer(request))
	if err != nil {
		t.Errorf("Error sending log %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}
	var data []byte
	data, _ = ioutil.ReadAll(response.Body)
	return string(data)
}

func TestPing(t *testing.T) {
	// create test server with handler
	ts := httptest.NewServer(LoadRestRoutes())
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

func TestRegistrationAdd(t *testing.T) {
	regJson, err := json.Marshal(testRegistration)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}
	var tests = []struct {
		name   string
		data   string
		status int
	}{
		{"emptyPost", "", http.StatusBadRequest},
		{"emptyPost", "{}", http.StatusBadRequest},
		{"invalidJSON", "aa", http.StatusBadRequest},
		{"ok", `{"origin":1471806386919,"name":"NAME","addressable":{"origin":1471806386919,"name":"AnotherName","method":"POST","protocol":"TCP","address":"127.0.0.1","port":1883,"publisher":"SomePublisher","user":"dummy","password":"dummy","topic":"SomeTopic"},"format":"JSON","enable":true, "destination":"MQTT_TOPIC","compression":"NONE"}`, http.StatusOK},
		{"ok", string(regJson), http.StatusOK},
	}

	ts := prepareTest(t)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := http.Post(ts.URL+clients.ApiRegistrationRoute, "application/json",
				strings.NewReader(tt.data))
			if err != nil {
				t.Errorf("Error sending log %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
		})
	}

}

func TestRegistrationAddTwice(t *testing.T) {
	ts := prepareTest(t)
	defer ts.Close()

	createRegistration(t, ts.URL)

	request, err := json.Marshal(testRegistration)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}
	response, err := http.Post(ts.URL+clients.ApiRegistrationRoute, clients.ContentTypeJSON, bytes.NewBuffer(request))
	if err != nil {
		t.Errorf("Error sending log %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}
}

func requestMethod(t *testing.T, method string, url string, b io.Reader) *http.Response {
	req, err := http.NewRequest(method, url, b)
	req.Header.Add(clients.ContentType, clients.ContentTypeJSON)

	if err != nil {
		t.Fatalf("Error creating request: %v", err)
		return nil
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error getting response: %v", err)
		return nil
	}
	return response
}

func TestRegistrationUpdate(t *testing.T) {
	jsonOK, err := json.Marshal(testRegistration)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	testNoName := testRegistration
	testNoName.Name = "noname"
	jsonNoName, err := json.Marshal(testNoName)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	testNoId := testRegistration
	testNoId.ID = "507f1f77bcf86cd799439011"
	jsonNoId, err := json.Marshal(testNoId)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	testUpdateId := testRegistration
	testUpdateId.ID = "%s"
	jsonUpdateId, err := json.Marshal(testUpdateId)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	testUpdateName := testRegistration
	testUpdateName.ID = "%s"
	testUpdateName.Name = "OtherName"
	jsonUpdateName, err := json.Marshal(testUpdateName)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	var tests = []struct {
		name   string
		data   string
		status int
	}{
		{"emptyPost", "", http.StatusBadRequest},
		{"emptyPost", "{}", http.StatusBadRequest},
		{"notFound", string(jsonNoName), http.StatusNotFound},
		{"notFound", string(jsonNoId), http.StatusNotFound},
		{"updById", string(jsonUpdateId), http.StatusOK},
		{"updById", string(jsonUpdateName), http.StatusOK},
		{"updById", `{"id":"%s", "compression":"INVALID"}`, http.StatusBadRequest},
		{"updByName", string(jsonOK), http.StatusOK},
		{"updByName", `{"Name":"OSIClient", "compression":"INVALID"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize for each test, to have the same registration before updating
			ts := prepareTest(t)
			defer ts.Close()

			id := createRegistration(t, ts.URL)

			var d string
			// If there is a %s in the string substitute it with the id
			if strings.Index(tt.data, "%s") != -1 {
				d = fmt.Sprintf(tt.data, id)
			} else {
				d = tt.data
			}

			response := requestMethod(t, http.MethodPut, ts.URL+clients.ApiRegistrationRoute,
				bytes.NewBufferString(d))
			defer response.Body.Close()

			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
		})
	}
}

func TestRegistrationDelByName(t *testing.T) {
	ts := prepareTest(t)
	defer ts.Close()

	createRegistration(t, ts.URL)

	response := requestMethod(t, http.MethodDelete, ts.URL+clients.ApiRegistrationRoute+"/name/invalid",
		nil)
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusNotFound)
	}
	response = requestMethod(t, http.MethodDelete, ts.URL+clients.ApiRegistrationRoute+"/name/OSIClient",
		nil)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}

	response = requestMethod(t, http.MethodDelete, ts.URL+clients.ApiRegistrationRoute+"/name/OSIClient",
		nil)
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestRegistrationDelById(t *testing.T) {
	ts := prepareTest(t)
	defer ts.Close()

	id := createRegistration(t, ts.URL)

	response := requestMethod(t, http.MethodDelete, ts.URL+clients.ApiRegistrationRoute+"/id/invalid",
		nil)
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusNotFound)
	}

	response = requestMethod(t, http.MethodDelete, ts.URL+clients.ApiRegistrationRoute+"/id/"+id,
		nil)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}

	response = requestMethod(t, http.MethodDelete, ts.URL+clients.ApiRegistrationRoute+"/id/"+id,
		nil)
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestRegistrationGetByName(t *testing.T) {
	ts := prepareTest(t)
	defer ts.Close()

	createRegistration(t, ts.URL)

	response, err := http.Get(ts.URL + clients.ApiRegistrationRoute + "/name/OSIClient")
	if err != nil {
		t.Errorf("Error getting registration: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}

	response, err = http.Get(ts.URL + clients.ApiRegistrationRoute + "/name/invalid")
	if err != nil {
		t.Errorf("Error getting registration: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestRegistrationGetById(t *testing.T) {
	ts := prepareTest(t)
	defer ts.Close()

	id := createRegistration(t, ts.URL)

	response, err := http.Get(ts.URL + clients.ApiRegistrationRoute + "/" + id)
	if err != nil {
		t.Errorf("Error getting registration: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}

	response, err = http.Get(ts.URL + clients.ApiRegistrationRoute + "/invalid")
	if err != nil {
		t.Errorf("Error getting registration: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestRegistrationGetList(t *testing.T) {
	var tests = []struct {
		typeStr string
		status  int
	}{
		{"", http.StatusNotFound},
		{"invalid", http.StatusBadRequest},
		{typeAlgorithms, http.StatusOK},
		{typeCompressions, http.StatusOK},
		{typeFormats, http.StatusOK},
		{typeDestinations, http.StatusOK},
	}

	ts := prepareTest(t)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.typeStr, func(t *testing.T) {
			response, err := http.Get(ts.URL + clients.ApiRegistrationRoute + "/reference/" + tt.typeStr)
			if err != nil {
				t.Errorf("Error getting reference type: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
		})
	}
}

func getRegistrations(t *testing.T, serverUrl string) []models.Registration {
	response, err := http.Get(serverUrl + clients.ApiRegistrationRoute)
	if err != nil {
		t.Errorf("Error getting registrations: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}

	var data []byte
	data, _ = ioutil.ReadAll(response.Body)
	var regs []models.Registration
	if err := json.Unmarshal(data, &regs); err != nil {
		t.Errorf("Registrations could not be parsed: %v", err)
	}

	return regs
}

func TestRegistrationGetAll(t *testing.T) {
	ts := prepareTest(t)
	defer ts.Close()

	regs := getRegistrations(t, ts.URL)
	if len(regs) != 0 {
		t.Errorf("There should not be registrations: %v", regs)
	}

	createRegistration(t, ts.URL)

	regs = getRegistrations(t, ts.URL)
	if len(regs) != 1 {
		t.Errorf("There should be only one registrations: %v", regs)
	}
}
