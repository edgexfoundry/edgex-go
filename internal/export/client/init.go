//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/export/distro"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
)

// Global variables
var dbClient export.DBClient
var LoggingClient logger.LoggingClient
var Configuration = &ConfigurationStruct{}
var dc distro.DistroClient

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the export-client service.
func BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {

	// update global variables.
	LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient = container.DBClientFrom(dic.Get)

	// initialize clients required by service.
	registryClient := bootstrapContainer.RegistryFrom(dic.Get)
	dc = distro.NewDistroClient(
		types.EndpointParams{
			ServiceKey:  clients.ExportDistroServiceKey,
			UseRegistry: registryClient != nil,
			Url:         Configuration.Clients["Distro"].Url(),
			Interval:    Configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registryClient})

	return true
}
