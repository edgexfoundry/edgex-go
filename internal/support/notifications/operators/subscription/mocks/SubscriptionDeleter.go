// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// SubscriptionDeleter is an autogenerated mock type for the SubscriptionDeleter type
type SubscriptionDeleter struct {
	mock.Mock
}

// DeleteSubscriptionById provides a mock function with given fields: id
func (_m *SubscriptionDeleter) DeleteSubscriptionById(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteSubscriptionBySlug provides a mock function with given fields: id
func (_m *SubscriptionDeleter) DeleteSubscriptionBySlug(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
