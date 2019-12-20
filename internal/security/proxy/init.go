/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
	"fmt"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/proxy/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type Bootstrap struct {
	insecureSkipVerify bool
	initNeeded         bool
	resetNeeded        bool
	userTobeCreated    string
	userOfGroup        string
	userToBeDeleted    string
}

func NewBootstrapHandler(
	insecureSkipVerify bool,
	initNeeded bool,
	resetNeeded bool,
	userTobeCreated string,
	userOfGroup string,
	userToBeDeleted string) *Bootstrap {

	return &Bootstrap{
		insecureSkipVerify: insecureSkipVerify,
		initNeeded:         initNeeded,
		resetNeeded:        resetNeeded,
		userTobeCreated:    userTobeCreated,
		userOfGroup:        userOfGroup,
		userToBeDeleted:    userToBeDeleted,
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
func (b *Bootstrap) Handler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	req := NewRequestor(
		b.insecureSkipVerify,
		configuration.Writable.RequestTimeout,
		configuration.SecretService.CACertPath,
		lc)
	if req == nil {
		os.Exit(1)
	}

	s := NewService(req, lc, configuration)
	b.haltIfError(lc, s.CheckProxyServiceStatus())

	if b.initNeeded {
		if b.resetNeeded {
			b.errorAndHalt(lc, "can't run initialization and reset at the same time for security service")
		}

		b.haltIfError(
			lc,
			s.Init(
				NewCertificateLoader(
					req,
					configuration.SecretService.CertPath,
					configuration.SecretService.TokenPath,
					configuration.SecretService.GetSecretSvcBaseURL(),
					lc,
				),
			),
		) // Where the Service init is called
	} else if b.resetNeeded {
		b.haltIfError(lc, s.ResetProxy())
	}

	if b.userTobeCreated != "" && b.userOfGroup != "" {
		c := NewConsumer(b.userTobeCreated, req, lc, configuration)
		b.haltIfError(lc, c.Create(EdgeXKong))
		b.haltIfError(lc, c.AssociateWithGroup(b.userOfGroup))

		t, err := c.CreateToken()
		if err != nil {
			b.errorAndHalt(lc, fmt.Sprintf("failed to create access token for edgex service due to error %s", err.Error()))
		}

		fmt.Println(fmt.Sprintf("the access token for user %s is: %s. Please keep the token for accessing edgex services", b.userTobeCreated, t))

		file, err := os.Create(configuration.KongAuth.OutputPath)
		b.haltIfError(lc, err)

		utp := &UserTokenPair{User: b.userTobeCreated, Token: t}
		b.haltIfError(lc, utp.Save(file))
	}

	if b.userToBeDeleted != "" {
		t := NewConsumer(b.userToBeDeleted, req, lc, configuration)
		b.haltIfError(lc, t.Delete())
	}

	return false
}
