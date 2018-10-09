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

type ReadingClient interface {
	Readings() ([]models.Reading, error)
	ReadingCount() (int, error)
	Reading(id string) (models.Reading, error)
	ReadingsForDevice(deviceId string, limit int) ([]models.Reading, error)
	ReadingsForNameAndDevice(name string, deviceId string, limit int) ([]models.Reading, error)
	ReadingsForName(name string, limit int) ([]models.Reading, error)
	ReadingsForUOMLabel(uomLabel string, limit int) ([]models.Reading, error)
	ReadingsForLabel(label string, limit int) ([]models.Reading, error)
	ReadingsForType(readingType string, limit int) ([]models.Reading, error)
	ReadingsForInterval(start int, end int, limit int) ([]models.Reading, error)
	Add(readiing *models.Reading) (string, error)
	Delete(id string) error
}

type ReadingRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewReadingClient(params types.EndpointParams, m clients.Endpointer) ReadingClient {
	r := ReadingRestClient{endpoint: m}
	r.init(params)
	return &r
}

func (r *ReadingRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go r.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					r.url = url
				}
			}
		}(ch)
	} else {
		r.url = params.Url
	}
}

// Helper method to request and decode a reading slice
func (r *ReadingRestClient) requestReadingSlice(url string) ([]models.Reading, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return []models.Reading{}, err
	}

	rSlice := make([]models.Reading, 0)
	err = json.Unmarshal(data, &rSlice)
	return rSlice, err
}

// Helper method to request and decode a reading
func (r *ReadingRestClient) requestReading(url string) (models.Reading, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.Reading{}, err
	}

	reading := models.Reading{}
	err = json.Unmarshal(data, &reading)
	return reading, err
}

// Get a list of all readings
func (r *ReadingRestClient) Readings() ([]models.Reading, error) {
	return r.requestReadingSlice(r.url)
}

// Get the reading by id
func (r *ReadingRestClient) Reading(id string) (models.Reading, error) {
	return r.requestReading(r.url + "/" + id)
}

// Get reading count
func (r *ReadingRestClient) ReadingCount() (int, error) {
	return clients.CountRequest(r.url + "/count")
}

// Get the readings for a device
func (r *ReadingRestClient) ReadingsForDevice(deviceId string, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/device/" + url.QueryEscape(deviceId) + "/" + strconv.Itoa(limit))
}

// Get the readings for name and device
func (r *ReadingRestClient) ReadingsForNameAndDevice(name string, deviceId string, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/name/" + url.QueryEscape(name) + "/device/" + url.QueryEscape(deviceId) + "/" + strconv.Itoa(limit))
}

// Get readings by name
func (r *ReadingRestClient) ReadingsForName(name string, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/name/" + url.QueryEscape(name) + "/" + strconv.Itoa(limit))
}

// Get readings for UOM Label
func (r *ReadingRestClient) ReadingsForUOMLabel(uomLabel string, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/uomlabel/" + url.QueryEscape(uomLabel) + "/" + strconv.Itoa(limit))
}

// Get readings for label
func (r *ReadingRestClient) ReadingsForLabel(label string, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/label/" + url.QueryEscape(label) + "/" + strconv.Itoa(limit))
}

// Get readings for type
func (r *ReadingRestClient) ReadingsForType(readingType string, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/type/" + url.QueryEscape(readingType) + "/" + strconv.Itoa(limit))
}

// Get readings for interval
func (r *ReadingRestClient) ReadingsForInterval(start int, end int, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/" + strconv.Itoa(start) + "/" + strconv.Itoa(end) + "/" + strconv.Itoa(limit))
}

// Get readings for device and value descriptor
func (r *ReadingRestClient) ReadingsForDeviceAndValueDescriptor(deviceId string, vd string, limit int) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url + "/device/" + url.QueryEscape(deviceId) + "/valuedescriptor/" + url.QueryEscape(vd) + "/" + strconv.Itoa(limit))
}

// Add a reading
func (r *ReadingRestClient) Add(reading *models.Reading) (string, error) {
	return clients.PostJsonRequest(r.url, reading)
}

// Delete a reading by id
func (r *ReadingRestClient) Delete(id string) error {
	return clients.DeleteRequest(r.url + "/id/" + id)
}
