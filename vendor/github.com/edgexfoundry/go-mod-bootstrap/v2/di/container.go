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

// di implements a simple generic dependency injection container.
//
// Sample usage:
//
//		package main
//
//		import (
//			"fmt"
//			"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
//		)
//
//		type foo struct {
//			FooMessage string
//		}
//
//		func NewFoo(m string) *foo {
//			return &foo{
//				FooMessage: m,
//			}
//		}
//
//		type bar struct {
//			BarMessage string
//			Foo        *foo
//		}
//
//		func NewBar(m string, foo *foo) *bar {
//			return &bar{
//				BarMessage: m,
//				Foo:        foo,
//			}
//		}
//
//		func main() {
//			container := di.NewContainer(
//				di.ServiceConstructorMap{
//					"foo": func(get di.Get) interface{} {
//						return NewFoo("fooMessage")
//					},
//					"bar": func(get di.Get) interface{} {
//						return NewBar("barMessage", get("foo").(*foo))
//					},
//				})
//
//			b := container.Get("bar").(*bar)
//			fmt.Println(b.BarMessage)
//			fmt.Println(b.Foo.FooMessage)
//		}
//
package di

import "sync"

type Get func(serviceName string) interface{}

// ServiceConstructor defines the contract for a function/closure to create a service.
type ServiceConstructor func(get Get) interface{}

// ServiceConstructorMap maps a service name to a function/closure to create that service.
type ServiceConstructorMap map[string]ServiceConstructor

// service is an internal structure used to track a specific service's constructor and constructed instance.
type service struct {
	constructor ServiceConstructor
	instance    interface{}
}

// Container is a receiver that maintains a list of services, their constructors, and their constructed instances in a
// thread-safe manner.
type Container struct {
	serviceMap map[string]service
	mutex      sync.RWMutex
}

// NewContainer is a factory method that returns an initialized Container receiver struct.
func NewContainer(serviceConstructors ServiceConstructorMap) *Container {
	c := Container{
		serviceMap: map[string]service{},
		mutex:      sync.RWMutex{},
	}
	if serviceConstructors != nil {
		c.Update(serviceConstructors)
	}
	return &c
}

// Set updates its internal serviceMap with the contents of the provided ServiceConstructorMap.
func (c *Container) Update(serviceConstructors ServiceConstructorMap) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for serviceName, constructor := range serviceConstructors {
		c.serviceMap[serviceName] = service{
			constructor: constructor,
			instance:    nil,
		}
	}
}

// get looks up the requested serviceName and, if it exists, returns a constructed instance.  If the requested service
// does not exist, it returns nil.  Get wraps instance construction in a singleton; the implementation assumes an instance,
// once constructed, will be reused and returned for all subsequent get(serviceName) calls.
func (c *Container) get(serviceName string) interface{} {
	service, ok := c.serviceMap[serviceName]
	if !ok {
		// Returning nil allows the DIC to be queried for a object and not panic if it doesn't exist.
		return nil
	}
	if service.instance == nil {
		service.instance = service.constructor(c.get)
		c.serviceMap[serviceName] = service
	}
	return service.instance
}

// Get wraps get to make it thread-safe.
func (c *Container) Get(serviceName string) interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.get(serviceName)
}
