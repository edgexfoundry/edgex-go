/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
 *
 * @author: Trevor Conn, Dell
 * @version: 0.5.0
 *******************************************************************************/
package clients

import (
	"time"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2/bson"
)

type MockParams struct {
	EventId bson.ObjectId
	Device string
	Origin int64
}

var mockParams *MockParams

func NewMockParams() *MockParams {
	if mockParams == nil {
		mockParams = &MockParams{
			EventId:bson.NewObjectId(),
			Device:"test device",
			Origin:123456789}
	}
	return mockParams
}

type MockDb struct {

}

func (mc *MockDb) AddReading(r models.Reading) (bson.ObjectId, error) {
	return bson.NewObjectId(), nil
}

//DatabaseClient interface methods
func (mc *MockDb) Events() ([]models.Event, error) {
	ticks := time.Now().Unix()
	events := []models.Event{}

	evt1 := models.Event{ID:mockParams.EventId, Pushed:1, Device:mockParams.Device, Created:ticks, Modified:ticks,
		Origin:mockParams.Origin, Schedule:"TestScheduleA", Event:"SampleEvent", Readings:[]models.Reading{}}

	events = append(events, evt1)
	return events, nil
}

func (mc *MockDb) AddEvent(e *models.Event) (bson.ObjectId, error){
	return bson.NewObjectId(), nil
}

func (mc *MockDb) UpdateEvent(e models.Event) error {
	return nil
}

func (mc *MockDb) EventById(id string) (models.Event, error){
	ticks := time.Now().Unix()

	if id == mockParams.EventId.Hex() {
		return models.Event{ID:mockParams.EventId, Pushed:1, Device:mockParams.Device, Created:ticks, Modified:ticks,
			Origin:mockParams.Origin, Schedule:"TestScheduleA", Event:"SampleEvent", Readings:[]models.Reading{}}, nil
	}
	return models.Event{}, nil
}

func (mc *MockDb) EventCount() (int, error) {
	return 0, nil
}

func (mc *MockDb) EventCountByDeviceId(id string) (int, error) {
	return 0, nil
}

func (mc *MockDb) DeleteEventById(id string) error {
	return nil
}

func (mc *MockDb) EventsForDeviceLimit(id string, limit int) ([]models.Event, error) {
	return []models.Event{}, nil
}

func (mc *MockDb) EventsForDevice(id string) ([]models.Event, error){
	return []models.Event{}, nil
}

func (mc *MockDb) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
	return []models.Event{}, nil
}

func (mc *MockDb) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
	return []models.Reading{}, nil
}

func (mc *MockDb) EventsOlderThanAge(age int64) ([]models.Event, error) {
	return []models.Event{}, nil
}

func (mc *MockDb) EventsPushed() ([]models.Event, error) {
	return []models.Event{}, nil
}

func (mc *MockDb) ScrubAllEvents() error {
	return nil
}

func (mc *MockDb) Readings() ([]models.Reading, error) {
	return []models.Reading{}, nil
}

func (mc *MockDb) UpdateReading(r models.Reading) error {
	return nil
}

func (mc *MockDb) ReadingById(id string) (models.Reading, error) {
	return models.Reading{}, nil
}

func (mc *MockDb) ReadingCount() (int, error) {
	return 0, nil
}

func (mc *MockDb) DeleteReadingById(id string) error {
	return nil
}

func (mc *MockDb) ReadingsByDevice(id string, limit int) ([]models.Reading, error) {
	return []models.Reading{}, nil
}

func (mc *MockDb) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
	return []models.Reading{}, nil
}

func (mc *MockDb) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
	return []models.Reading{}, nil
}

func (mc *MockDb) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
	return []models.Reading{}, nil
}

func (mc *MockDb) AddValueDescriptor(v models.ValueDescriptor) (bson.ObjectId, error) {
	return bson.NewObjectId(), nil
}

func (mc *MockDb) ValueDescriptors() ([]models.ValueDescriptor, error) {
	return []models.ValueDescriptor{}, nil
}

func (mc *MockDb) UpdateValueDescriptor(v models.ValueDescriptor) error {
	return nil
}

func (mc *MockDb) DeleteValueDescriptorById(id string) error {
	return nil
}

func (mc *MockDb) ValueDescriptorByName(name string) (models.ValueDescriptor, error) {
	return models.ValueDescriptor{}, nil
}

func (mc *MockDb) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	return []models.ValueDescriptor{}, nil
}

func (mc *MockDb) ValueDescriptorById(id string) (models.ValueDescriptor, error) {
	return models.ValueDescriptor{}, nil
}


func (mc *MockDb) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	return []models.ValueDescriptor{}, nil
}

func (mc *MockDb) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	return []models.ValueDescriptor{}, nil
}

func (mc *MockDb) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	return []models.ValueDescriptor{}, nil
}
