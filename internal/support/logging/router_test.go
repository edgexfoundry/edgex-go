//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/container"
)

func TestPing(t *testing.T) {
	// create test server with handler
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.PersistenceName: func(get di.Get) interface{} {
			return dummyPersist{}
		},
	})
	ts := httptest.NewServer(LoadRestRoutes(dic))
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
