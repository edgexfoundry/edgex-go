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
	"context"
	"encoding/json"
	"net/url"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// ScheduleClient is an interface used to operate on Core Metadata schedule objects.
type ScheduleClient interface {
	Add(dev *models.Schedule, ctx context.Context) (string, error)
	Delete(id string, ctx context.Context) error
	DeleteByName(name string, ctx context.Context) error
	Schedule(id string, ctx context.Context) (models.Schedule, error)
	ScheduleForName(name string, ctx context.Context) (models.Schedule, error)
	Schedules(ctx context.Context) ([]models.Schedule, error)
	Update(dev models.Schedule, ctx context.Context) error
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

// Helper method to request and decode a schedule
func (s *ScheduleRestClient) requestSchedule(url string, ctx context.Context) (models.Schedule, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.Schedule{}, err
	}

	sched := models.Schedule{}
	err = json.Unmarshal(data, &sched)
	return sched, err
}

// Helper method to request and decode a schedule slice
func (s *ScheduleRestClient) requestScheduleSlice(url string, ctx context.Context) ([]models.Schedule, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return []models.Schedule{}, err
	}

	sSlice := make([]models.Schedule, 0)
	err = json.Unmarshal(data, &sSlice)
	return sSlice, err
}

// Add a schedule.
func (s *ScheduleRestClient) Add(sched *models.Schedule, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(s.url, sched, ctx)
}

// Delete a schedule (specified by id).
func (s *ScheduleRestClient) Delete(id string, ctx context.Context) error {
	return clients.DeleteRequest(s.url+"/id/"+id, ctx)
}

// Delete a schedule (specified by name).
func (s *ScheduleRestClient) DeleteByName(name string, ctx context.Context) error {
	return clients.DeleteRequest(s.url+"/name/"+url.QueryEscape(name), ctx)
}

// Schedule returns the Schedule specified by id.
func (s *ScheduleRestClient) Schedule(id string, ctx context.Context) (models.Schedule, error) {
	return s.requestSchedule(s.url+"/"+id, ctx)
}

// ScheduleForName returns the Schedule specified by name.
func (s *ScheduleRestClient) ScheduleForName(name string, ctx context.Context) (models.Schedule, error) {
	return s.requestSchedule(s.url+"/name/"+url.QueryEscape(name), ctx)
}

// Schedules returns the list of all schedules.
func (s *ScheduleRestClient) Schedules(ctx context.Context) ([]models.Schedule, error) {
	return s.requestScheduleSlice(s.url, ctx)
}

// Update a schedule.
func (s *ScheduleRestClient) Update(sched models.Schedule, ctx context.Context) error {
	return clients.UpdateRequest(s.url, sched, ctx)
}
