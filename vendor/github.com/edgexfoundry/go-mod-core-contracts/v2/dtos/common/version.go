//
// Copyright (C) 2020 IOTech Ltd
// Copyright (C) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

// VersionResponse defines the latest version supported by the service.
// This object and its properties correspond to the VersionResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/VersionResponse
type VersionResponse struct {
	Versionable `json:",inline"`
	Version     string `json:"version"`
}

// VersionSdkResponse defines the latest sdk version supported by the service.
// This object and its properties correspond to the VersionSdkResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/VersionSdkResponse
type VersionSdkResponse struct {
	VersionResponse `json:",inline"`
	SdkVersion      string `json:"sdk_version"`
}

// NewVersionResponse creates new VersionResponse with all fields set appropriately
func NewVersionResponse(version string) VersionResponse {
	return VersionResponse{
		Versionable: NewVersionable(),
		Version:     version,
	}
}

// NewVersionSdkResponse creates new VersionSdkResponse with all fields set appropriately
func NewVersionSdkResponse(appVersion string, sdkVersion string) VersionSdkResponse {
	return VersionSdkResponse{
		VersionResponse: NewVersionResponse(appVersion),
		SdkVersion:      sdkVersion,
	}
}
