// Code generated by mockery v2.49.0. DO NOT EDIT.

package mocks

import (
	errors "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	mock "github.com/stretchr/testify/mock"
)

// Crypto is an autogenerated mock type for the Crypto type
type Crypto struct {
	mock.Mock
}

// Decrypt provides a mock function with given fields: _a0
func (_m *Crypto) Decrypt(_a0 string) ([]byte, errors.EdgeX) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Decrypt")
	}

	var r0 []byte
	var r1 errors.EdgeX
	if rf, ok := ret.Get(0).(func(string) ([]byte, errors.EdgeX)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string) errors.EdgeX); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// Encrypt provides a mock function with given fields: _a0
func (_m *Crypto) Encrypt(_a0 string) (string, errors.EdgeX) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Encrypt")
	}

	var r0 string
	var r1 errors.EdgeX
	if rf, ok := ret.Get(0).(func(string) (string, errors.EdgeX)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) errors.EdgeX); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// NewCrypto creates a new instance of Crypto. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCrypto(t interface {
	mock.TestingT
	Cleanup(func())
}) *Crypto {
	mock := &Crypto{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}