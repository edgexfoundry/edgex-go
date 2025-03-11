//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type Registration struct {
	DBTimestamp
	ServiceId   string
	Status      string
	Host        string
	Port        int
	HealthCheck HealthCheck
}

type HealthCheck struct {
	Interval string
	Path     string
	Type     string
}
