// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
	"encoding/binary"
	"os/user"
	"strconv"
)

type unixPeerAuthenticator struct {
}

// NewUnixPeerAuthenticator creates a new authenticator that uses current unix user id as
// means of authentication.
func NewUnixPeerAuthenticator() Authenticator {
	return &unixPeerAuthenticator{}
}

// NewRequestAuth creates a new request authentication payload
func (a unixPeerAuthenticator) NewRequestAuth() (RequestAuthToken, error) {
	r := &DefaultRequestAuthToken{buf: &bytes.Buffer{}, authType: AuthUnixPeerCredentials}
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseUint(currentUser.Uid, 10, 32) //nolint // base 10 and 32 bit number obvious from context
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	err = binary.Write(r.buf, binary.LittleEndian, uint32(uid))
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GetType get the type of the authenticator
func (a *unixPeerAuthenticator) GetType() AuthenticationType {
	return AuthUnixPeerCredentials
}
