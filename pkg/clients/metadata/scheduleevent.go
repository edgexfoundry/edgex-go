/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (C) 2018 Canonical Ltd
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
package metadata

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// ScheduleEventClient is an interface used to operate on Core Metadata schedule event objects.
type ScheduleEventClient interface {
	Add(dev *models.ScheduleEvent) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
	ScheduleEvent(id string) (models.ScheduleEvent, error)
	ScheduleEventForName(name string) (models.ScheduleEvent, error)
	ScheduleEvents() ([]models.ScheduleEvent, error)
	ScheduleEventsForAddressable(name string) ([]models.ScheduleEvent, error)
	ScheduleEventsForAddressableByName(name string) ([]models.ScheduleEvent, error)
	ScheduleEventsForServiceByName(name string) ([]models.ScheduleEvent, error)
	Update(dev models.ScheduleEvent) error
}

// ScheduleEventRestClient is struct used as a receiver for ScheduleEventClient interface methods.
type ScheduleEventRestClient struct {
	url      string
	endpoint clients.Endpointer
}

// NewScheduleEventClient returns a new instance of ScheduleEventClient.
func NewScheduleEventClient(params types.EndpointParams, m clients.Endpointer) ScheduleEventClient {
	s := ScheduleEventRestClient{endpoint: m}
	s.init(params)
	return &s
}

func (s *ScheduleEventRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go s.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					s.url = url
				}
			}
		}(ch)
	} else {
		s.url = params.Url
	}
}

// Help method to decode a schedule event slice
func (s *ScheduleEventRestClient) decodeScheduleEventSlice(resp *http.Response) ([]models.ScheduleEvent, error) {
	dec := json.NewDecoder(resp.Body)
	seSlice := []models.ScheduleEvent{}

	err := dec.Decode(&seSlice)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}

	return seSlice, err
}

// Helper method to decode a schedule and return the schedule event
func (s *ScheduleEventRestClient) decodeScheduleEvent(resp *http.Response) (models.ScheduleEvent, error) {
	dec := json.NewDecoder(resp.Body)
	event := models.ScheduleEvent{}

	err := dec.Decode(&event)
	if err != nil {
		return models.ScheduleEvent{}, err
	}

	return event, err
}

// Add a schedule event.
func (s *ScheduleEventRestClient) Add(dev *models.ScheduleEvent) (string, error) {
	jsonStr, err := json.Marshal(dev)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(jsonStr))
	if err != nil {
		return "", err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the body
	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return string(bodyBytes), nil
}

// Delete a schedule event (specified by id).
func (s *ScheduleEventRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, s.url+"/id/"+id, nil)
	if err != nil {
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}

// Delete a schedule event (specified by name).
func (s *ScheduleEventRestClient) DeleteByName(name string) error {
	req, err := http.NewRequest(http.MethodDelete, s.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}

// ScheduleEvent returns the ScheduleEvent specified by id.
func (s *ScheduleEventRestClient) ScheduleEvent(id string) (models.ScheduleEvent, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/"+id, nil)
	if err != nil {
		return models.ScheduleEvent{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.ScheduleEvent{}, err
	}
	if resp == nil {
		return models.ScheduleEvent{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.ScheduleEvent{}, err
		}

		return models.ScheduleEvent{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return s.decodeScheduleEvent(resp)
}

// ScheduleEventForName returns the ScheduleEvent specified by name.
func (s *ScheduleEventRestClient) ScheduleEventForName(name string) (models.ScheduleEvent, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		return models.ScheduleEvent{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.ScheduleEvent{}, err
	}
	if resp == nil {
		return models.ScheduleEvent{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.ScheduleEvent{}, err
		}

		return models.ScheduleEvent{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return s.decodeScheduleEvent(resp)
}

// Get a list of all schedules events.
func (s *ScheduleEventRestClient) ScheduleEvents() ([]models.ScheduleEvent, error) {
	req, err := http.NewRequest(http.MethodGet, s.url, nil)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}
	if resp == nil {
		return []models.ScheduleEvent{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ScheduleEvent{}, err
		}

		return []models.ScheduleEvent{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return s.decodeScheduleEventSlice(resp)
}

// ScheduleEventForAddressable returns the ScheduleEvent specified by addressable.
func (s *ScheduleEventRestClient) ScheduleEventsForAddressable(addressable string) ([]models.ScheduleEvent, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/addressable/"+url.QueryEscape(addressable), nil)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}
	if resp == nil {
		return []models.ScheduleEvent{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ScheduleEvent{}, err
		}

		return []models.ScheduleEvent{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return s.decodeScheduleEventSlice(resp)
}

// ScheduleEventForAddressableByName returns the ScheduleEvent specified by addressable name.
func (s *ScheduleEventRestClient) ScheduleEventsForAddressableByName(name string) ([]models.ScheduleEvent, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/addressablename/"+url.QueryEscape(name), nil)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}
	if resp == nil {
		return []models.ScheduleEvent{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ScheduleEvent{}, err
		}

		return []models.ScheduleEvent{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return s.decodeScheduleEventSlice(resp)
}

// Get the schedule event for service by name.
func (s *ScheduleEventRestClient) ScheduleEventsForServiceByName(name string) ([]models.ScheduleEvent, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/servicename/"+url.QueryEscape(name), nil)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}
	if resp == nil {
		return []models.ScheduleEvent{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ScheduleEvent{}, err
		}

		return []models.ScheduleEvent{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return s.decodeScheduleEventSlice(resp)
}

// Update a schedule event - handle error codes
func (s *ScheduleEventRestClient) Update(dev models.ScheduleEvent) error {
	jsonStr, err := json.Marshal(&dev)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, s.url, bytes.NewReader(jsonStr))
	if err != nil {
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}
