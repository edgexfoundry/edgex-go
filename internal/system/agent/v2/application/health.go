//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-registry/v2/registry"
)

const healthy string = "healthy"

func GetHealth(services []string, registryClient registry.Client) map[string]string {
	health := make(map[string]string)
	for _, service := range services {
		if registryClient == nil {
			health[service] = "registry is required to obtain service health status."
			continue
		}

		// the registry service returns nil for a healthy service
		ok, err := registryClient.IsServiceAvailable(service)
		if err != nil {
			health[service] = err.Error()
			continue
		}
		if !ok {
			health[service] = fmt.Sprintf("service %s is not available", service)
			continue
		}
		health[service] = healthy
	}

	return health
}
