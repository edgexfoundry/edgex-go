// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import models "github.com/edgexfoundry/go-mod-core-contracts/v2/models"

// Loader is an autogenerated mock type for the Loader type
type Loader struct {
	mock.Mock
}

// ValueDescriptors provides a mock function with given fields:
func (_m *Loader) ValueDescriptors() ([]models.ValueDescriptor, error) {
	ret := _m.Called()

	var r0 []models.ValueDescriptor
	if rf, ok := ret.Get(0).(func() []models.ValueDescriptor); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.ValueDescriptor)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValueDescriptorsByName provides a mock function with given fields: names
func (_m *Loader) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	ret := _m.Called(names)

	var r0 []models.ValueDescriptor
	if rf, ok := ret.Get(0).(func([]string) []models.ValueDescriptor); ok {
		r0 = rf(names)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.ValueDescriptor)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(names)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
