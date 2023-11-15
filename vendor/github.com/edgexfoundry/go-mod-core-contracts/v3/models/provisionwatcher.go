//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type ProvisionWatcher struct {
	DBTimestamp
	Id                  string
	Name                string
	ServiceName         string
	Labels              []string
	Identifiers         map[string]string
	BlockingIdentifiers map[string][]string
	AdminState          AdminState
	DiscoveredDevice    DiscoveredDevice
}
