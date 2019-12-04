package main

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	schedConfig "github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"

	"testing"
)

func TestToml(t *testing.T) {
	configuration := &schedConfig.ConfigurationStruct{}
	if err := config.VerifyTomlFiles(configuration); err != nil {
		t.Fatalf("%v", err)
	}
	if configuration.Service.StartupMsg == "" {
		t.Errorf("configuration.StartupMsg is zero length.")
	}
}
