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
 *******************************************************************************/

package v2

import (
	"context"
	"sync"
	"time"

	v2 "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container/v2"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/persistence/memory"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

// httpServer defines the contract used to determine whether or not the http httpServer is running.
type httpServer interface {
	IsRunning() bool
}

// Handler contains references to dependencies required by the database bootstrap implementation.
type Handler struct {
	httpServer httpServer
}

// NewPersistence is a factory method that returns an initialized Database receiver struct.
func NewPersistence(httpServer httpServer) Handler {
	return Handler{
		httpServer: httpServer,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and initializes the database.
func (d Handler) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	dic.Update(di.ServiceConstructorMap{
		v2.PersistenceInterfaceName: func(get di.Get) interface{} {
			return memory.New()
		},
	})

	lc.Info("APIv2 persistence connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before collapsing persistence.
			if d.httpServer.IsRunning() == false {
				break
			}
			time.Sleep(time.Second)
		}
		lc.Info("APIv2 persistence disconnected")
	}()

	return true
}
