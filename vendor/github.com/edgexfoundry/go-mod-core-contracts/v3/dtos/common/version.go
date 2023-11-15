//
// Copyright (C) 2020 IOTech Ltd
// Copyright (C) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

// VersionResponse defines the latest version supported by the service.
type VersionResponse struct {
	Versionable `json:",inline"`
	Version     string `json:"version"`
	ServiceName string `json:"serviceName"`
}

// VersionSdkResponse defines the latest sdk version supported by the service.
type VersionSdkResponse struct {
	VersionResponse `json:",inline"`
	SdkVersion      string `json:"sdk_version"`
}

// NewVersionResponse creates new VersionResponse with all fields set appropriately
func NewVersionResponse(version string, serviceName string) VersionResponse {
	return VersionResponse{
		Versionable: NewVersionable(),
		Version:     version,
		ServiceName: serviceName,
	}
}

// NewVersionSdkResponse creates new VersionSdkResponse with all fields set appropriately
func NewVersionSdkResponse(appVersion string, sdkVersion string, serviceName string) VersionSdkResponse {
	return VersionSdkResponse{
		VersionResponse: NewVersionResponse(appVersion, serviceName),
		SdkVersion:      sdkVersion,
	}
}
