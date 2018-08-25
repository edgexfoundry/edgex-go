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

// ScheduleClient is an interface used to operate on Core Metadata schedule objects.
type ScheduleClient interface {
	Add(dev *models.Schedule) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
	Schedule(id string) (models.Schedule, error)
	ScheduleForName(name string) (models.Schedule, error)
	Schedules() ([]models.Schedule, error)
	Update(dev models.Schedule) error
}

// ScheduleRestClient is struct used as a receiver for ScheduleClient interface methods.
type ScheduleRestClient struct {
	url      string
	endpoint clients.Endpointer
}

// NewScheduleClient returns a new instance of ScheduleClient.
func NewScheduleClient(params types.EndpointParams, m clients.Endpointer) ScheduleClient {
	s := ScheduleRestClient{endpoint: m}
	s.init(params)
	return &s
}

func (s *ScheduleRestClient) init(params types.EndpointParams) {
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

// Help method to decode a schedule slice
func (s *ScheduleRestClient) decodeScheduleSlice(resp *http.Response) ([]models.Schedule, error) {
	dec := json.NewDecoder(resp.Body)
	sSlice := []models.Schedule{}

	err := dec.Decode(&sSlice)
	if err != nil {
		return []models.Schedule{}, err
	}

	return sSlice, err
}

// Helper method to decode a schedule and return the schedule
func (s *ScheduleRestClient) decodeSchedule(resp *http.Response) (models.Schedule, error) {
	dec := json.NewDecoder(resp.Body)
	sched := models.Schedule{}

	err := dec.Decode(&sched)
	if err != nil {
		return models.Schedule{}, err
	}

	return sched, err
}

// Add a schedule.
func (s *ScheduleRestClient) Add(dev *models.Schedule) (string, error) {
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

// Delete a schedule (specified by id).
func (s *ScheduleRestClient) Delete(id string) error {
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

// Delete a schedule (specified by name).
func (s *ScheduleRestClient) DeleteByName(name string) error {
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

// Schedule returns the Schedule specified by id.
func (s *ScheduleRestClient) Schedule(id string) (models.Schedule, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/"+id, nil)
	if err != nil {
		return models.Schedule{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.Schedule{}, err
	}
	if resp == nil {
		return models.Schedule{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.Schedule{}, err
		}

		return models.Schedule{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return s.decodeSchedule(resp)
}

// ScheduleForName returns the Schedule specified by name.
func (s *ScheduleRestClient) ScheduleForName(name string) (models.Schedule, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		return models.Schedule{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.Schedule{}, err
	}
	if resp == nil {
		return models.Schedule{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.Schedule{}, err
		}

		return models.Schedule{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return s.decodeSchedule(resp)
}

// Schedules returns the list of all schedules.
func (s *ScheduleRestClient) Schedules() ([]models.Schedule, error) {
	req, err := http.NewRequest(http.MethodGet, s.url, nil)
	if err != nil {
		return []models.Schedule{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Schedule{}, err
	}
	if resp == nil {
		return []models.Schedule{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Schedule{}, err
		}

		return []models.Schedule{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return s.decodeScheduleSlice(resp)
}

// Update a schedule.
func (s *ScheduleRestClient) Update(dev models.Schedule) error {
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
