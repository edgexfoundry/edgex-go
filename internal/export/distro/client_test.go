//
// Copyright (c) 2017 Cavium
// Copyright (c) 2019 Dell Technologies, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var testAddressable = models.Addressable{Name: "OTROMAS-1", HTTPMethod: "POST", Protocol: "TCP", Address: "127.0.0.1", Port: 1883,
	Publisher: "FuseExportPublisher_OTROMAS-1", User: "dummy", Password: "dummy", Topic: "FuseDataTopic"}
var testRegistration = models.Registration{ID: "5a15918fa4a9b92af1c94bab", Origin: 1471806386919, Name: "OTROMAS-1",
	Addressable: testAddressable, Format: models.FormatJSON, Enable: true, Destination: models.DestMQTT}

const (
	emptyRegistrationList = "[]"
	invalidReply1         = "[[]]"
	invalidReply2         = ""
)

func TestMain(m *testing.M) {
	Configuration = &ConfigurationStruct{}
	LoggingClient = logger.NewMockClient()
	os.Exit(m.Run())
}

func TestClientRegistrationsEmpty(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, emptyRegistrationList)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	regs, err := getRegistrationsURL(ts.URL)
	if err != nil {
		t.Error(err)
	}
	if regs == nil {
		t.Fatal("nil registration list")
	}
	if len(regs) != 0 {
		t.Fatal("Registration should be empty")
	}
}

func TestClientRegistrations(t *testing.T) {
	data, err := json.Marshal(testRegistration)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "["+string(data)+"]")
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	regs, err := getRegistrationsURL(ts.URL)
	if err != nil {
		t.Error(err)
	}
	if regs == nil {
		t.Fatal("nil registration list")
	}
	if len(regs) != 1 {
		t.Fatal("Registration list should have only a registration")
	}
}

func TestClientRegistrationsInvalid(t *testing.T) {
	invalidList := []string{invalidReply1, invalidReply2}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var ir string
		ir, invalidList = invalidList[0], invalidList[1:]
		fmt.Fprint(w, ir)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	for range invalidList {
		regs, _ := getRegistrationsURL(ts.URL)
		if len(regs) > 0 {
			t.Fatal("Registration list should be empty", regs)
		}
	}
}

func TestClientRegistrationsInvalidRegistration(t *testing.T) {
	valid := testRegistration
	validData, err := json.Marshal(valid)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	invalid := testRegistration
	invalid.Compression = "ZERO"
	invalidData, err := json.Marshal(invalid)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	invalidList := []struct {
		str       string
		validRegs int
	}{
		{"[" + string(invalidData) + "]", 0},
		{"[" + string(validData) + "," + string(invalidData) + "]", 1},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, invalidList[0].str)
		invalidList = invalidList[1:]
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	for _, v := range invalidList {
		regs, err := getRegistrationsURL(ts.URL)
		if err != nil {
			t.Error(err)
		}
		if regs == nil {
			t.Fatal("nil registration list")
		}
		if len(regs) != v.validRegs {
			t.Fatal("Registration list should have ", v.validRegs, ". It had ", len(regs))
		}
	}
}

func TestClientRegistrationsInvalidRegistration2(t *testing.T) {
	valid := testRegistration
	validData, err := json.Marshal(valid)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	invalid := testRegistration
	invalid.Compression = "ZERO"
	invalidData, err := json.Marshal(invalid)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "["+string(validData)+","+string(invalidData)+"]")
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	regs, err := getRegistrationsURL(ts.URL)
	if err != nil {
		t.Error(err)
	}
	if regs == nil {
		t.Fatal("nil registration list")
	}
	if len(regs) != 1 {
		t.Fatal("Registration should be empty")
	}
}

func TestClientRegistrationByName(t *testing.T) {
	data, err := json.Marshal(testRegistration)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(data))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	reg := getRegistrationByNameURL(ts.URL)
	if reg == nil {
		t.Fatal("nil registration")
	}
}

func TestClientRegistrationByNameError(t *testing.T) {
	invalid := testRegistration
	invalid.Compression = "ZERO"
	data, err := json.Marshal(invalid)
	if err != nil {
		t.Errorf("marshaling error %v", err)
	}

	invalidList := []string{invalidReply1, invalidReply2, string(data)}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var ir string
		ir, invalidList = invalidList[0], invalidList[1:]
		fmt.Fprint(w, ir)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	for range invalidList {
		reg := getRegistrationByNameURL(ts.URL)
		if reg != nil {
			t.Fatal("Registration should be nil")
		}
	}
}
