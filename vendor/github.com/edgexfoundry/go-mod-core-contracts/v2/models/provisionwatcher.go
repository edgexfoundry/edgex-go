//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// ProvisionWatcher and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.x#/ProvisionWatcher
// Model fields are same as the DTOs documented by this swagger. Exceptions, if any, are noted below.
type ProvisionWatcher struct {
	DBTimestamp
	Id                  string
	Name                string
	Labels              []string
	Identifiers         map[string]string
	BlockingIdentifiers map[string][]string
	ProfileName         string
	ServiceName         string
	AdminState          AdminState
	AutoEvents          []AutoEvent
}
