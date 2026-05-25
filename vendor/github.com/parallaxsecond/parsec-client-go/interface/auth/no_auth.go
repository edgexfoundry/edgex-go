// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
)

type noAuthAuthenticator struct {
}

// NewNoAuthAuthenticator create a new authenticator that provides no authentication.
// Used for testing and for core operations such as list_providers and list_authenticators
func NewNoAuthAuthenticator() Authenticator {
	return &noAuthAuthenticator{}
}

// NewRequestAuth creates a new request authentication payload
func (a noAuthAuthenticator) NewRequestAuth() (RequestAuthToken, error) {
	r := &DefaultRequestAuthToken{buf: &bytes.Buffer{}, authType: AuthNoAuth}
	return r, nil
}

// GetType get the type of the authenticator
func (a *noAuthAuthenticator) GetType() AuthenticationType {
	return AuthNoAuth
}
