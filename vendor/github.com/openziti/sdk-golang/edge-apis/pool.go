/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package edge_apis

import (
	"github.com/go-openapi/runtime"
	"github.com/michaelquigley/pfxlog"
	cmap "github.com/orcaman/concurrent-map/v2"
	errors "github.com/pkg/errors"
	"golang.org/x/exp/rand"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

type ApiClientTransport struct {
	runtime.ClientTransport
	ApiUrl *url.URL
}

// ClientTransportPool abstracts the concept of multiple `runtime.ClientTransport` (openapi interface) representing one
// target OpenZiti network. In situations where controllers are running in HA mode (multiple controllers) this
// interface can attempt to try different controller during outages or partitioning.
type ClientTransportPool interface {
	runtime.ClientTransport

	Add(apiUrl *url.URL, transport runtime.ClientTransport)
	Remove(apiUrl *url.URL)

	GetActiveTransport() *ApiClientTransport
	SetActiveTransport(*ApiClientTransport)
	GetApiUrls() []*url.URL
	IterateTransportsRandomly() chan<- *ApiClientTransport

	TryTransportsForOp(operation *runtime.ClientOperation) (any, error)
	TryTransportForF(cb func(*ApiClientTransport) (any, error)) (any, error)
}

var _ runtime.ClientTransport = (ClientTransportPool)(nil)
var _ ClientTransportPool = (*ClientTransportPoolRandom)(nil)

// ClientTransportPoolRandom selects a client transport (controller) at random until it is unreachable. Controllers
// are tried at random until a controller is reached. The newly connected controller is set for use on future requests
// until is too becomes unreachable.
type ClientTransportPoolRandom struct {
	pool    cmap.ConcurrentMap[string, *ApiClientTransport]
	current atomic.Pointer[ApiClientTransport]
}

func (c *ClientTransportPoolRandom) IterateTransportsRandomly() chan<- *ApiClientTransport {
	channel := make(chan *ApiClientTransport, 1)

	go func() {
		var transports []*ApiClientTransport

		for tpl := range c.pool.IterBuffered() {
			transports = append(transports, tpl.Val)
		}

		for len(transports) > 0 {
			var selected *ApiClientTransport
			selected, transports = selectAndRemoveRandom(transports, nil)

			if selected != nil {
				channel <- selected
			}
		}
	}()

	return channel
}

func (c *ClientTransportPoolRandom) GetApiUrls() []*url.URL {
	var result []*url.URL

	for tpl := range c.pool.IterBuffered() {
		result = append(result, tpl.Val.ApiUrl)
	}

	return result
}

func (c *ClientTransportPoolRandom) GetActiveTransport() *ApiClientTransport {
	active := c.current.Load()
	if active == nil {
		active = c.AnyTransport()
		c.SetActiveTransport(active)
	}

	return active
}

func (c *ClientTransportPoolRandom) GetApiClientTransports() []*ApiClientTransport {
	var result []*ApiClientTransport

	for tpl := range c.pool.IterBuffered() {
		result = append(result, tpl.Val)
	}

	return result
}

func NewClientTransportPoolRandom() *ClientTransportPoolRandom {
	return &ClientTransportPoolRandom{
		pool:    cmap.New[*ApiClientTransport](),
		current: atomic.Pointer[ApiClientTransport]{},
	}
}

func (c *ClientTransportPoolRandom) SetActiveTransport(transport *ApiClientTransport) {
	pfxlog.Logger().WithField("key", transport.ApiUrl.String()).Debug("setting active controller")
	c.current.Store(transport)
}

func (c *ClientTransportPoolRandom) Add(apiUrl *url.URL, transport runtime.ClientTransport) {
	c.pool.Set(apiUrl.String(), &ApiClientTransport{
		ClientTransport: transport,
		ApiUrl:          apiUrl,
	})
}

func (c *ClientTransportPoolRandom) Remove(apiUrl *url.URL) {
	c.pool.Remove(apiUrl.String())
}

func (c *ClientTransportPoolRandom) Submit(operation *runtime.ClientOperation) (any, error) {
	return c.TryTransportsForOp(operation)
}

func (c *ClientTransportPoolRandom) TryTransportsForOp(operation *runtime.ClientOperation) (any, error) {
	result, err := c.TryTransportForF(func(transport *ApiClientTransport) (any, error) {
		return transport.Submit(operation)
	})

	return result, err
}

func (c *ClientTransportPoolRandom) IterateRandomTransport() <-chan *ApiClientTransport {
	var transportsToTry []*cmap.Tuple[string, *ApiClientTransport]
	for tpl := range c.pool.IterBuffered() {
		transportsToTry = append(transportsToTry, &tpl)
	}

	ch := make(chan *ApiClientTransport, len(transportsToTry))

	go func() {
		for len(transportsToTry) > 0 {
			var transportTpl *cmap.Tuple[string, *ApiClientTransport]
			transportTpl, transportsToTry = selectAndRemoveRandom(transportsToTry, nil)
			ch <- transportTpl.Val
		}
	}()

	return ch
}

func (c *ClientTransportPoolRandom) TryTransportForF(cb func(*ApiClientTransport) (any, error)) (any, error) {
	//try active first if we have it
	active := c.GetActiveTransport()
	activeKey := ""

	if active != nil {
		activeKey = active.ApiUrl.String()
		result, err := cb(active)

		if err == nil {
			return result, err
		}

		if !errorIndicatesControllerSwap(err) {
			pfxlog.Logger().WithError(err).Debugf("determined that error (%T) does not indicate controller swap, returning error", err)
			return result, err
		}

		pfxlog.Logger().WithError(err).Debugf("encountered error (%T) while submitting request indicating controller swap", err)

		if c.pool.Count() == 1 {
			pfxlog.Logger().Debug("active transport failed, only 1 transport in pool")

			return result, err
		}
	}

	// either no active or active failed, lets start trying them at random
	pfxlog.Logger().Debug("trying random transports from pool")

	ch := c.IterateRandomTransport()

	var lastResult any
	lastErr := errors.New("no transports to try, active transport already failed or was nil") //default err should never be returned
	attempts := 0
	for transport := range ch {
		// skip the already attempted active key
		if activeKey != "" && transport.ApiUrl.String() == activeKey {
			continue
		}

		attempts = attempts + 1
		lastResult, lastErr = cb(transport)

		if lastErr == nil {
			c.SetActiveTransport(transport)
			return lastResult, nil
		}
	}

	return lastResult, lastErr
}

func (c *ClientTransportPoolRandom) AnyTransport() *ApiClientTransport {
	rand.Seed(uint64(time.Now().UnixNano()))
	transportBuffer := c.pool.Items()
	var keys []string

	for key := range transportBuffer {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil
	}
	index := rand.Intn(len(keys))
	return transportBuffer[keys[index]]
}

var _ runtime.ClientTransport = (*ClientTransportPoolRandom)(nil)
var _ ClientTransportPool = (*ClientTransportPoolRandom)(nil)

var opError = &net.OpError{}

func errorIndicatesControllerSwap(err error) bool {
	pfxlog.Logger().WithError(err).Debugf("checking for network errror on type (%T) and its wrapped errors", err)

	if errors.As(err, &opError) {
		pfxlog.Logger().Debug("detected net.OpError")
		return true
	}

	//others? rate limiting? http timeout?

	return false
}

func selectAndRemoveRandom[T any](slice []T, zero T) (selected T, modifiedSlice []T) {
	rand.Seed(uint64(time.Now().UnixNano()))
	if len(slice) == 0 {
		return zero, slice
	}
	index := rand.Intn(len(slice))
	selected = slice[index]
	modifiedSlice = append(slice[:index], slice[index+1:]...)
	return selected, modifiedSlice
}
