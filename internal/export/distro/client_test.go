//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

const (
	emptyRegistrationList    = "[]"
	registrationStr          = `{"_id":"5a15918fa4a9b92af1c94bab","created":0,"modified":0,"origin":1471806386919,"name":"OTROMAS-1","addressable":{"Name":"OTROMAS-1","Method":"POST","Protocol":"TCP","Address":"127.0.0.1","Port":1883,"Path":"","Publisher":"FuseExportPublisher_OTROMAS-1","User":"dummy","Password":"dummy","Topic":"FuseDataTopic"},"format":"JSON","filter":{},"encryption":{},"compression":"NONE","enable":true,"destination":"MQTT_TOPIC"}`
	registrationInvalidStr   = `{"_id":"5a15918fa4a9b92af1c94bab","created":0,"modified":0,"origin":1471806386919,"name":"OTROMAS-1","addressable":{"Name":"OTROMAS-1","Method":"POST","Protocol":"TCP","Address":"127.0.0.1","Port":1883,"Path":"","Publisher":"FuseExportPublisher_OTROMAS-1","User":"dummy","Password":"dummy","Topic":"FuseDataTopic"},"format":"JSON","filter":{},"encryption":{},"compression":"ZERO","enable":true,"destination":"MQTT_TOPIC"}`
	oneRegistrationList      = "[" + registrationStr + "]"
	invalidReply1            = "[[]]"
	invalidReply2            = ""
	invalidRegistrationList1 = "[" + registrationInvalidStr + "]"
	invalidRegistrationList2 = "[" + registrationInvalidStr + "," + registrationStr + "]"
)

func TestClientRegistrationsEmpty(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, emptyRegistrationList)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	regs := getRegistrationsURL(ts.URL)
	if regs == nil {
		t.Fatal("nil registration list")
	}
	if len(regs) != 0 {
		t.Fatal("Registration should be empty")
	}
}

func TestClientRegistrations(t *testing.T) {
	logger = zap.NewNop()
	defer logger.Sync()

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, oneRegistrationList)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	regs := getRegistrationsURL(ts.URL)
	if regs == nil {
		t.Fatal("nil registration list")
	}
	if len(regs) != 1 {
		t.Fatal("Registration list should have only a registration")
	}
}

func TestClientRegistrationsInvalid(t *testing.T) {
	logger = zap.NewNop()
	defer logger.Sync()

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
		regs := getRegistrationsURL(ts.URL)
		if regs != nil {
			t.Fatal("Registration list should be nil", regs)
		}
	}
}

func TestClientRegistrationsInvalidRegistration(t *testing.T) {
	invalidList := []struct {
		str       string
		validRegs int
	}{
		{invalidRegistrationList1, 0},
		{invalidRegistrationList2, 1},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, invalidList[0].str)
		invalidList = invalidList[1:]
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	for _, v := range invalidList {
		regs := getRegistrationsURL(ts.URL)
		if regs == nil {
			t.Fatal("nil registration list")
		}
		if len(regs) != v.validRegs {
			t.Fatal("Registration list should have ", v.validRegs, ". It had ", len(regs))
		}
	}
}

func TestClientRegistrationsInvalidRegistration2(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, invalidRegistrationList2)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	regs := getRegistrationsURL(ts.URL)
	if regs == nil {
		t.Fatal("nil registration list")
	}
	if len(regs) != 1 {
		t.Fatal("Registration should be empty")
	}
}

func TestClientRegistrationByName(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, registrationStr)
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
	invalidList := []string{invalidReply1, invalidReply2, registrationInvalidStr}

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
