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

/*
Provision Watcher client for interacting with the provision watcher section of metadata
*/
type ProvisionWatcherClient interface {
	Add(dev *models.ProvisionWatcher, ctx context.Context) (string, error)
	Delete(id string, ctx context.Context) error
	ProvisionWatcher(id string, ctx context.Context) (models.ProvisionWatcher, error)
	ProvisionWatcherForName(name string, ctx context.Context) (models.ProvisionWatcher, error)
	ProvisionWatchers(ctx context.Context) ([]models.ProvisionWatcher, error)
	ProvisionWatchersForService(serviceId string, ctx context.Context) ([]models.ProvisionWatcher, error)
	ProvisionWatchersForServiceByName(serviceName string, ctx context.Context) ([]models.ProvisionWatcher, error)
	ProvisionWatchersForProfile(profileid string, ctx context.Context) ([]models.ProvisionWatcher, error)
	ProvisionWatchersForProfileByName(profileName string, ctx context.Context) ([]models.ProvisionWatcher, error)
	Update(dev models.ProvisionWatcher, ctx context.Context) error
}

type ProvisionWatcherRestClient struct {
	url      string
	endpoint clients.Endpointer
}

/*
Return an instance of ProvisionWatcherClient
*/
func NewProvisionWatcherClient(params types.EndpointParams, m clients.Endpointer) ProvisionWatcherClient {
	pw := ProvisionWatcherRestClient{endpoint: m}
	pw.init(params)
	return &pw
}

func (pw *ProvisionWatcherRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go pw.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					pw.url = url
				}
			}
		}(ch)
	} else {
		pw.url = params.Url
	}
}

// Helper method to request and decode a provision watcher
func (pw *ProvisionWatcherRestClient) requestProvisionWatcher(url string, ctx context.Context) (models.ProvisionWatcher, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.ProvisionWatcher{}, err
	}

	watcher := models.ProvisionWatcher{}
	err = json.Unmarshal(data, &watcher)
	return watcher, err
}

// Helper method to request and decode a provision watcher slice
func (pw *ProvisionWatcherRestClient) requestProvisionWatcherSlice(url string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return []models.ProvisionWatcher{}, err
	}

	pwSlice := make([]models.ProvisionWatcher, 0)
	err = json.Unmarshal(data, &pwSlice)
	return pwSlice, err
}

// Get the provision watcher by id
func (pw *ProvisionWatcherRestClient) ProvisionWatcher(id string, ctx context.Context) (models.ProvisionWatcher, error) {
	return pw.requestProvisionWatcher(pw.url+"/"+id, ctx)
}

// Get a list of all provision watchers
func (pw *ProvisionWatcherRestClient) ProvisionWatchers(ctx context.Context) ([]models.ProvisionWatcher, error) {
	return pw.requestProvisionWatcherSlice(pw.url, ctx)
}

// Get the provision watcher by name
func (pw *ProvisionWatcherRestClient) ProvisionWatcherForName(name string, ctx context.Context) (models.ProvisionWatcher, error) {
	return pw.requestProvisionWatcher(pw.url+"/name/"+url.QueryEscape(name), ctx)
}

// Get the provision watchers that are on a service
func (pw *ProvisionWatcherRestClient) ProvisionWatchersForService(serviceId string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	return pw.requestProvisionWatcherSlice(pw.url+"/service/"+serviceId, ctx)
}

// Get the provision watchers that are on a service(by name)
func (pw *ProvisionWatcherRestClient) ProvisionWatchersForServiceByName(serviceName string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	return pw.requestProvisionWatcherSlice(pw.url+"/servicename/"+url.QueryEscape(serviceName), ctx)
}

// Get the provision watchers for a profile
func (pw *ProvisionWatcherRestClient) ProvisionWatchersForProfile(profileId string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	return pw.requestProvisionWatcherSlice(pw.url+"/profile/"+profileId, ctx)
}

// Get the provision watchers for a profile (by name)
func (pw *ProvisionWatcherRestClient) ProvisionWatchersForProfileByName(profileName string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	return pw.requestProvisionWatcherSlice(pw.url+"/profilename/"+url.QueryEscape(profileName), ctx)
}

// Add a provision watcher - handle error codes
func (pw *ProvisionWatcherRestClient) Add(dev *models.ProvisionWatcher, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(pw.url, dev, ctx)
}

// Update a provision watcher - handle error codes
func (pw *ProvisionWatcherRestClient) Update(dev models.ProvisionWatcher, ctx context.Context) error {
	return clients.UpdateRequest(pw.url, dev, ctx)
}

// Delete a provision watcher (specified by id)
func (pw *ProvisionWatcherRestClient) Delete(id string, ctx context.Context) error {
	return clients.DeleteRequest(pw.url+"/id/"+id, ctx)
}
