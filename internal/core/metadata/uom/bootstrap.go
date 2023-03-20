//
// Copyright (C) 2022-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package uom

import (
	"context"
	"gopkg.in/yaml.v3"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

func BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	uomImpl := &UnitsOfMeasureImpl{}
	filepath := config.UoM.UoMFile
	// backward compatability for using older 2.x configuration
	// TODO: Remove in EdgeX 3.0
	if filepath == "" {
		dic.Update(di.ServiceConstructorMap{
			container.UnitsOfMeasureInterfaceName: func(get di.Get) interface{} {
				return uomImpl
			},
		})

		lc.Warn("UoM.UoMFile field not set in configuration file, unit of measure validation is disabled")
		return true
	}

	contents, err := os.ReadFile(filepath)
	if err != nil {
		lc.Errorf("could not load unit of measure configuration file (%s): %s", filepath, err.Error())
		return false
	}

	if err = yaml.Unmarshal(contents, uomImpl); err != nil {
		lc.Errorf("could not load unit of measure configuration file (%s): %s", filepath, err.Error())
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		container.UnitsOfMeasureInterfaceName: func(get di.Get) interface{} {
			return uomImpl
		},
	})

	lc.Infof("Loaded unit of measure configuration from %s", filepath)

	return true
}
