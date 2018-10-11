/*******************************************************************************
 * Copyright 1995-2018 Hitachi Vantara Corporation. All rights reserved.
 *
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
 *******************************************************************************/
package coredata

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type EventClient interface {
	Events() ([]models.Event, error)
	Event(id string) (models.Event, error)
	EventCount() (int, error)
	EventCountForDevice(deviceId string) (int, error)
	EventsForDevice(id string, limit int) ([]models.Event, error)
	EventsForInterval(start int, end int, limit int) ([]models.Event, error)
	EventsForDeviceAndValueDescriptor(deviceId string, vd string, limit int) ([]models.Event, error)
	Add(event *models.Event) (string, error)
	DeleteForDevice(id string) error
	DeleteOld(age int) error
	Delete(id string) error
	MarkPushed(id string) error
}

type EventRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewEventClient(params types.EndpointParams, m clients.Endpointer) EventClient {
	e := EventRestClient{endpoint: m}
	e.init(params)
	return &e
}

func (e *EventRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go e.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					e.url = url
				}
			}
		}(ch)
	} else {
		e.url = params.Url
	}
}

// Helper method to request and decode an event slice
func (e *EventRestClient) requestEventSlice(url string) ([]models.Event, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return []models.Event{}, err
	}

	eSlice := make([]models.Event, 0)
	err = json.Unmarshal(data, &eSlice)
	return eSlice, err
}

// Helper method to request and decode an event
func (e *EventRestClient) requestEvent(url string) (models.Event, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.Event{}, err
	}

	ev := models.Event{}
	err = json.Unmarshal(data, &ev)
	return ev, err
}

// Get a list of all events
func (e *EventRestClient) Events() ([]models.Event, error) {
	return e.requestEventSlice(e.url)
}

// Get the event by id
func (e *EventRestClient) Event(id string) (models.Event, error) {
	return e.requestEvent(e.url + "/" + id)
}

// Get event count
func (e *EventRestClient) EventCount() (int, error) {
	return clients.CountRequest(e.url + "/count")
}

// Get event count for device
func (e *EventRestClient) EventCountForDevice(deviceId string) (int, error) {
	return clients.CountRequest(e.url + "/count/" + url.QueryEscape(deviceId))
}

// Get events for device
func (e *EventRestClient) EventsForDevice(deviceId string, limit int) ([]models.Event, error) {
	return e.requestEventSlice(e.url + "/device/" + url.QueryEscape(deviceId) + "/" + strconv.Itoa(limit))
}

// Get events for interval
func (e *EventRestClient) EventsForInterval(start int, end int, limit int) ([]models.Event, error) {
	return e.requestEventSlice(e.url + "/" + strconv.Itoa(start) + "/" + strconv.Itoa(end) + "/" + strconv.Itoa(limit))
}

// Get events for device and value descriptor
func (e *EventRestClient) EventsForDeviceAndValueDescriptor(deviceId string, vd string, limit int) ([]models.Event, error) {
	return e.requestEventSlice(e.url + "/device/" + url.QueryEscape(deviceId) + "/valuedescriptor/" + url.QueryEscape(vd) + "/" + strconv.Itoa(limit))
}

// Add event
func (e *EventRestClient) Add(event *models.Event) (string, error) {
	return clients.PostJsonRequest(e.url, event)
}

// Delete event by id
func (e *EventRestClient) Delete(id string) error {
	return clients.DeleteRequest(e.url + "/id/" + id)
}

// Delete events by device name
func (e *EventRestClient) DeleteForDevice(deviceId string) error {
	return clients.DeleteRequest(e.url + "/device/" + url.QueryEscape(deviceId))
}

// Delete events by age
func (e *EventRestClient) DeleteOld(age int) error {
	return clients.DeleteRequest(e.url + "/removeold/age/" + strconv.Itoa(age))
}

// Mark event as pushed
func (e *EventRestClient) MarkPushed(id string) error {
	_, err := clients.PutRequest(e.url+"/id/"+id, nil)
	return err
}
