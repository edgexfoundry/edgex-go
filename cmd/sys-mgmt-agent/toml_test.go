//
// Copyright (c) 2018
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
)

func TestToml(t *testing.T) {
	configuration := &interfaces.ConfigurationStruct{}
	if err := config.VerifyTomlFiles(configuration); err != nil {
		t.Fatalf("%v", err)
	}
	if configuration.AppOpenMsg == "" {
		t.Errorf("configuration.StartupMsg is zero length.")
	}
}
