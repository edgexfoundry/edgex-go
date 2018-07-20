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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

// Helper method to decode an event slice
func (e *EventRestClient) decodeEventSlice(resp *http.Response) ([]models.Event, error) {
	eSlice := make([]models.Event, 0)
	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&eSlice)
	if err != nil {
		fmt.Println(err)
	}

	return eSlice, err
}

// Helper method to decode an event and return the event
func (e *EventRestClient) decodeEvent(resp *http.Response) (models.Event, error) {
	dec := json.NewDecoder(resp.Body)
	ev := models.Event{}
	err := dec.Decode(&ev)
	if err != nil {
		fmt.Println(err)
	}

	return ev, err
}

// Get a list of all events
func (e *EventRestClient) Events() ([]models.Event, error) {
	req, err := http.NewRequest(http.MethodGet, e.url, nil)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Event{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Event{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Event{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Event{}, err
		}
		bodyString := string(bodyBytes)
		return []models.Event{}, errors.New(string(bodyString))
	}

	return e.decodeEventSlice(resp)
}

// Get the event by id
func (e *EventRestClient) Event(id string) (models.Event, error) {
	req, err := http.NewRequest(http.MethodGet, e.url+"/"+id, nil)
	if err != nil {
		fmt.Println(err)
		return models.Event{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return models.Event{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.Event{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.Event{}, err
		}
		bodyString := string(bodyBytes)

		return models.Event{}, errors.New(bodyString)
	}

	return e.decodeEvent(resp)
}

// Get event count
func (e *EventRestClient) EventCount() (int, error) {
	req, err := http.NewRequest(http.MethodGet, e.url+"/count", nil)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return 0, ErrResponseNil
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(bodyString)
	}
	count, err := strconv.Atoi(bodyString)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Get event count for device
func (e *EventRestClient) EventCountForDevice(deviceId string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, e.url+"/count/"+url.QueryEscape(deviceId), nil)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return 0, ErrResponseNil
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(bodyString)
	}

	count, err := strconv.Atoi(bodyString)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Get events for device
func (e *EventRestClient) EventsForDevice(deviceId string, limit int) ([]models.Event, error) {
	req, err := http.NewRequest(http.MethodGet, e.url+"/device/"+url.QueryEscape(deviceId)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Event{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Event{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Event{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Event{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Event{}, errors.New(bodyString)
	}
	return e.decodeEventSlice(resp)
}

// Get events for interval
func (e *EventRestClient) EventsForInterval(start int, end int, limit int) ([]models.Event, error) {
	req, err := http.NewRequest(http.MethodGet, e.url+"/"+strconv.Itoa(start)+"/"+strconv.Itoa(end)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Event{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Event{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Event{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Event{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Event{}, errors.New(bodyString)
	}
	return e.decodeEventSlice(resp)
}

// Get events for device and value descriptor
func (e *EventRestClient) EventsForDeviceAndValueDescriptor(deviceId string, vd string, limit int) ([]models.Event, error) {
	req, err := http.NewRequest(http.MethodGet, e.url+"/device/"+url.QueryEscape(deviceId)+"/valuedescriptor/"+url.QueryEscape(vd)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Event{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Event{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Event{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Event{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Event{}, errors.New(bodyString)
	}
	return e.decodeEventSlice(resp)
}

// Add event
func (e *EventRestClient) Add(event *models.Event) (string, error) {
	jsonStr, err := json.Marshal(event)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, e.url, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the response body
	bodyBytes, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}

// Delete event by id
func (e *EventRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, e.url+"/id/"+id, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Delete events by device name
func (e *EventRestClient) DeleteForDevice(deviceId string) error {
	req, err := http.NewRequest(http.MethodDelete, e.url+"/device/"+url.QueryEscape(deviceId), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Delete events by age
func (e *EventRestClient) DeleteOld(age int) error {
	req, err := http.NewRequest(http.MethodDelete, e.url+"/removeold/age/"+strconv.Itoa(age), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Mark event as pushed
func (e *EventRestClient) MarkPushed(id string) error {
	req, err := http.NewRequest(http.MethodPut, e.url+"/id/"+id, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if resp == nil {
		fmt.Println(ErrResponseNil)
		return ErrResponseNil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}
