//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	existedRegistrations, err := dbClient.Registrations()
	if err != nil {
		lc.Errorf("Failed to get registrations from database: %s", err.Error())
		return false
	}

	c := NewRegistry(ctx, wg, dic)
	for _, r := range existedRegistrations {
		if r.Status != models.Halt {
			c.Register(r)
		}
	}

	dic.Update(di.ServiceConstructorMap{
		container.RegistryInterfaceName: func(get di.Get) interface{} {
			return c
		},
	})

	return true
}
