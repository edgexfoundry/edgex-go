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
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/export"
)

const (
	clientPort int = 48071
)

func getRegistrationBaseURL(host string) string {
	return "http://" + host + ":" + strconv.Itoa(clientPort) +
		"/api/v1/registration"
}

func getRegistrations() ([]export.Registration, error) {
	url := getRegistrationBaseURL(configuration.ClientHost)
	return getRegistrationsURL(url)
}

func getRegistrationsURL(url string) ([]export.Registration, error) {
	response, err := http.Get(url)
	if err != nil {
		LoggingClient.Error(err.Error())
		LoggingClient.Warn(fmt.Sprintf("Error getting all registrations: %s",  url))
		return nil, err
	}
	defer response.Body.Close()

	// ensure we have an empty slice instead of a nil slice for better handling of JSON
	registrations := make([]export.Registration, 0)
	if err := json.NewDecoder(response.Body).Decode(&registrations); err != nil {
		LoggingClient.Error(err.Error())
		LoggingClient.Warn("Could not parse json")
		return nil, err
	}

	results := make([]export.Registration, 0)
	for _, reg := range registrations {
		if valid, err := reg.Validate(); valid {
			results = append(results, reg)
		} else {
			LoggingClient.Error(err.Error())
			LoggingClient.Warn("Could not validate registration")
		}
	}
	return results, nil
}

func getRegistrationByName(name string) *export.Registration {
	url := getRegistrationBaseURL(configuration.ClientHost) + "/name/" + name
	return getRegistrationByNameURL(url)
}

func getRegistrationByNameURL(url string) *export.Registration {

	response, err := http.Get(url)
	if err != nil {
		LoggingClient.Error(err.Error())
		LoggingClient.Error(fmt.Sprintf("Error getting all registrations: %s", url))
		return nil
	}
	defer response.Body.Close()

	reg := export.Registration{}
	if err := json.NewDecoder(response.Body).Decode(&reg); err != nil {
		LoggingClient.Error(err.Error())
		LoggingClient.Error("Could not parse json")
		return nil
	}

	if valid, err := reg.Validate(); !valid {
		LoggingClient.Error(err.Error())
		LoggingClient.Error("Failed to validate registrations fields")
		return nil
	}
	return &reg
}
