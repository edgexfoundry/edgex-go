// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// IntervalDeleter is an autogenerated mock type for the IntervalDeleter type
type IntervalDeleter struct {
	mock.Mock
}

// DeleteIntervalById provides a mock function with given fields: id
func (_m *IntervalDeleter) DeleteIntervalById(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
