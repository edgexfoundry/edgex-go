// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	errors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	mock "github.com/stretchr/testify/mock"

	models "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// DBClient is an autogenerated mock type for the DBClient type
type DBClient struct {
	mock.Mock
}

// AddNotification provides a mock function with given fields: n
func (_m *DBClient) AddNotification(n models.Notification) (models.Notification, errors.EdgeX) {
	ret := _m.Called(n)

	var r0 models.Notification
	if rf, ok := ret.Get(0).(func(models.Notification) models.Notification); ok {
		r0 = rf(n)
	} else {
		r0 = ret.Get(0).(models.Notification)
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(models.Notification) errors.EdgeX); ok {
		r1 = rf(n)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// AddSubscription provides a mock function with given fields: e
func (_m *DBClient) AddSubscription(e models.Subscription) (models.Subscription, errors.EdgeX) {
	ret := _m.Called(e)

	var r0 models.Subscription
	if rf, ok := ret.Get(0).(func(models.Subscription) models.Subscription); ok {
		r0 = rf(e)
	} else {
		r0 = ret.Get(0).(models.Subscription)
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(models.Subscription) errors.EdgeX); ok {
		r1 = rf(e)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// AllSubscriptions provides a mock function with given fields: offset, limit
func (_m *DBClient) AllSubscriptions(offset int, limit int) ([]models.Subscription, errors.EdgeX) {
	ret := _m.Called(offset, limit)

	var r0 []models.Subscription
	if rf, ok := ret.Get(0).(func(int, int) []models.Subscription); ok {
		r0 = rf(offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Subscription)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int) errors.EdgeX); ok {
		r1 = rf(offset, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// CloseSession provides a mock function with given fields:
func (_m *DBClient) CloseSession() {
	_m.Called()
}

// DeleteNotificationById provides a mock function with given fields: id
func (_m *DBClient) DeleteNotificationById(id string) errors.EdgeX {
	ret := _m.Called(id)

	var r0 errors.EdgeX
	if rf, ok := ret.Get(0).(func(string) errors.EdgeX); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(errors.EdgeX)
		}
	}

	return r0
}

// DeleteSubscriptionByName provides a mock function with given fields: name
func (_m *DBClient) DeleteSubscriptionByName(name string) errors.EdgeX {
	ret := _m.Called(name)

	var r0 errors.EdgeX
	if rf, ok := ret.Get(0).(func(string) errors.EdgeX); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(errors.EdgeX)
		}
	}

	return r0
}

// NotificationById provides a mock function with given fields: id
func (_m *DBClient) NotificationById(id string) (models.Notification, errors.EdgeX) {
	ret := _m.Called(id)

	var r0 models.Notification
	if rf, ok := ret.Get(0).(func(string) models.Notification); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(models.Notification)
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(string) errors.EdgeX); ok {
		r1 = rf(id)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// NotificationsByCategory provides a mock function with given fields: offset, limit, category
func (_m *DBClient) NotificationsByCategory(offset int, limit int, category string) ([]models.Notification, errors.EdgeX) {
	ret := _m.Called(offset, limit, category)

	var r0 []models.Notification
	if rf, ok := ret.Get(0).(func(int, int, string) []models.Notification); ok {
		r0 = rf(offset, limit, category)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Notification)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int, string) errors.EdgeX); ok {
		r1 = rf(offset, limit, category)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// NotificationsByLabel provides a mock function with given fields: offset, limit, label
func (_m *DBClient) NotificationsByLabel(offset int, limit int, label string) ([]models.Notification, errors.EdgeX) {
	ret := _m.Called(offset, limit, label)

	var r0 []models.Notification
	if rf, ok := ret.Get(0).(func(int, int, string) []models.Notification); ok {
		r0 = rf(offset, limit, label)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Notification)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int, string) errors.EdgeX); ok {
		r1 = rf(offset, limit, label)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// NotificationsByStatus provides a mock function with given fields: offset, limit, status
func (_m *DBClient) NotificationsByStatus(offset int, limit int, status string) ([]models.Notification, errors.EdgeX) {
	ret := _m.Called(offset, limit, status)

	var r0 []models.Notification
	if rf, ok := ret.Get(0).(func(int, int, string) []models.Notification); ok {
		r0 = rf(offset, limit, status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Notification)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int, string) errors.EdgeX); ok {
		r1 = rf(offset, limit, status)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// NotificationsByTimeRange provides a mock function with given fields: start, end, offset, limit
func (_m *DBClient) NotificationsByTimeRange(start int, end int, offset int, limit int) ([]models.Notification, errors.EdgeX) {
	ret := _m.Called(start, end, offset, limit)

	var r0 []models.Notification
	if rf, ok := ret.Get(0).(func(int, int, int, int) []models.Notification); ok {
		r0 = rf(start, end, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Notification)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int, int, int) errors.EdgeX); ok {
		r1 = rf(start, end, offset, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// SubscriptionById provides a mock function with given fields: id
func (_m *DBClient) SubscriptionById(id string) (models.Subscription, errors.EdgeX) {
	ret := _m.Called(id)

	var r0 models.Subscription
	if rf, ok := ret.Get(0).(func(string) models.Subscription); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(models.Subscription)
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(string) errors.EdgeX); ok {
		r1 = rf(id)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// SubscriptionByName provides a mock function with given fields: name
func (_m *DBClient) SubscriptionByName(name string) (models.Subscription, errors.EdgeX) {
	ret := _m.Called(name)

	var r0 models.Subscription
	if rf, ok := ret.Get(0).(func(string) models.Subscription); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(models.Subscription)
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(string) errors.EdgeX); ok {
		r1 = rf(name)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// SubscriptionsByCategory provides a mock function with given fields: offset, limit, category
func (_m *DBClient) SubscriptionsByCategory(offset int, limit int, category string) ([]models.Subscription, errors.EdgeX) {
	ret := _m.Called(offset, limit, category)

	var r0 []models.Subscription
	if rf, ok := ret.Get(0).(func(int, int, string) []models.Subscription); ok {
		r0 = rf(offset, limit, category)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Subscription)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int, string) errors.EdgeX); ok {
		r1 = rf(offset, limit, category)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// SubscriptionsByLabel provides a mock function with given fields: offset, limit, label
func (_m *DBClient) SubscriptionsByLabel(offset int, limit int, label string) ([]models.Subscription, errors.EdgeX) {
	ret := _m.Called(offset, limit, label)

	var r0 []models.Subscription
	if rf, ok := ret.Get(0).(func(int, int, string) []models.Subscription); ok {
		r0 = rf(offset, limit, label)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Subscription)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int, string) errors.EdgeX); ok {
		r1 = rf(offset, limit, label)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// SubscriptionsByReceiver provides a mock function with given fields: offset, limit, receiver
func (_m *DBClient) SubscriptionsByReceiver(offset int, limit int, receiver string) ([]models.Subscription, errors.EdgeX) {
	ret := _m.Called(offset, limit, receiver)

	var r0 []models.Subscription
	if rf, ok := ret.Get(0).(func(int, int, string) []models.Subscription); ok {
		r0 = rf(offset, limit, receiver)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Subscription)
		}
	}

	var r1 errors.EdgeX
	if rf, ok := ret.Get(1).(func(int, int, string) errors.EdgeX); ok {
		r1 = rf(offset, limit, receiver)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(errors.EdgeX)
		}
	}

	return r0, r1
}

// UpdateSubscription provides a mock function with given fields: s
func (_m *DBClient) UpdateSubscription(s models.Subscription) errors.EdgeX {
	ret := _m.Called(s)

	var r0 errors.EdgeX
	if rf, ok := ret.Get(0).(func(models.Subscription) errors.EdgeX); ok {
		r0 = rf(s)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(errors.EdgeX)
		}
	}

	return r0
}
