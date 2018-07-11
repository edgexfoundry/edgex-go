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

// Help method to decode a reading slice
func (r *ReadingRestClient) decodeReadingSlice(resp *http.Response) ([]models.Reading, error) {
	rSlice := make([]models.Reading, 0)

	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&rSlice)
	if err != nil {
		fmt.Println(err)
	}

	return rSlice, err
}

// Helper method to decode a reading and return the reading
func (r *ReadingRestClient) decodeReading(resp *http.Response) (models.Reading, error) {
	dec := json.NewDecoder(resp.Body)
	reading := models.Reading{}

	err := dec.Decode(&reading)
	if err != nil {
		fmt.Println(err)
	}

	return reading, err
}

// Get a list of all readings
func (r *ReadingRestClient) Readings() ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url, nil)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	// Response was not OK
	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)
		return []models.Reading{}, errors.New(string(bodyString))
	}

	return r.decodeReadingSlice(resp)
}

// Get the reading by id
func (r *ReadingRestClient) Reading(id string) (models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/"+id, nil)
	if err != nil {
		fmt.Println(err)
		return models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return models.Reading{}, errors.New(bodyString)
	}

	return r.decodeReading(resp)
}

// Get reading count
func (r *ReadingRestClient) ReadingCount() (int, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/count", nil)
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

// Get the readings for a device
func (r *ReadingRestClient) ReadingsForDevice(deviceId string, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/device/"+url.QueryEscape(deviceId)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Get the readings for name and device
func (r *ReadingRestClient) ReadingsForNameAndDevice(name string, deviceId string, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/name/"+url.QueryEscape(name)+"/device/"+url.QueryEscape(deviceId)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Get readings by name
func (r *ReadingRestClient) ReadingsForName(name string, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/name/"+url.QueryEscape(name)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Get readings for UOM Label
func (r *ReadingRestClient) ReadingsForUOMLabel(uomLabel string, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/uomlabel/"+url.QueryEscape(uomLabel)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Get readings for label
func (r *ReadingRestClient) ReadingsForLabel(label string, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/label/"+url.QueryEscape(label)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Get readings for type
func (r *ReadingRestClient) ReadingsForType(readingType string, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/type/"+url.QueryEscape(readingType)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Get readings for interval
func (r *ReadingRestClient) ReadingsForInterval(start int, end int, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/"+strconv.Itoa(start)+"/"+strconv.Itoa(end)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Get readings for device and value descriptor
func (r *ReadingRestClient) ReadingsForDeviceAndValueDescriptor(deviceId string, vd string, limit int) ([]models.Reading, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+"/device/"+url.QueryEscape(deviceId)+"/valuedescriptor/"+url.QueryEscape(vd)+"/"+strconv.Itoa(limit), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Reading{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Reading{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Reading{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Reading{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Reading{}, errors.New(bodyString)
	}
	return r.decodeReadingSlice(resp)
}

// Add a reading
func (r *ReadingRestClient) Add(reading *models.Reading) (string, error) {
	jsonStr, err := json.Marshal(reading)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, r.url, bytes.NewReader(jsonStr))
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

// Delete a reading by id
func (r *ReadingRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, r.url+"/id/"+id, nil)
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
