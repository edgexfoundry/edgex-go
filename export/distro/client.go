//
// Copyright (c) 2017
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/export"

	"go.uber.org/zap"
)

const (
	clientPort int = 48071
)

func getRegistrationBaseURL(host string) string {
	return "http://" + host + ":" + strconv.Itoa(clientPort) +
		"/api/v1/registration"
}

func getRegistrations() []export.Registration {
	url := getRegistrationBaseURL(cfg.ClientHost)
	return getRegistrationsURL(url)
}

func getRegistrationsURL(url string) []export.Registration {
	response, err := http.Get(url)
	if err != nil {
		logger.Warn("Error getting all registrations", zap.String("url", url))
		return nil
	}
	defer response.Body.Close()

	registrations := []export.Registration{}
	if err := json.NewDecoder(response.Body).Decode(&registrations); err != nil {
		logger.Warn("Could not parse json", zap.Error(err))
		return nil
	}

	results := registrations[:0]
	for _, reg := range registrations {
		if reg.Validate() {
			results = append(results, reg)
		}
	}
	return results
}

func getRegistrationByName(name string) *export.Registration {
	url := getRegistrationBaseURL(cfg.ClientHost) + "/name/" + name
	return getRegistrationByNameURL(url)
}

func getRegistrationByNameURL(url string) *export.Registration {

	response, err := http.Get(url)
	if err != nil {
		logger.Error("Error getting all registrations", zap.String("url", url))
		return nil
	}
	defer response.Body.Close()

	reg := export.Registration{}
	if err := json.NewDecoder(response.Body).Decode(&reg); err != nil {
		logger.Error("Could not parse json", zap.Error(err))
		return nil
	}

	if !reg.Validate() {
		logger.Error("Failed to validate registrations fields")
		return nil
	}
	return &reg
}
