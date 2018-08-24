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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

func (m *MemDB) AddReading(r models.Reading) (bson.ObjectId, error) {
	currentTime := db.MakeTimestamp()
	r.Created = currentTime
	r.Modified = currentTime
	r.Id = bson.NewObjectId()

	m.readings = append(m.readings, r)

	return r.Id, nil
}

func (m *MemDB) Events() ([]models.Event, error) {
	cpy := make([]models.Event, len(m.events))
	copy(cpy, m.events)
	return cpy, nil
}

func (m *MemDB) EventsWithLimit(limit int) ([]models.Event, error) {
	if limit > len(m.events) {
		limit = len(m.events)
	}

	cpy := []models.Event{}
	for i := range m.events {
		if i >= limit {
			break
		}
		cpy = append(cpy, m.events[i])
	}
	return cpy, nil
}

func (m *MemDB) AddEvent(e *models.Event) (bson.ObjectId, error) {
	currentTime := db.MakeTimestamp()

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

func (m *MemDB) UpdateEvent(event models.Event) error {
	for i, e := range m.events {
		if e.ID == event.ID {
			m.events[i] = event
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) EventById(id string) (models.Event, error) {
	for _, e := range m.events {
		if e.ID.Hex() == id {
			return e, nil
		}
	}
	return models.Event{}, db.ErrNotFound
}

func (m *MemDB) EventCount() (int, error) {
	return len(m.events), nil
}

func (m *MemDB) EventCountByDeviceId(id string) (int, error) {
	count := 0
	for _, e := range m.events {
		if e.Device == id {
			count += 1
		}
	}
	return count, nil
}

func (m *MemDB) DeleteEventById(id string) error {
	for i, e := range m.events {
		if e.ID.Hex() == id {
			m.events = append(m.events[:i], m.events[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) EventsForDeviceLimit(id string, limit int) ([]models.Event, error) {
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

func (m *MemDB) EventsForDevice(id string) ([]models.Event, error) {
	events := []models.Event{}
	for _, e := range m.events {
		if e.Device == id {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *MemDB) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
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

func (m *MemDB) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
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

func (m *MemDB) EventsOlderThanAge(age int64) ([]models.Event, error) {
	currentTime := db.MakeTimestamp()
	events := []models.Event{}
	for _, e := range m.events {
		if currentTime-e.Created >= age {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *MemDB) EventsPushed() ([]models.Event, error) {
	events := []models.Event{}
	for _, e := range m.events {
		if e.Pushed != 0 {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *MemDB) ScrubAllEvents() error {
	m.events = nil
	m.readings = nil
	return nil
}

func (m *MemDB) Readings() ([]models.Reading, error) {
	cpy := make([]models.Reading, len(m.readings))
	copy(cpy, m.readings)
	return cpy, nil
}

func (m *MemDB) UpdateReading(reading models.Reading) error {
	for i, r := range m.readings {
		if r.Id == reading.Id {
			m.readings[i] = reading
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) ReadingById(id string) (models.Reading, error) {
	for _, r := range m.readings {
		if r.Id.Hex() == id {
			return r, nil
		}
	}
	return models.Reading{}, db.ErrNotFound
}

func (m *MemDB) ReadingCount() (int, error) {
	return len(m.readings), nil
}

func (m *MemDB) DeleteReadingById(id string) error {
	for i, r := range m.readings {
		if r.Id.Hex() == id {
			m.readings = append(m.readings[:i], m.readings[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) ReadingsByDevice(id string, limit int) ([]models.Reading, error) {
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

func (m *MemDB) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
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

func (m *MemDB) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
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

func (m *MemDB) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
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

func (m *MemDB) AddValueDescriptor(value models.ValueDescriptor) (bson.ObjectId, error) {
	currentTime := db.MakeTimestamp()
	value.Created = currentTime
	value.Modified = currentTime
	value.Id = bson.NewObjectId()

	for _, v := range m.vDescriptors {
		if v.Name == value.Name {
			return v.Id, db.ErrNotUnique
		}
	}

	m.vDescriptors = append(m.vDescriptors, value)

	return value.Id, nil
}

func (m *MemDB) ValueDescriptors() ([]models.ValueDescriptor, error) {
	cpy := make([]models.ValueDescriptor, len(m.vDescriptors))
	copy(cpy, m.vDescriptors)
	return cpy, nil
}

func (m *MemDB) UpdateValueDescriptor(value models.ValueDescriptor) error {
	for i, v := range m.vDescriptors {
		if v.Id == value.Id {
			m.vDescriptors[i] = value
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) DeleteValueDescriptorById(id string) error {
	for i, v := range m.vDescriptors {
		if v.Id.Hex() == id {
			m.vDescriptors = append(m.vDescriptors[:i], m.vDescriptors[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MemDB) ValueDescriptorByName(name string) (models.ValueDescriptor, error) {
	for _, v := range m.vDescriptors {
		if v.Name == name {
			return v, nil
		}
	}
	return models.ValueDescriptor{}, db.ErrNotFound
}

func (m *MemDB) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if stringInSlice(v.Name, names) {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *MemDB) ValueDescriptorById(id string) (models.ValueDescriptor, error) {
	for _, v := range m.vDescriptors {
		if v.Id.Hex() == id {
			return v, nil
		}
	}
	return models.ValueDescriptor{}, db.ErrNotFound
}

func (m *MemDB) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if v.UomLabel == uomLabel {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *MemDB) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if stringInSlice(label, v.Labels) {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *MemDB) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	for _, v := range m.vDescriptors {
		if v.Type == t {
			vds = append(vds, v)
		}
	}
	return vds, nil
}

func (m *MemDB) ScrubAllValueDescriptors() error {
	m.vDescriptors = nil
	return nil
}
