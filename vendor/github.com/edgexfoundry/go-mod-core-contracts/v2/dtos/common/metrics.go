//
// Copyright (C) 2020-2021 IOTech Ltd
// Copyright (C) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

type Metrics struct {
	MemAlloc       uint64 `json:"memAlloc"`
	MemFrees       uint64 `json:"memFrees"`
	MemLiveObjects uint64 `json:"memLiveObjects"`
	MemMallocs     uint64 `json:"memMallocs"`
	MemSys         uint64 `json:"memSys"`
	MemTotalAlloc  uint64 `json:"memTotalAlloc"`
	CpuBusyAvg     uint8  `json:"cpuBusyAvg"`
}

// MetricsResponse defines the providing memory and cpu utilization stats of the service.
// This object and its properties correspond to the MetricsResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/MetricsResponse
type MetricsResponse struct {
	Versionable `json:",inline"`
	Metrics     Metrics `json:"metrics"`
}

// NewMetricsResponse creates new MetricsResponse with all fields set appropriately
func NewMetricsResponse(metrics Metrics) MetricsResponse {
	return MetricsResponse{
		Versionable: NewVersionable(),
		Metrics:     metrics,
	}
}
