//
// Copyright (c) 2017
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
)

//TODO: Since this is a service-to-service client, it should be in /pkg/clients/export
func getRegistrations() ([]export.Registration, error) {
	url := Configuration.Clients["Export"].Url() + clients.ApiRegistrationRoute
	return getRegistrationsURL(url)
}

func getRegistrationsURL(url string) ([]export.Registration, error) {
	response, err := http.Get(url)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error getting all registrations: %s. Error: %s", url, err.Error()))
		return nil, err
	}
	defer response.Body.Close()

	// ensure we have an empty slice instead of a nil slice for better handling of JSON
	registrations := make([]export.Registration, 0)
	if err := json.NewDecoder(response.Body).Decode(&registrations); err != nil {
		LoggingClient.Error(fmt.Sprintf("Could not parse json. Error: %s", err.Error()))
		return nil, err
	}

	results := make([]export.Registration, 0)
	for _, reg := range registrations {
		if valid, err := reg.Validate(); valid {
			results = append(results, reg)
		} else {
			LoggingClient.Error(fmt.Sprintf("Could not validate registration. Error: %s", err.Error()))
		}
	}
	return results, nil
}

func getRegistrationByName(name string) *export.Registration {
	url := fmt.Sprintf("%s%s/%s", Configuration.Clients["Export"].Url(), clients.ApiRegistrationByNameRoute, name)
	return getRegistrationByNameURL(url)
}

func getRegistrationByNameURL(url string) *export.Registration {

	response, err := http.Get(url)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error getting all registrations: %s. Error: %s", url, err.Error()))
		return nil
	}
	defer response.Body.Close()

	reg := export.Registration{}
	if err := json.NewDecoder(response.Body).Decode(&reg); err != nil {
		LoggingClient.Error(fmt.Sprintf("Could not parse json. Error: %s", err.Error()))
		return nil
	}

	if valid, err := reg.Validate(); !valid {
		LoggingClient.Error(fmt.Sprintf("Failed to validate registrations fields. Error: %s", err.Error()))
		return nil
	}
	return &reg
}
