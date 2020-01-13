//
// Copyright (c) 2018
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/support/logging"

	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logging.Main(ctx, cancel, mux.NewRouter(), nil)
}
