/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2020 Intel Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package proxy

import (
	"context"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Bootstrap struct {
	insecureSkipVerify bool
	initNeeded         bool
	resetNeeded        bool
}

func NewBootstrap(
	insecureSkipVerify bool,
	initNeeded bool,
	resetNeeded bool) *Bootstrap {

	return &Bootstrap{
		insecureSkipVerify: insecureSkipVerify,
		initNeeded:         initNeeded,
		resetNeeded:        resetNeeded,
	}
}

func (b *Bootstrap) errorAndHalt(lc logger.LoggingClient, message string) {
	lc.Error(message)
	os.Exit(1)
}

func (b *Bootstrap) haltIfError(lc logger.LoggingClient, err error) {
	if err != nil {
		b.errorAndHalt(lc, err.Error())
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	var req internal.HttpCaller
	if len(configuration.SecretStore.RootCaCertPath) > 0 {
		req = NewRequestor(
			b.insecureSkipVerify,
			configuration.RequestTimeout,
			configuration.SecretStore.RootCaCertPath,
			lc)
	} else {
		req = NewRequestor(
			true, // non-TLS mode internally
			configuration.RequestTimeout,
			"", // irrelevant
			lc)
	}

	if req == nil {
		os.Exit(1)
	}

	s := NewService(req, lc, configuration)
	b.haltIfError(lc, s.CheckProxyServiceStatus())

	if b.initNeeded {
		if b.resetNeeded {
			b.errorAndHalt(lc, "can't run initialization and reset at the same time for security service")
		}

		// Based on the ADR: No certificate pair internally any more
		b.haltIfError(lc, s.Init()) // Where the Service init is called
	} else if b.resetNeeded {
		b.haltIfError(lc, s.ResetProxy())
	}

	return true
}
