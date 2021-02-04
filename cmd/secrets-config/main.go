//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package main

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/config"
)

func main() {
	os.Setenv("LOGLEVEL", "ERROR") // Workaround for https://github.com/edgexfoundry/edgex-go/issues/2922
	ctx, cancel := context.WithCancel(context.Background())
	exitStatusCode := config.Main(ctx, cancel)
	os.Exit(exitStatusCode)
}
