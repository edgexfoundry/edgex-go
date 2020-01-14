//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/middleware/debugging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/correlationid"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/config"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/container"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/logger/file"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/logger/mongo"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/logging"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gorilla/mux"
)

const (
	PersistenceDB   = "database"
	PersistenceFile = "file"
)

type server interface {
	IsRunning() bool
}

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	muxRouter              *mux.Router
	server                 server
	serviceKey             string
	inDebugMode            bool
	inV2AcceptanceTestMode bool
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(
	router *mux.Router,
	server server,
	serviceKey string,
	inDebugMode bool,
	inV2AcceptanceTestMode bool) *Bootstrap {

	return &Bootstrap{
		muxRouter:              router,
		server:                 server,
		serviceKey:             serviceKey,
		inDebugMode:            inDebugMode,
		inV2AcceptanceTestMode: inV2AcceptanceTestMode,
	}
}

// getPersistence is private method to factory an interfaces.Persistence implementation based on configuration.
func (b *Bootstrap) getPersistence(
	credentials *bootstrapConfig.Credentials,
	configuration *config.ConfigurationStruct) (interfaces.Persistence, error) {

	if b.inV2AcceptanceTestMode {
		return file.NewLogger(configuration.Logging.File), nil
	}

	switch configuration.Writable.Persistence {
	case PersistenceFile:
		return file.NewLogger(configuration.Logging.File), nil
	case PersistenceDB:
		return mongo.NewLogger(credentials, configuration)
	default:
		return nil, fmt.Errorf("unrecognized value Configuration.Persistence: %s", configuration.Writable.Persistence)
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization for the logging service.
func (b *Bootstrap) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	lc := logging.FactoryToStdout(b.serviceKey)
	configuration := container.ConfigurationFrom(dic.Get)

	if !b.inV2AcceptanceTestMode {
		loadV1Routes(b.muxRouter, dic)
	}
	b.loadV2Routes(dic, lc)

	// get database credentials.
	credentialsProvider := bootstrapContainer.CredentialsProviderFrom(dic.Get)
	var credentials bootstrapConfig.Credentials
	for startupTimer.HasNotElapsed() {
		var err error
		credentials, err = credentialsProvider.GetDatabaseCredentials(configuration.Databases["Primary"])
		if err == nil {
			break
		}
		lc.Warn(fmt.Sprintf("couldn't retrieve database credentials: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	// initialize database.
	var persistenceClient interfaces.Persistence
	var err error
	for startupTimer.HasNotElapsed() {
		persistenceClient, err = b.getPersistence(&credentials, configuration)
		if err == nil {
			break
		}
		lc.Warn(fmt.Sprintf("couldn't create database client: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	if err != nil {
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.PersistenceInterfaceName: func(get di.Get) interface{} {
			return persistenceClient
		},
	})

	lc.Info("Database connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before closing the database connection.
			if b.server.IsRunning() == false {
				lc.Info("Database disconnecting")
				persistenceClient.CloseSession()
				break
			}
			time.Sleep(time.Second)
		}
	}()

	return true
}

// loadV2Routes creates a new command-query router and handles the related mux.Router initialization for API V2 routes.
func (b *Bootstrap) loadV2Routes(_ *di.Container, lc logger.LoggingClient) {
	correlationid.WireUp(b.muxRouter)

	handlers := []delegate.Handler{}
	if b.inDebugMode {
		handlers = append(handlers, debugging.New(lc).Handler)
	}

	router.Initialize(
		b.muxRouter,
		handlers,
		common.V2Routes(
			b.inV2AcceptanceTestMode,
			[]router.Controller{},
		),
	)
}
