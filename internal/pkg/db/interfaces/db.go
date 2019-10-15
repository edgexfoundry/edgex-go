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

package interfaces

import (
	correlation "github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DBClient interface {
	CloseSession()

	Events() ([]contract.Event, error)
	EventsWithLimit(limit int) ([]contract.Event, error)
	AddEvent(e correlation.Event) (string, error)
	UpdateEvent(e correlation.Event) error
	EventById(id string) (contract.Event, error)
	EventsByChecksum(checksum string) ([]contract.Event, error)
	EventCount() (int, error)
	EventCountByDeviceId(id string) (int, error)
	DeleteEventById(id string) error
	DeleteEventsByDevice(deviceId string) (int, error)
	EventsForDeviceLimit(id string, limit int) ([]contract.Event, error)
	EventsForDevice(id string) ([]contract.Event, error)
	EventsByCreationTime(startTime, endTime int64, limit int) ([]contract.Event, error)
	EventsOlderThanAge(age int64) ([]contract.Event, error)
	EventsPushed() ([]contract.Event, error)
	ScrubAllEvents() error

	Readings() ([]contract.Reading, error)
	AddReading(r contract.Reading) (string, error)
	UpdateReading(r contract.Reading) error
	ReadingById(id string) (contract.Reading, error)
	ReadingCount() (int, error)
	DeleteReadingById(id string) error
	DeleteReadingsByDevice(deviceId string) error
	ReadingsByDevice(id string, limit int) ([]contract.Reading, error)
	ReadingsByValueDescriptor(name string, limit int) ([]contract.Reading, error)
	ReadingsByValueDescriptorNames(names []string, limit int) ([]contract.Reading, error)
	ReadingsByCreationTime(start, end int64, limit int) ([]contract.Reading, error)
	ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]contract.Reading, error)

	ValueDescriptors() ([]contract.ValueDescriptor, error)
	AddValueDescriptor(v contract.ValueDescriptor) (string, error)
	UpdateValueDescriptor(cvd contract.ValueDescriptor) error
	DeleteValueDescriptorById(id string) error
	ValueDescriptorByName(name string) (contract.ValueDescriptor, error)
	ValueDescriptorsByName(names []string) ([]contract.ValueDescriptor, error)
	ValueDescriptorById(id string) (contract.ValueDescriptor, error)
	ValueDescriptorsByUomLabel(uomLabel string) ([]contract.ValueDescriptor, error)
	ValueDescriptorsByLabel(label string) ([]contract.ValueDescriptor, error)
	ValueDescriptorsByType(t string) ([]contract.ValueDescriptor, error)
	ScrubAllValueDescriptors() error

	Registrations() ([]contract.Registration, error)
	AddRegistration(r contract.Registration) (string, error)
	UpdateRegistration(reg contract.Registration) error
	RegistrationById(id string) (contract.Registration, error)
	RegistrationByName(name string) (contract.Registration, error)
	DeleteRegistrationById(id string) error
	DeleteRegistrationByName(name string) error
	ScrubAllRegistrations() error

	GetAllDeviceReports() ([]contract.DeviceReport, error)
	GetDeviceReportByName(n string) (contract.DeviceReport, error)
	GetDeviceReportByDeviceName(n string) ([]contract.DeviceReport, error)
	GetDeviceReportById(id string) (contract.DeviceReport, error)
	GetDeviceReportsByAction(n string) ([]contract.DeviceReport, error)
	AddDeviceReport(d contract.DeviceReport) (string, error)
	UpdateDeviceReport(dr contract.DeviceReport) error
	DeleteDeviceReportById(id string) error

	GetAllDevices() ([]contract.Device, error)
	AddDevice(d contract.Device, commands []contract.Command) (string, error)
	UpdateDevice(d contract.Device) error
	DeleteDeviceById(id string) error
	GetDevicesByProfileId(id string) ([]contract.Device, error)
	GetDeviceById(id string) (contract.Device, error)
	GetDeviceByName(n string) (contract.Device, error)
	GetDevicesByServiceId(id string) ([]contract.Device, error)
	GetDevicesWithLabel(l string) ([]contract.Device, error)

	GetAllDeviceProfiles() ([]contract.DeviceProfile, error)
	GetDeviceProfileById(id string) (contract.DeviceProfile, error)
	GetDeviceProfilesByModel(model string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesWithLabel(l string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesByManufacturerModel(man string, mod string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesByManufacturer(man string) ([]contract.DeviceProfile, error)
	GetDeviceProfileByName(n string) (contract.DeviceProfile, error)
	AddDeviceProfile(dp contract.DeviceProfile) (string, error)
	UpdateDeviceProfile(dp contract.DeviceProfile) error
	DeleteDeviceProfileById(id string) error

	GetAddressables() ([]contract.Addressable, error)
	UpdateAddressable(a contract.Addressable) error
	GetAddressableById(id string) (contract.Addressable, error)
	AddAddressable(a contract.Addressable) (string, error)
	GetAddressableByName(n string) (contract.Addressable, error)
	GetAddressablesByTopic(t string) ([]contract.Addressable, error)
	GetAddressablesByPort(p int) ([]contract.Addressable, error)
	GetAddressablesByPublisher(p string) ([]contract.Addressable, error)
	GetAddressablesByAddress(add string) ([]contract.Addressable, error)
	DeleteAddressableById(id string) error

	GetDeviceServiceByName(n string) (contract.DeviceService, error)
	GetDeviceServiceById(id string) (contract.DeviceService, error)
	GetAllDeviceServices() ([]contract.DeviceService, error)
	GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error)
	GetDeviceServicesWithLabel(l string) ([]contract.DeviceService, error)
	AddDeviceService(ds contract.DeviceService) (string, error)
	UpdateDeviceService(ds contract.DeviceService) error
	DeleteDeviceServiceById(id string) error

	GetAllProvisionWatchers() (pw []contract.ProvisionWatcher, err error)
	GetProvisionWatcherByName(n string) (pw contract.ProvisionWatcher, err error)
	GetProvisionWatchersByIdentifier(k string, v string) (pw []contract.ProvisionWatcher, err error)
	GetProvisionWatchersByServiceId(id string) (pw []contract.ProvisionWatcher, err error)
	GetProvisionWatchersByProfileId(id string) (pw []contract.ProvisionWatcher, err error)
	GetProvisionWatcherById(id string) (pw contract.ProvisionWatcher, err error)
	AddProvisionWatcher(pw contract.ProvisionWatcher) (string, error)
	UpdateProvisionWatcher(pw contract.ProvisionWatcher) error
	DeleteProvisionWatcherById(id string) error

	GetAllCommands() ([]contract.Command, error)
	GetCommandById(id string) (contract.Command, error)
	GetCommandsByName(n string) ([]contract.Command, error)
	GetCommandsByDeviceId(did string) ([]contract.Command, error)
	GetCommandByNameAndDeviceId(cname string, did string) (contract.Command, error)

	ScrubMetadata() error

	GetNotifications() ([]contract.Notification, error)
	GetNotificationById(id string) (contract.Notification, error)
	GetNotificationBySlug(slug string) (contract.Notification, error)
	GetNotificationBySender(sender string, limit int) ([]contract.Notification, error)
	GetNotificationsByLabels(labels []string, limit int) ([]contract.Notification, error)
	GetNotificationsByStartEnd(start int64, end int64, limit int) ([]contract.Notification, error)
	GetNotificationsByStart(start int64, limit int) ([]contract.Notification, error)
	GetNotificationsByEnd(end int64, limit int) ([]contract.Notification, error)
	GetNewNotifications(limit int) ([]contract.Notification, error)
	GetNewNormalNotifications(limit int) ([]contract.Notification, error)
	AddNotification(n contract.Notification) (string, error)
	UpdateNotification(n contract.Notification) error
	MarkNotificationProcessed(n contract.Notification) error
	DeleteNotificationById(id string) error
	DeleteNotificationBySlug(slug string) error
	DeleteNotificationsOld(age int) error

	GetSubscriptionBySlug(slug string) (contract.Subscription, error)
	GetSubscriptionByCategories(categories []string) ([]contract.Subscription, error)
	GetSubscriptionByLabels(labels []string) ([]contract.Subscription, error)
	GetSubscriptionByCategoriesLabels(categories []string, labels []string) ([]contract.Subscription, error)
	GetSubscriptionByReceiver(receiver string) ([]contract.Subscription, error)
	GetSubscriptionById(id string) (contract.Subscription, error)
	DeleteSubscriptionById(id string) error
	AddSubscription(sub contract.Subscription) (string, error)
	UpdateSubscription(sub contract.Subscription) error
	DeleteSubscriptionBySlug(slug string) error
	GetSubscriptions() ([]contract.Subscription, error)

	AddTransmission(t contract.Transmission) (string, error)
	UpdateTransmission(t contract.Transmission) error
	DeleteTransmission(age int64, status contract.TransmissionStatus) error
	GetTransmissionById(id string) (contract.Transmission, error)
	GetTransmissionsByNotificationSlug(slug string, limit int) ([]contract.Transmission, error)
	GetTransmissionsByNotificationSlugAndStartEnd(slug string, start int64, end int64, limit int) ([]contract.Transmission, error)
	GetTransmissionsByStartEnd(start int64, end int64, limit int) ([]contract.Transmission, error)
	GetTransmissionsByStart(start int64, limit int) ([]contract.Transmission, error)
	GetTransmissionsByEnd(end int64, limit int) ([]contract.Transmission, error)
	GetTransmissionsByStatus(limit int, status contract.TransmissionStatus) ([]contract.Transmission, error)

	Cleanup() error
	CleanupOld(age int) error

	Intervals() ([]contract.Interval, error)
	IntervalsWithLimit(limit int) ([]contract.Interval, error)
	IntervalByName(name string) (contract.Interval, error)
	IntervalById(id string) (contract.Interval, error)
	AddInterval(interval contract.Interval) (string, error)
	UpdateInterval(interval contract.Interval) error
	DeleteIntervalById(id string) error

	IntervalActions() ([]contract.IntervalAction, error)
	IntervalActionsWithLimit(limit int) ([]contract.IntervalAction, error)
	IntervalActionsByIntervalName(name string) ([]contract.IntervalAction, error)
	IntervalActionsByTarget(name string) ([]contract.IntervalAction, error)
	IntervalActionById(id string) (contract.IntervalAction, error)
	IntervalActionByName(name string) (contract.IntervalAction, error)
	AddIntervalAction(action contract.IntervalAction) (string, error)
	UpdateIntervalAction(action contract.IntervalAction) error
	DeleteIntervalActionById(id string) error

	ScrubAllIntervalActions() (int, error)
	ScrubAllIntervals() (int, error)
}
