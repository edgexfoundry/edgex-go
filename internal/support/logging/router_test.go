//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

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
