//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"net/http"
)

// AuthenticationInjector defines an interface to obtain a JWT and secure transport for remote service calls
type AuthenticationInjector interface {
	// AddAuthenticationData mutates an HTTP request to add authentication data
	// (suth as an Authorization: header) to an outbound HTTP request
	AddAuthenticationData(_ *http.Request) error
	// Returns the configured *http.Transport to use when making the request
	SecureTransportProvider
}

// SecureTransportProvider defines an interface to obtain a secure http.Transport to use when making http requests
type SecureTransportProvider interface {
	// Returns the configured *http.Transport to use when making the request
	RoundTripper() http.RoundTripper
}
