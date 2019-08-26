// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import models "github.com/edgexfoundry/go-mod-core-contracts/models"

// Distributor is an autogenerated mock type for the Distributor type
type Distributor struct {
	mock.Mock
}

// DistributeAndMark provides a mock function with given fields: n
func (_m *Distributor) DistributeAndMark(n models.Notification) error {
	ret := _m.Called(n)

	var r0 error
	if rf, ok := ret.Get(0).(func(models.Notification) error); ok {
		r0 = rf(n)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
