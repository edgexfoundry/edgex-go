// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
)

type directAuthenticator struct {
	appName string
}

// NewDirectAuthenticator creates a new Direct authenticator that uses appName as the
// means of authentication.
func NewDirectAuthenticator(appName string) Authenticator {
	return &directAuthenticator{appName: appName}
}

// NewRequestAuth creates a new request authentication payload
func (a *directAuthenticator) NewRequestAuth() (RequestAuthToken, error) {
	buf := &bytes.Buffer{}
	_, err := buf.WriteString(a.appName)
	if err != nil {
		return nil, err
	}
	r := &DefaultRequestAuthToken{buf: buf, authType: AuthDirect}
	return r, nil
}

// GetType get the type of the authenticator
func (a *directAuthenticator) GetType() AuthenticationType {
	return AuthDirect
}
