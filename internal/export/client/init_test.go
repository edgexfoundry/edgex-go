//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"
)

func TestDestroy(t *testing.T) {
	// Set global state
	dbc = &memory.MemDB{}

	Destroy()

	// Call it twice does not fail
	Destroy()
}
