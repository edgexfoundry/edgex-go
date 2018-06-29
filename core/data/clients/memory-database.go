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
package clients

import (
	"time"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2/bson"
)

type memDB struct {
	readings     []models.Reading
	events       []models.Event
	vDescriptors []models.ValueDescriptor
}

func (m *memDB) CloseSession() {
}

func (m *memDB) AddReading(r models.Reading) (bson.ObjectId, error) {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	r.Created = currentTime
	r.Modified = currentTime
	r.Id = bson.NewObjectId()

	m.readings = append(m.readings, r)

	return r.Id, nil
}

func (m *memDB) Events() ([]models.Event, error) {
	return m.events, nil
}

func (m *memDB) AddEvent(e *models.Event) (bson.ObjectId, error) {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)

	for i := range e.Readings {
		e.Readings[i].Id = bson.NewObjectId()
		e.Readings[i].Created = currentTime
		e.Readings[i].Modified = currentTime
		e.Readings[i].Device = e.Device
		m.readings = append(m.readings, e.Readings[i])
	}

	e.Created = currentTime
	e.Modified = currentTime
	e.ID = bson.NewObjectId()

	m.events = append(m.events, *e)

	return e.ID, nil
}

func (m *memDB) UpdateEvent(event models.Event) error {
	for i, e := range m.events {
		if e.ID == event.ID {
			m.events[i] = event
			return nil
		}
	}
	return ErrNotFound
}

func (m *memDB) EventById(id string) (models.Event, error) {
	for _, e := range m.events {
		if e.ID.Hex() == id {
			return e, nil
		}
	}
	return models.Event{}, ErrNotFound
}

func (m *memDB) EventCount() (int, error) {
	return len(m.events), nil
}

func (m *memDB) EventCountByDeviceId(id string) (int, error) {
	count := 0
	for _, e := range m.events {
		if e.Device == id {
			count += 1
		}
	}
	return count, nil
}

func (m *memDB) DeleteEventById(id string) error {
	for i, e := range m.events {
		if e.ID.Hex() == id {
			m.events = append(m.events[:i], m.events[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *memDB) EventsForDeviceLimit(id string, limit int) ([]models.Event, error) {
	events := []models.Event{}
	count := 0
	for _, e := range m.events {
		if e.Device == id {
			events = append(events, e)
			count += 1
			if count == limit {
				break
			}
		}
	}
	return events, nil
}

func (m *memDB) EventsForDevice(id string) ([]models.Event, error) {
	events := []models.Event{}
	for _, e := range m.events {
		if e.Device == id {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *memDB) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
	events := []models.Event{}
	count := 0
	for _, e := range m.events {
		if e.Created >= startTime && e.Created <= endTime {
			events = append(events, e)
			count += 1
			if count == limit {
				break
			}
		}
	}
	return events, nil
}

func (m *memDB) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
	readings := []models.Reading{}
	count := 0
	for _, r := range m.readings {
		if r.Device == deviceId && r.Name == valueDescriptor {
			readings = append(readings, r)
			count += 1
			if count == limit {
				break
			}
		}
	}
	return readings, nil
}

func (m *memDB) EventsOlderThanAge(age int64) ([]models.Event, error) {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	events := []models.Event{}
	for _, e := range m.events {
		if currentTime-e.Created >= age {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *memDB) EventsPushed() ([]models.Event, error) {
	events := []models.Event{}
	for _, e := range m.events {
		if e.Pushed != 0 {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *memDB) ScrubAllEvents() error {
	m.events = nil
	m.readings = nil
	return nil
}

func (m *memDB) Readings() ([]models.Reading, error) {
	return m.readings, nil
}

func (m *memDB) UpdateReading(reading models.Reading) error {
	for i, r := range m.readings {
		if r.Id == reading.Id {
			m.readings[i] = reading
			return nil
		}
	}
	return ErrNotFound
}

func (m *memDB) ReadingById(id string) (models.Reading, error) {
	for _, r := range m.readings {
		if r.Id.Hex() == id {
			return r, nil
		}
	}
	return models.Reading{}, ErrNotFound
}

func (m *memDB) ReadingCount() (int, error) {
	return len(m.readings), nil
}

func (m *memDB) DeleteReadingById(id string) error {
	for i, r := range m.readings {
		if r.Id.Hex() == id {
			m.readings = append(m.readings[:i], m.readings[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *memDB) ReadingsByDevice(id string, limit int) ([]models.Reading, error) {
	readings := []models.Reading{}
	count := 0
	for _, r := range m.readings {
		if r.Device == id {
			readings = append(readings, r)
			count += 1
			if count == limit {
				break
			}
		}
	}
	return readings, nil
}

func (m *memDB) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
	readings := []models.Reading{}
	count := 0
	for _, r := range m.readings {
		if r.Name == name {
			readings = append(readings, r)
			count += 1
			if count == limit {
				break
			}
		}
	}
	return readings, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (m *memDB) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
	readings := []models.Reading{}
	count := 0
	for _, r := range m.readings {
		if stringInSlice(r.Name, names) {
			readings = append(readings, r)
			count += 1
			if count == limit {
				break
			}
		}
	}
	return readings, nil
}

func (m *memDB) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
	readings := []models.Reading{}
	count := 0
	for _, r := range m.readings {
		if r.Created >= start && r.Created <= end {
			readings = append(readings, r)
			count += 1
			if count == limit {
				break
			}
		}
	}
	return readings, nil
}

func (m *memDB) AddValueDescriptor(value models.ValueDescriptor) (bson.ObjectId, error) {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	value.Created = currentTime
	value.Modified = currentTime
	value.Id = bson.NewObjectId()

	for _, v := range m.vDescriptors {
		if v.Name == value.Name {
			return v.Id, ErrNotUnique
		}
	}

	m.vDescriptors = append(m.vDescriptors, value)

	return value.Id, nil
}

func (m *memDB) ValueDescriptors() ([]models.ValueDescriptor, error) {
	return m.vDescriptors, nil
}

func (m *memDB) UpdateValueDescriptor(value models.ValueDescriptor) error {
	for i, v := range m.vDescriptors {
		if v.Id == value.Id {
			m.vDescriptors[i] = value
			return nil
		}
	}
	return ErrNotFound
}

func (m *memDB) DeleteValueDescriptorById(id string) error {
	for i, v := range m.vDescriptors {
		if v.Id.Hex() == id {
			m.vDescriptors = append(m.vDescriptors[:i], m.vDescriptors[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *memDB) ValueDescriptorByName(name string) (models.ValueDescriptor, error) {
	for _, v := range m.vDescriptors {
		if v.Name == name {
			return v, nil
		}
	}
	return models.ValueDescriptor{}, ErrNotFound
}

func (m *memDB) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if stringInSlice(v.Name, names) {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *memDB) ValueDescriptorById(id string) (models.ValueDescriptor, error) {
	for _, v := range m.vDescriptors {
		if v.Id.Hex() == id {
			return v, nil
		}
	}
	return models.ValueDescriptor{}, ErrNotFound
}

func (m *memDB) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if v.UomLabel == uomLabel {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *memDB) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if stringInSlice(label, v.Labels) {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *memDB) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if v.Type == t {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *memDB) ScrubAllValueDescriptors() error {
	m.vDescriptors = nil
	return nil
}
