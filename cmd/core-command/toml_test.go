//
// Copyright (c) 2018
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"testing"

	commandConfig "github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
)

func TestToml(t *testing.T) {
	configuration := &commandConfig.ConfigurationStruct{}
	if err := config.VerifyTomlFiles(configuration); err != nil {
		t.Fatalf("%v", err)
	}
	if configuration.Service.StartupMsg == "" {
		t.Errorf("configuration.StartupMsg is zero length.")
	}
}
