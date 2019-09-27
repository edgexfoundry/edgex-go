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

package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const serviceName = "serviceName"

func TestGetUnknownServicePanics(t *testing.T) {
	sut := NewContainer(ServiceConstructorMap{})

	assert.Panics(t, func() { sut.Get("unknownService") })
}

func TestGetKnownServiceReturnsExpectedConstructorResult(t *testing.T) {
	type serviceType struct{}
	var service serviceType
	var serviceConstructor = func(get Get) interface{} { return service }
	sut := NewContainer(ServiceConstructorMap{serviceName: serviceConstructor})

	result := sut.Get(serviceName)

	assert.Equal(t, service, result)
}

func TestGetKnownServiceImplementsSingleton(t *testing.T) {
	type serviceType struct {
		value int
	}
	instanceCount := 0
	var serviceConstructor = func(get Get) interface{} {
		instanceCount += 1
		return serviceType{value: instanceCount}
	}
	sut := NewContainer(ServiceConstructorMap{serviceName: serviceConstructor})

	first := sut.Get(serviceName)
	second := sut.Get(serviceName)

	assert.Equal(t, first.(serviceType).value, second.(serviceType).value)
}

func TestUpdateOfNonExistentServiceAdds(t *testing.T) {
	type serviceType struct{}
	var service serviceType
	var serviceConstructor = func(get Get) interface{} { return service }
	sut := NewContainer(ServiceConstructorMap{})
	sut.Update(ServiceConstructorMap{serviceName: serviceConstructor})

	result := sut.Get(serviceName)

	assert.Equal(t, service, result)
}

func TestUpdateOfExistingServiceReplaces(t *testing.T) {
	const original = "original"
	const replacement = "replacement"
	type serviceType struct{ value string }
	var originalConstructor = func(get Get) interface{} { return serviceType{value: original} }
	sut := NewContainer(ServiceConstructorMap{serviceName: originalConstructor})
	var replacementConstructor = func(get Get) interface{} { return serviceType{value: replacement} }
	sut.Update(ServiceConstructorMap{serviceName: replacementConstructor})

	result := sut.Get(serviceName)

	assert.Equal(t, replacement, result.(serviceType).value)
}

func TestGetInsideGetReturnsAsExpected(t *testing.T) {
	const fooName = "foo"
	type foo struct {
		FooMessage string
	}

	const barName = "bar"
	type bar struct {
		BarMessage string
		Foo        *foo
	}

	sut := NewContainer(ServiceConstructorMap{
		fooName: func(get Get) interface{} { return &foo{FooMessage: fooName} },
		barName: func(get Get) interface{} { return &bar{BarMessage: barName, Foo: get(fooName).(*foo)} },
	})

	result := sut.Get(barName).(*bar)

	assert.Equal(t, barName, result.BarMessage)
	assert.NotNil(t, result.Foo)
	assert.Equal(t, fooName, result.Foo.FooMessage)
}
