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
	"errors"
	"math/rand/v2"
	"net"
	"net/url"
	"slices"
	"sync/atomic"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/michaelquigley/pfxlog"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// ApiClientTransport wraps a runtime.ClientTransport with its associated API URL,
// enabling tracking of which controller endpoint a transport communicates with.
type ApiClientTransport struct {
	runtime.ClientTransport
	ApiUrl *url.URL
}

// ClientTransportPool manages multiple runtime.ClientTransport instances representing
// different controller endpoints in a high-availability OpenZiti network. It provides
// automatic failover capabilities when individual controllers become unavailable.
type ClientTransportPool interface {
	runtime.ClientTransport

	// Add registers a new transport for the specified API URL.
	Add(apiUrl *url.URL, transport runtime.ClientTransport)

	// Remove unregisters the transport for the specified API URL.
	Remove(apiUrl *url.URL)

	// GetActiveTransport returns the currently selected transport.
	GetActiveTransport() *ApiClientTransport

	// SetActiveTransport designates which transport to use for subsequent operations.
	SetActiveTransport(*ApiClientTransport)

	// GetApiUrls returns all registered API URLs.
	GetApiUrls() []*url.URL

	// IterateTransportsRandomly provides a channel for iterating through available transports
	// in random order.
	IterateTransportsRandomly() chan<- *ApiClientTransport

	// TryTransportsForOp attempts to execute an operation, trying different transports
	// on connection failures.
	TryTransportsForOp(operation *runtime.ClientOperation) (any, error)

	// TryTransportForF executes a callback function, trying different transports
	// on connection failures.
	TryTransportForF(cb func(*ApiClientTransport) (any, error)) (any, error)
}

var _ runtime.ClientTransport = (ClientTransportPool)(nil)
var _ ClientTransportPool = (*ClientTransportPoolRandom)(nil)

// ClientTransportPoolRandom implements a randomized failover strategy for controller selection.
// It maintains an active transport and switches to randomly selected alternatives when the active
// transport becomes unreachable.
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

// GetApiClientTransports returns a snapshot of all registered transports.
func (c *ClientTransportPoolRandom) GetApiClientTransports() []*ApiClientTransport {
	var result []*ApiClientTransport

	for tpl := range c.pool.IterBuffered() {
		result = append(result, tpl.Val)
	}

	return result
}

// NewClientTransportPoolRandom creates a new transport pool with randomized failover.
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

func (c *ClientTransportPoolRandom) IterateRandomTransport() []*ApiClientTransport {
	var result []*ApiClientTransport
	c.pool.IterCb(func(_ string, v *ApiClientTransport) {
		result = append(result, v)
	})

	Randomize(result)
	return result
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

	transports := c.IterateRandomTransport()

	var lastResult any
	lastErr := errors.New("no transports to try, active transport already failed or was nil") //default err should never be returned
	attempts := 0
	for _, transport := range transports {
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

// AnyTransport returns a randomly selected transport from the pool, or nil if empty.
func (c *ClientTransportPoolRandom) AnyTransport() *ApiClientTransport {
	transportBuffer := c.pool.Items()
	var keys []string

	for key := range transportBuffer {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil
	}
	seed := uint64(time.Now().UnixNano())
	rng := rand.New(rand.NewPCG(seed, seed))
	index := rng.IntN(len(keys))
	return transportBuffer[keys[index]]
}

var _ runtime.ClientTransport = (*ClientTransportPoolRandom)(nil)
var _ ClientTransportPool = (*ClientTransportPoolRandom)(nil)

var opError = &net.OpError{}

// errorIndicatesControllerSwap determines whether an error suggests the need to
// switch to a different controller endpoint.
func errorIndicatesControllerSwap(err error) bool {
	pfxlog.Logger().WithError(err).Debugf("checking for network errror on type (%T) and its wrapped errors", err)

	if errors.As(err, &opError) {
		pfxlog.Logger().Debug("detected net.OpError")
		return true
	}

	//others? rate limiting? http timeout?

	return false
}

func Randomize[T any](s []T) {
	for i := 0; i < len(s); i++ {
		idx := rand.IntN(len(s))
		e1 := s[i]
		e2 := s[idx]
		s[i] = e2
		s[idx] = e1
	}
}

func selectAndRemoveRandom[T any](slice []T, zero T) (selected T, modifiedSlice []T) {
	if len(slice) == 0 {
		return zero, slice
	}
	seed := uint64(time.Now().UnixNano())
	rng := rand.New(rand.NewPCG(seed, seed))
	index := rng.IntN(len(slice))
	selected = slice[index]
	return selected, slices.Delete(slice, index, index+1)
}
