/*******************************************************************************
 * Copyright 2020 Dell Inc.
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
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// httpServer defines the contract used to determine whether or not the http httpServer is running.
type httpServer interface {
	IsRunning() bool
}

// Ready contains references to dependencies required by the testing implementation.
type Ready struct {
	httpServer httpServer
	stream     chan<- bool
}

// NewReady is a factory method that returns an initialized Ready receiver struct.
func NewReady(httpServer httpServer, stream chan<- bool) *Ready {
	return &Ready{
		httpServer: httpServer,
		stream:     stream,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract.  During normal production execution, a nil stream
// will be supplied.  A non-nil stream indicates we're running within the test runner context and that we should
// wait for the httpServer to start running before sending confirmation over the stream.  If the httpServer doesn't
// start running within the defined startup time, no confirmation is sent over the stream and the application
// bootstrapping is aborted.
func (r *Ready) BootstrapHandler(
	_ context.Context,
	_ *sync.WaitGroup,
	startupTimer startup.Timer,
	_ *di.Container) bool {

	if r.stream != nil {
		for startupTimer.HasNotElapsed() {
			if r.httpServer.IsRunning() {
				r.stream <- true
				return true
			}
			startupTimer.SleepForInterval()
		}
		return false
	}
	return true
}
