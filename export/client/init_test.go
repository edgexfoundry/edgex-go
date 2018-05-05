//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/export/client/clients"
)

func TestDestroy(t *testing.T) {
	// Set global state
	dbc = nil
	Destroy()

	var err error
	dbc, err = clients.NewDBClient(clients.DBConfiguration{
		DbType: clients.MEMORY,
	})
	if err != nil {
		t.Errorf("Error getting a memory client: %v", err)
	}
	Destroy()

	// Call it twice does not fail
	Destroy()
}
