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

package handlers

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"

	"github.com/gorilla/mux"
)

// Server contains references to dependencies required by the http server implementation.
type Server struct {
	router    *mux.Router
	isRunning bool
}

// NewServerBootstrap is a factory method that returns an initialized Server receiver struct.
func NewServerBootstrap(router *mux.Router) Server {
	return Server{
		router:    router,
		isRunning: false,
	}
}

// IsRunning returns whether or not the http server is running.  It is provided to support delayed shutdown of
// any resources required to successfully process http requests until after all outstanding requests have been
// processed (e.g. a database connection).
func (h *Server) IsRunning() bool {
	return h.isRunning
}

// Handler fulfills the BootstrapHandler contract.  It creates two go routines -- one that executes ListenAndServe()
// and another that waits on closure of a context's done channel before calling Shutdown() to cleanly shut down the
// http server.
func (h *Server) Handler(
	wg *sync.WaitGroup,
	ctx context.Context,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	bootstrapConfig := container.ConfigurationFrom(dic.Get).GetBootstrap()
	addr := bootstrapConfig.Service.Host + ":" + strconv.Itoa(bootstrapConfig.Service.Port)
	timeout := time.Millisecond * time.Duration(bootstrapConfig.Service.Timeout)
	server := &http.Server{
		Addr:         addr,
		Handler:      h.router,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	}

	loggingClient := container.LoggingClientFrom(dic.Get)
	loggingClient.Info("Web server starting (" + addr + ")")

	wg.Add(1)
	go func() {
		defer wg.Done()

		h.isRunning = true
		server.ListenAndServe()
		loggingClient.Info("Web server stopped")
		h.isRunning = false
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		loggingClient.Info("Web server shutting down")
		server.Shutdown(context.Background())
		loggingClient.Info("Web server shut down")
	}()

	return true
}
