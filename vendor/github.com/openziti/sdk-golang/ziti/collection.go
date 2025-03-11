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

package ziti

import (
	"context"
	"github.com/michaelquigley/pfxlog"
	cmap "github.com/orcaman/concurrent-map/v2"
	"net"
	"os"
	"strings"
)

// An CtxCollection allows Context instances to be instantiated and maintained as a group. Useful in scenarios
// where multiple Context instances are managed together. Instead of using ziti.NewContext() like functions, use
// the function provided on this type to automatically have contexts added as they are created. If ConfigTypes
// is set, they will be automatically added to any instantiated Context through `New*` functions.
//
// Context instances can be created directly from CtxCollection instances. Doing so automatically adds new Context
// instances to the CtxCollection:
//
// ```
// collection := ziti.NewCtxCollection()
// cfg, err := ziti.NewConfigFromFile(str)
// ctx, err := collection.NewContext(cfg) //return the ctx and adds it to the collection
// ```
//
// If more control over Config and Context instantiation is desired before the Context is added to the collection
// the Add() function can be used:
//
// ```
// collection := ziti.NewCtxCollection()
// cfg, err := ziti.NewConfigFromFile(str)
// ctx, err := ziti.NewContext(cfg)
// collection.Add(ctx) //manual collection add
// ```
type CtxCollection struct {
	contexts    cmap.ConcurrentMap[string, Context]
	ConfigTypes []string
}

// NewSdkCollection creates a new empty collection.
func NewSdkCollection() *CtxCollection {
	return &CtxCollection{
		contexts: cmap.New[Context](),
	}
}

// NewSdkCollectionFromEnv will create an empty CtxCollection and then attempt to populate it from configuration files
// provided in a semicolon separate list of file paths retrieved from an environment variable.
func NewSdkCollectionFromEnv(envVariable string, configTypes ...string) *CtxCollection {
	collection := NewSdkCollection()

	envValue := os.Getenv(envVariable)

	identityFiles := strings.Split(envValue, ";")

	for _, identityFile := range identityFiles {

		if identityFile == "" {
			continue
		}

		cfg, err := NewConfigFromFile(identityFile)

		if err != nil {
			pfxlog.Logger().Errorf("failed to load config from file '%s'", identityFile)
			continue
		}

		cfg.ConfigTypes = append(cfg.ConfigTypes, configTypes...)

		//collection.NewContext stores the new ctx in its internal collection
		_, err = collection.NewContext(cfg)

		if err != nil {
			pfxlog.Logger().Errorf("failed to create context from '%s'", identityFile)
			continue
		}
	}

	return collection
}

// Add allows the arbitrary idempotent inclusion of a Context in the current collection. If a Context with the same id
// as an existing Context is added and is a different instance, the original is closed and removed.
func (set *CtxCollection) Add(ctx Context) {
	set.contexts.Upsert(ctx.GetId(), ctx, func(exist bool, valueInMap Context, newValue Context) Context {
		if exist && valueInMap != nil && valueInMap != newValue {
			valueInMap.Close()
		}

		return newValue
	})
}

// Remove removes the supplied Context from the collection. It is not closed or altered in any way.
func (set *CtxCollection) Remove(ctx Context) {
	set.contexts.Remove(ctx.GetId())
}

// RemoveById removes a context by its string id.  It is not closed or altered in any way.
func (set *CtxCollection) RemoveById(id string) {
	set.contexts.Remove(id)
}

// ForAll call the provided function `f` on each Context.
func (set *CtxCollection) ForAll(f func(ctx Context)) {
	set.contexts.IterCb(func(key string, ctx Context) {
		f(ctx)
	})
}

// NewContextFromFile is the same as ziti.NewContextFromFile but will also add the resulting
// context to the current collection.
func (set *CtxCollection) NewContextFromFile(file string) (Context, error) {
	return set.NewContextFromFileWithOpts(file, nil)
}

// NewContextFromFileWithOpts is the same as ziti.NewContextFromFileWithOpts but will also add
// the resulting context to the current collection.
func (set *CtxCollection) NewContextFromFileWithOpts(file string, options *Options) (Context, error) {
	cfg, err := NewConfigFromFile(file)

	if err != nil {
		return nil, err
	}

	return set.NewContextWithOpts(cfg, options)
}

// NewContext is the same as ziti.NewContext but will also add the resulting context to the current collection.
func (set *CtxCollection) NewContext(cfg *Config) (Context, error) {
	return set.NewContextWithOpts(cfg, nil)
}

// NewContextWithOpts is the same as ziti.NewContextWithOpts but will also add the resulting context to the current
// collection.
func (set *CtxCollection) NewContextWithOpts(cfg *Config, options *Options) (Context, error) {
	cfg.ConfigTypes = append(cfg.ConfigTypes, set.ConfigTypes...)

	ctx, err := NewContextWithOpts(cfg, options)

	if err != nil {
		return nil, err
	}

	set.Add(ctx)

	return ctx, nil
}

// NewDialer will return a dialer that will iterate over the Context instances inside the collection, searching for the
// context that best matches the service.
//
// If a matching service is not found, an error is returned. Matching is based on Match() logic in edge.InterceptV1Config.
func (set *CtxCollection) NewDialer() Dialer {
	return &dialer{
		collection: set,
	}
}

// NewDialerWithFallback will return a dialer that will iterate over the Context instances inside the collection,
// searching for the context that best matches the service.
//
// If a matching service is not found, a dial is attempted using the fallback dialer. Matching is based on Match() logic
// in edge.InterceptV1Config.
func (set *CtxCollection) NewDialerWithFallback(ctx context.Context, fallback Dialer) Dialer {
	if fallback == nil {
		fallback = &net.Dialer{}
	}
	return &dialer{
		fallback:   fallback,
		context:    ctx,
		collection: set,
	}
}
