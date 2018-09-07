/*******************************************************************************
 * Copyright 2018 Cavium
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

package memory

import (
	"errors"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

// Schedule event
func (m *MemDB) GetAllScheduleEvents(se *[]models.ScheduleEvent) error {
	cpy := make([]models.ScheduleEvent, len(m.scheduleEvents))
	copy(cpy, m.scheduleEvents)
	*se = cpy
	return nil
}

func (m *MemDB) AddScheduleEvent(se *models.ScheduleEvent) error {
	currentTime := db.MakeTimestamp()
	se.Created = currentTime
	se.Modified = currentTime
	se.Id = bson.NewObjectId()

	for _, s := range m.scheduleEvents {
		if s.Name == se.Name {
			return db.ErrNotUnique
		}
	}

	validAddressable := false
	// Test addressable id or name exists
	for _, a := range m.addressables {
		if a.Name == se.Addressable.Name {
			validAddressable = true
			break
		}
		if a.Id == se.Addressable.Id {
			validAddressable = true
			break
		}
	}

	if !validAddressable {
		return errors.New("Invalid addressable")
	}

	m.scheduleEvents = append(m.scheduleEvents, *se)
	return nil
}

func (m *MemDB) GetScheduleEventByName(se *models.ScheduleEvent, n string) error {
	for _, s := range m.scheduleEvents {
		if s.Name == n {
			err := m.GetAddressableById(&s.Addressable, s.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for ds %s",
					se.Addressable.Id.Hex(), se.Id.Hex())
			}
			*se = s
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) UpdateScheduleEvent(se models.ScheduleEvent) error {
	for i, s := range m.scheduleEvents {
		if s.Id == se.Id {
			m.scheduleEvents[i] = se
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetScheduleEventById(se *models.ScheduleEvent, id string) error {
	for _, s := range m.scheduleEvents {
		if s.Id.Hex() == id {
			err := m.GetAddressableById(&s.Addressable, s.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for ds %s",
					se.Addressable.Id.Hex(), se.Id.Hex())
			}
			*se = s
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetScheduleEventsByScheduleName(ses *[]models.ScheduleEvent, n string) error {
	l := []models.ScheduleEvent{}
	for _, se := range m.scheduleEvents {
		if se.Schedule == n {
			err := m.GetAddressableById(&se.Addressable, se.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for se %s",
					se.Addressable.Id.Hex(), se.Id.Hex())
			}
			l = append(l, se)
		}
	}
	*ses = l
	return nil
}

func (m *MemDB) GetScheduleEventsByAddressableId(ses *[]models.ScheduleEvent, id string) error {
	l := []models.ScheduleEvent{}
	for _, se := range m.scheduleEvents {
		if se.Addressable.Id.Hex() == id {
			err := m.GetAddressableById(&se.Addressable, se.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for se %s",
					se.Addressable.Id.Hex(), se.Id.Hex())
			}
			l = append(l, se)
		}
	}
	*ses = l
	return nil
}

func (m *MemDB) GetScheduleEventsByServiceName(ses *[]models.ScheduleEvent, n string) error {
	l := []models.ScheduleEvent{}
	for _, se := range m.scheduleEvents {
		if se.Service == n {
			err := m.GetAddressableById(&se.Addressable, se.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for se %s",
					se.Addressable.Id.Hex(), se.Id.Hex())
			}
			l = append(l, se)
		}
	}
	*ses = l
	return nil
}

func (m *MemDB) DeleteScheduleEventById(id string) error {
	for i, s := range m.scheduleEvents {
		if s.Id.Hex() == id {
			m.scheduleEvents = append(m.scheduleEvents[:i], m.scheduleEvents[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

// Schedule
func (m *MemDB) GetAllSchedules(s *[]models.Schedule) error {
	cpy := make([]models.Schedule, len(m.schedules))
	copy(cpy, m.schedules)
	*s = cpy
	return nil
}

func (m *MemDB) AddSchedule(s *models.Schedule) error {
	currentTime := db.MakeTimestamp()
	s.Created = currentTime
	s.Modified = currentTime
	s.Id = bson.NewObjectId()

	for _, ss := range m.schedules {
		if ss.Name == s.Name {
			return db.ErrNotUnique
		}
	}

	m.schedules = append(m.schedules, *s)
	return nil
}

func (m *MemDB) GetScheduleByName(s *models.Schedule, n string) error {
	for _, ss := range m.schedules {
		if ss.Name == n {
			*s = ss
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) UpdateSchedule(s models.Schedule) error {
	s.Modified = db.MakeTimestamp()
	for i, ss := range m.schedules {
		if ss.Id == s.Id {
			m.schedules[i] = s
			return nil
		}
	}

	return db.ErrNotFound
}

func (m *MemDB) GetScheduleById(s *models.Schedule, id string) error {
	for _, ss := range m.schedules {
		if ss.Id.Hex() == id {
			*s = ss
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) DeleteScheduleById(id string) error {
	for i, ss := range m.schedules {
		if ss.Id.Hex() == id {
			m.schedules = append(m.schedules[:i], m.schedules[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

// Device Report
func (m *MemDB) GetAllDeviceReports(drs *[]models.DeviceReport) error {
	cpy := make([]models.DeviceReport, len(m.deviceReports))
	copy(cpy, m.deviceReports)
	*drs = cpy
	return nil
}

func (m *MemDB) GetDeviceReportByDeviceName(drs *[]models.DeviceReport, n string) error {
	l := []models.DeviceReport{}
	for _, dr := range m.deviceReports {
		if dr.Name == n {
			l = append(l, dr)
		}
	}
	*drs = l
	return nil
}

func (m *MemDB) GetDeviceReportByName(dr *models.DeviceReport, n string) error {
	for _, d := range m.deviceReports {
		if d.Name == n {
			*dr = d
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetDeviceReportById(dr *models.DeviceReport, id string) error {
	for _, d := range m.deviceReports {
		if d.Id.Hex() == id {
			*dr = d
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) AddDeviceReport(dr *models.DeviceReport) error {
	currentTime := db.MakeTimestamp()
	dr.Created = currentTime
	dr.Modified = currentTime
	dr.Id = bson.NewObjectId()

	dummy := models.DeviceReport{}
	if m.GetDeviceReportByName(&dummy, dr.Name) == nil {
		return db.ErrNotUnique
	}

	m.deviceReports = append(m.deviceReports, *dr)
	return nil

}

func (m *MemDB) UpdateDeviceReport(dr *models.DeviceReport) error {
	for i, d := range m.deviceReports {
		if d.Id == dr.Id {
			m.deviceReports[i] = *dr
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetDeviceReportsByScheduleEventName(drs *[]models.DeviceReport, n string) error {
	l := []models.DeviceReport{}
	for _, dr := range m.deviceReports {
		if dr.Event == n {
			l = append(l, dr)
		}
	}
	*drs = l
	return nil
}

func (m *MemDB) DeleteDeviceReportById(id string) error {
	for i, c := range m.deviceReports {
		if c.Id.Hex() == id {
			m.deviceReports = append(m.deviceReports[:i], m.deviceReports[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

// Device
func (m *MemDB) updateDeviceValues(d *models.Device) error {
	err := m.GetAddressableById(&d.Addressable, d.Addressable.Id.Hex())
	if err != nil {
		return fmt.Errorf("Could not find addressable %s for ds %s",
			d.Addressable.Id.Hex(), d.Id.Hex())
	}
	err = m.GetDeviceServiceById(&d.Service, d.Service.Id.Hex())
	if err != nil {
		return fmt.Errorf("Could not find DeviceService %s for ds %s",
			d.Service.Id.Hex(), d.Id.Hex())
	}
	err = m.GetDeviceProfileById(&d.Profile, d.Profile.Id.Hex())
	if err != nil {
		return fmt.Errorf("Could not find DeviceProfile %s for ds %s",
			d.Profile.Id.Hex(), d.Id.Hex())
	}
	return nil
}

type deviceCmp func(models.Device) bool

func (m *MemDB) getDeviceBy(d *models.Device, f deviceCmp) error {
	for _, dd := range m.devices {
		if f(dd) {
			if err := m.updateDeviceValues(&dd); err != nil {
				return err
			}
			*d = dd
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) getDevicesBy(d *[]models.Device, f deviceCmp) error {
	l := []models.Device{}
	for _, dd := range m.devices {
		if f(dd) {
			if err := m.updateDeviceValues(&dd); err != nil {
				return err
			}
			l = append(l, dd)
		}
	}
	*d = l
	return nil
}

func (m *MemDB) UpdateDevice(d models.Device) error {
	for i, dd := range m.devices {
		if dd.Id == d.Id {
			m.devices[i] = d
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetDeviceById(d *models.Device, id string) error {
	return m.getDeviceBy(d,
		func(dd models.Device) bool {
			return dd.Id.Hex() == id
		})
}

func (m *MemDB) GetDeviceByName(d *models.Device, n string) error {
	return m.getDeviceBy(d,
		func(dd models.Device) bool {
			return dd.Name == n
		})
}

func (m *MemDB) GetAllDevices(d *[]models.Device) error {
	cpy := make([]models.Device, len(m.devices))
	copy(cpy, m.devices)
	*d = cpy
	return nil
}

func (m *MemDB) GetDevicesByProfileId(d *[]models.Device, id string) error {
	return m.getDevicesBy(d,
		func(dd models.Device) bool {
			return dd.Profile.Id.Hex() == id
		})
}

func (m *MemDB) GetDevicesByServiceId(d *[]models.Device, id string) error {
	return m.getDevicesBy(d,
		func(dd models.Device) bool {
			return dd.Service.Id.Hex() == id
		})
}

func (m *MemDB) GetDevicesByAddressableId(d *[]models.Device, id string) error {
	return m.getDevicesBy(d,
		func(dd models.Device) bool {
			return dd.Addressable.Id.Hex() == id
		})
}

func (m *MemDB) GetDevicesWithLabel(d *[]models.Device, l string) error {
	return m.getDevicesBy(d,
		func(dd models.Device) bool {
			return stringInSlice(l, dd.Labels)
		})
}

func (m *MemDB) AddDevice(d *models.Device) error {
	currentTime := db.MakeTimestamp()
	d.Created = currentTime
	d.Modified = currentTime
	d.Id = bson.NewObjectId()

	for _, dd := range m.devices {
		if dd.Name == d.Name {
			return db.ErrNotUnique
		}
	}

	validAddressable := false
	// Test addressable id or name exists
	for _, a := range m.addressables {
		if a.Name == d.Addressable.Name {
			validAddressable = true
			break
		}
		if a.Id == d.Addressable.Id {
			validAddressable = true
			break
		}
	}

	if !validAddressable {
		return errors.New("Invalid addressable")
	}

	m.devices = append(m.devices, *d)
	return nil
}

func (m *MemDB) DeleteDeviceById(id string) error {
	for i, dd := range m.devices {
		if dd.Id.Hex() == id {
			m.devices = append(m.devices[:i], m.devices[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) UpdateDeviceProfile(dp *models.DeviceProfile) error {
	for i, d := range m.deviceProfiles {
		if d.Id == dp.Id {
			m.deviceProfiles[i] = *dp
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) AddDeviceProfile(d *models.DeviceProfile) error {
	currentTime := db.MakeTimestamp()
	d.Created = currentTime
	d.Modified = currentTime
	d.Id = bson.NewObjectId()

	for _, dd := range m.deviceProfiles {
		if dd.Name == d.Name {
			return db.ErrNotUnique
		}
	}

	m.deviceProfiles = append(m.deviceProfiles, *d)
	return nil
}

func (m *MemDB) GetAllDeviceProfiles(d *[]models.DeviceProfile) error {
	cpy := make([]models.DeviceProfile, len(m.deviceProfiles))
	copy(cpy, m.deviceProfiles)
	*d = cpy
	return nil
}

func (m *MemDB) GetDeviceProfileById(d *models.DeviceProfile, id string) error {
	for _, dp := range m.deviceProfiles {
		if dp.Id.Hex() == id {
			*d = dp
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) DeleteDeviceProfileById(id string) error {
	for i, d := range m.deviceProfiles {
		if d.Id.Hex() == id {
			m.deviceProfiles = append(m.deviceProfiles[:i], m.deviceProfiles[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetDeviceProfilesByModel(dps *[]models.DeviceProfile, model string) error {
	l := []models.DeviceProfile{}
	for _, dp := range m.deviceProfiles {
		if dp.Model == model {
			l = append(l, dp)
		}
	}
	*dps = l
	return nil
}

func (m *MemDB) GetDeviceProfilesWithLabel(dps *[]models.DeviceProfile, label string) error {
	l := []models.DeviceProfile{}
	for _, dp := range m.deviceProfiles {
		if stringInSlice(label, dp.Labels) {
			l = append(l, dp)
		}
	}
	*dps = l
	return nil
}

func (m *MemDB) GetDeviceProfilesByManufacturerModel(dps *[]models.DeviceProfile, man string, mod string) error {
	l := []models.DeviceProfile{}
	for _, dp := range m.deviceProfiles {
		if dp.Manufacturer == man && dp.Model == mod {
			l = append(l, dp)
		}
	}
	*dps = l
	return nil
}

func (m *MemDB) GetDeviceProfilesByManufacturer(dps *[]models.DeviceProfile, man string) error {
	l := []models.DeviceProfile{}
	for _, dp := range m.deviceProfiles {
		if dp.Manufacturer == man {
			l = append(l, dp)
		}
	}
	*dps = l
	return nil
}

func (m *MemDB) GetDeviceProfileByName(d *models.DeviceProfile, n string) error {
	for _, dp := range m.deviceProfiles {
		if dp.Name == n {
			*d = dp
			return nil
		}
	}
	return db.ErrNotFound
}

// Addressable
func (m *MemDB) UpdateAddressable(updated *models.Addressable, orig *models.Addressable) error {
	if updated == nil {
		return nil
	}
	if updated.Name != "" {
		orig.Name = updated.Name
	}
	if updated.Protocol != "" {
		orig.Protocol = updated.Protocol
	}
	if updated.Address != "" {
		orig.Address = updated.Address
	}
	if updated.Port != int(0) {
		orig.Port = updated.Port
	}
	if updated.Path != "" {
		orig.Path = updated.Path
	}
	if updated.Publisher != "" {
		orig.Publisher = updated.Publisher
	}
	if updated.User != "" {
		orig.User = updated.User
	}
	if updated.Password != "" {
		orig.Password = updated.Password
	}
	if updated.Topic != "" {
		orig.Topic = updated.Topic
	}

	for i, aa := range m.addressables {
		if aa.Id == orig.Id {
			m.addressables[i] = *orig
			return nil
		}
	}

	return db.ErrNotFound
}

func (m *MemDB) AddAddressable(a *models.Addressable) (bson.ObjectId, error) {
	currentTime := db.MakeTimestamp()
	a.Created = currentTime
	a.Modified = currentTime
	a.Id = bson.NewObjectId()

	for _, aa := range m.addressables {
		if aa.Name == a.Name {
			return a.Id, db.ErrNotUnique
		}
	}

	m.addressables = append(m.addressables, *a)
	return a.Id, nil
}

type addressableCmp func(models.Addressable) bool

func (m *MemDB) getAddressableBy(a *models.Addressable, f addressableCmp) error {
	for _, aa := range m.addressables {
		if f(aa) {
			*a = aa
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) getAddressablesBy(a *[]models.Addressable, f addressableCmp) {
	l := []models.Addressable{}
	for _, aa := range m.addressables {
		if f(aa) {
			l = append(l, aa)
		}
	}
	*a = l
}

func (m *MemDB) GetAddressableById(a *models.Addressable, id string) error {
	return m.getAddressableBy(a,
		func(aa models.Addressable) bool {
			return aa.Id.Hex() == id
		})
}

func (m *MemDB) GetAddressableByName(a *models.Addressable, n string) error {
	return m.getAddressableBy(a,
		func(aa models.Addressable) bool {
			return aa.Name == n
		})
}

func (m *MemDB) GetAddressablesByTopic(a *[]models.Addressable, t string) error {
	m.getAddressablesBy(a,
		func(aa models.Addressable) bool {
			return aa.Topic == t
		})
	return nil
}

func (m *MemDB) GetAddressablesByPort(a *[]models.Addressable, p int) error {
	m.getAddressablesBy(a,
		func(aa models.Addressable) bool {
			return aa.Port == p
		})
	return nil
}

func (m *MemDB) GetAddressablesByPublisher(a *[]models.Addressable, p string) error {
	m.getAddressablesBy(a,
		func(aa models.Addressable) bool {
			return aa.Publisher == p
		})
	return nil
}

func (m *MemDB) GetAddressablesByAddress(a *[]models.Addressable, add string) error {
	m.getAddressablesBy(a,
		func(aa models.Addressable) bool {
			return aa.Address == add
		})
	return nil
}

func (m *MemDB) GetAddressables(d *[]models.Addressable) error {
	cpy := make([]models.Addressable, len(m.addressables))
	copy(cpy, m.addressables)
	*d = cpy
	return nil
}

func (m *MemDB) DeleteAddressableById(id string) error {
	for i, aa := range m.addressables {
		if aa.Id.Hex() == id {
			m.addressables = append(m.addressables[:i], m.addressables[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

// Device service
func (m *MemDB) UpdateDeviceService(ds models.DeviceService) error {
	for i, d := range m.deviceServices {
		if d.Id == ds.Id {
			m.deviceServices[i] = ds
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetDeviceServicesByAddressableId(d *[]models.DeviceService, id string) error {
	l := []models.DeviceService{}
	for _, ds := range m.deviceServices {
		if ds.Addressable.Id.Hex() == id {
			err := m.GetAddressableById(&ds.Addressable, ds.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for ds %s",
					ds.Addressable.Id.Hex(), ds.Id.Hex())
			}
			l = append(l, ds)
		}
	}
	*d = l
	return nil
}

func (m *MemDB) GetDeviceServicesWithLabel(d *[]models.DeviceService, label string) error {
	l := []models.DeviceService{}
	for _, ds := range m.deviceServices {
		if stringInSlice(label, ds.Labels) {
			err := m.GetAddressableById(&ds.Addressable, ds.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for ds %s",
					ds.Addressable.Id.Hex(), ds.Id.Hex())
			}
			l = append(l, ds)
		}
	}
	*d = l
	return nil
}

func (m *MemDB) GetDeviceServiceById(d *models.DeviceService, id string) error {
	for _, ds := range m.deviceServices {
		if ds.Id.Hex() == id {
			err := m.GetAddressableById(&ds.Addressable, ds.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for ds %s",
					ds.Addressable.Id.Hex(), ds.Id.Hex())
			}
			*d = ds
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetDeviceServiceByName(d *models.DeviceService, n string) error {
	for _, ds := range m.deviceServices {
		if ds.Name == n {
			err := m.GetAddressableById(&ds.Addressable, ds.Addressable.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find addressable %s for ds %s",
					ds.Addressable.Id.Hex(), ds.Id.Hex())
			}
			*d = ds
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetAllDeviceServices(d *[]models.DeviceService) error {
	for _, ds := range m.deviceServices {
		err := m.GetAddressableById(&ds.Addressable, ds.Addressable.Id.Hex())
		if err != nil {
			return fmt.Errorf("Could not find addressable %s for ds %s",
				ds.Addressable.Id.Hex(), ds.Id.Hex())
		}
	}
	cpy := make([]models.DeviceService, len(m.deviceServices))
	copy(cpy, m.deviceServices)
	*d = cpy
	return nil
}

func (m *MemDB) AddDeviceService(ds *models.DeviceService) error {
	currentTime := db.MakeTimestamp()
	ds.Created = currentTime
	ds.Modified = currentTime
	ds.Id = bson.NewObjectId()

	for _, d := range m.deviceServices {
		if d.Name == ds.Name {
			return db.ErrNotUnique
		}
	}

	validAddressable := false
	// Test addressable id or name exists
	for _, a := range m.addressables {
		if a.Name == ds.Addressable.Name {
			validAddressable = true
			break
		}
		if a.Id == ds.Addressable.Id {
			validAddressable = true
			break
		}
	}

	if !validAddressable {
		return errors.New("Invalid addressable")
	}

	m.deviceServices = append(m.deviceServices, *ds)
	return nil
}

func (m *MemDB) DeleteDeviceServiceById(id string) error {
	for i, d := range m.deviceServices {
		if d.Id.Hex() == id {
			m.deviceServices = append(m.deviceServices[:i], m.deviceServices[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

// Provision watcher
type provisionWatcherComp func(models.ProvisionWatcher) bool

func (m *MemDB) getProvisionWatcherBy(pw *models.ProvisionWatcher, f provisionWatcherComp) error {
	for _, p := range m.provisionWatchers {
		if f(p) {
			err := m.GetDeviceServiceById(&p.Service, p.Service.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find DeviceService %s for ds %s",
					p.Service.Id.Hex(), p.Id.Hex())
			}
			err = m.GetDeviceProfileById(&p.Profile, p.Profile.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find DeviceProfile %s for ds %s",
					p.Profile.Id.Hex(), p.Id.Hex())
			}
			*pw = p
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) getProvisionWatchersBy(pws *[]models.ProvisionWatcher, f provisionWatcherComp) error {
	l := []models.ProvisionWatcher{}
	for _, pw := range m.provisionWatchers {
		if f(pw) {
			err := m.GetDeviceServiceById(&pw.Service, pw.Service.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find DeviceService %s for ds %s",
					pw.Service.Id.Hex(), pw.Id.Hex())
			}
			err = m.GetDeviceProfileById(&pw.Profile, pw.Profile.Id.Hex())
			if err != nil {
				return fmt.Errorf("Could not find DeviceProfile %s for ds %s",
					pw.Profile.Id.Hex(), pw.Id.Hex())
			}
			l = append(l, pw)
		}
	}
	*pws = l
	return nil
}

func (m *MemDB) GetProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	return m.getProvisionWatcherBy(pw,
		func(p models.ProvisionWatcher) bool {
			return p.Id.Hex() == id
		})
}

func (m *MemDB) GetAllProvisionWatchers(pw *[]models.ProvisionWatcher) error {
	*pw = m.provisionWatchers
	return nil
}

func (m *MemDB) GetProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	return m.getProvisionWatcherBy(pw,
		func(p models.ProvisionWatcher) bool {
			return p.Name == n
		})
}

func (m *MemDB) GetProvisionWatchersByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	return m.getProvisionWatchersBy(pw,
		func(p models.ProvisionWatcher) bool {
			return p.Profile.Id.Hex() == id
		})
}

func (m *MemDB) GetProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	return m.getProvisionWatchersBy(pw,
		func(p models.ProvisionWatcher) bool {
			return p.Service.Id.Hex() == id
		})
}

func (m *MemDB) GetProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	return m.getProvisionWatchersBy(pw,
		func(p models.ProvisionWatcher) bool {
			return p.Identifiers[k] == v
		})
}

func (m *MemDB) updateProvisionWatcherValues(pw *models.ProvisionWatcher) error {
	// get Device Service
	validDeviceService := false
	var dev models.DeviceService
	var err error
	if pw.Service.Id.Hex() != "" {
		if err = m.GetDeviceServiceById(&dev, pw.Service.Id.Hex()); err == nil {
			validDeviceService = true
		}
	} else if pw.Service.Name != "" {
		if err = m.GetDeviceServiceByName(&dev, pw.Service.Name); err == nil {
			validDeviceService = true
		}
	} else {
		return errors.New("Device Service ID or Name is required")
	}
	if !validDeviceService {
		return fmt.Errorf("Invalid DeviceService: %v", err)
	}
	pw.Service = dev

	// get Device Profile
	validDeviceProfile := false
	var dp models.DeviceProfile
	if pw.Profile.Id.Hex() != "" {
		if err = m.GetDeviceProfileById(&dp, pw.Profile.Id.Hex()); err == nil {
			validDeviceProfile = true
		}
	} else if pw.Profile.Name != "" {
		if err = m.GetDeviceProfileByName(&dp, pw.Profile.Name); err == nil {
			validDeviceProfile = true
		}
	} else {
		return errors.New("Device Profile ID or Name is required")
	}
	if !validDeviceProfile {
		return fmt.Errorf("Invalid DeviceProfile: %v", err)
	}
	pw.Profile = dp
	return nil
}

func (m *MemDB) AddProvisionWatcher(pw *models.ProvisionWatcher) error {
	currentTime := db.MakeTimestamp()
	pw.Created = currentTime
	pw.Modified = currentTime
	pw.Id = bson.NewObjectId()

	p := models.ProvisionWatcher{}
	if err := m.GetProvisionWatcherByName(&p, pw.Name); err == nil {
		return db.ErrNotUnique
	}

	if err := m.updateProvisionWatcherValues(pw); err != nil {
		return err
	}
	m.provisionWatchers = append(m.provisionWatchers, *pw)
	return nil
}

func (m *MemDB) UpdateProvisionWatcher(pw models.ProvisionWatcher) error {
	pw.Modified = db.MakeTimestamp()

	if err := m.updateProvisionWatcherValues(&pw); err != nil {
		return err
	}
	for i, p := range m.provisionWatchers {
		if pw.Id == p.Id {
			m.provisionWatchers[i] = p
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) DeleteProvisionWatcherById(id string) error {
	for i, p := range m.provisionWatchers {
		if p.Id.Hex() == id {
			m.provisionWatchers = append(m.provisionWatchers[:i], m.provisionWatchers[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

// Command
func (m *MemDB) GetCommandById(c *models.Command, id string) error {
	for _, cc := range m.commands {
		if cc.Id.Hex() == id {
			*c = cc
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetCommandByName(d *[]models.Command, name string) error {
	cmds := []models.Command{}
	for _, cc := range m.commands {
		if cc.Name == name {
			cmds = append(cmds, cc)
		}
	}
	*d = cmds
	return nil
}

func (m *MemDB) AddCommand(c *models.Command) error {
	currentTime := db.MakeTimestamp()
	c.Created = currentTime
	c.Modified = currentTime
	c.Id = bson.NewObjectId()

	m.commands = append(m.commands, *c)
	return nil
}

func (m *MemDB) GetAllCommands(d *[]models.Command) error {
	cpy := make([]models.Command, len(m.commands))
	copy(cpy, m.commands)
	*d = cpy
	return nil
}

func (m *MemDB) UpdateCommand(updated *models.Command, orig *models.Command) error {
	if updated == nil {
		return nil
	}
	if updated.Name != "" {
		orig.Name = updated.Name
	}
	if updated.Get != nil && (updated.Get.String() != models.Get{}.String()) {
		orig.Get = updated.Get
	}
	if updated.Put != nil && (updated.Put.String() != models.Put{}.String()) {
		orig.Put = updated.Put
	}
	if updated.Origin != 0 {
		orig.Origin = updated.Origin
	}

	for i, c := range m.commands {
		if c.Id == orig.Id {
			m.commands[i] = *orig
			return nil
		}
	}

	return db.ErrNotFound
}

func (m *MemDB) DeleteCommandById(id string) error {
	for i, c := range m.commands {
		if c.Id.Hex() == id {
			m.commands = append(m.commands[:i], m.commands[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) GetDeviceProfilesUsingCommand(dps *[]models.DeviceProfile, c models.Command) error {
	l := []models.DeviceProfile{}
	for _, dp := range m.deviceProfiles {
		for _, cc := range dp.Commands {
			if cc.Id == c.Id {
				l = append(l, dp)
				break
			}
		}
	}
	*dps = l
	return nil
}

func (m *MemDB) ScrubMetadata() error {
	m.addressables = nil
	m.commands = nil
	m.deviceServices = nil
	m.schedules = nil
	m.scheduleEvents = nil
	m.provisionWatchers = nil
	m.deviceReports = nil
	m.deviceProfiles = nil
	m.devices = nil
	return nil
}
