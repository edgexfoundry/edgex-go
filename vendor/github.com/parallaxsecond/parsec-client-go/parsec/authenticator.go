// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package parsec

import "github.com/parallaxsecond/parsec-client-go/interface/auth"

// AuthenticatorType enum to identify authenticators
type AuthenticatorType uint8

// Authenticator Types
const (
	AuthNoAuth AuthenticatorType = AuthenticatorType(auth.AuthNoAuth)
	// Direct authentication
	AuthDirect AuthenticatorType = AuthenticatorType(auth.AuthDirect)
	// JSON Web Tokens (JWT) authentication (not currently supported)
	AuthJwt AuthenticatorType = AuthenticatorType(auth.AuthJwt)
	// Unix peer credentials authentication
	AuthUnixPeerCredentials AuthenticatorType = AuthenticatorType(auth.AuthUnixPeerCredentials)
	// Authentication verifying a JWT SPIFFE Verifiable Identity Document
	AuthJwtSvid AuthenticatorType = AuthenticatorType(auth.AuthJwtSvid)
)

// AuthenticatorInfo contains information about an authenticator.
// Id is the id used to select the authenticator
// Name name of the authenticator
type AuthenticatorInfo struct {
	ID          AuthenticatorType
	Description string
	VersionMaj  uint32
	VersionMin  uint32
	VersionRev  uint32
}

// Authenticator object providing authenticator functionality to the basic client.
type Authenticator interface {
	toNativeAuthenticator() auth.Authenticator
	// GetAuthenticatorType return the type of this authenticator.
	GetAuthenticatorType() AuthenticatorType
}

// Internal implementation of authenticator - just wrapps the interface version.
type authenticatorWrapper struct {
	nativeAuth auth.Authenticator
}

func (w *authenticatorWrapper) toNativeAuthenticator() auth.Authenticator {
	return w.nativeAuth
}

// GetAuthenticatorType return the type of this authenticator.
func (w *authenticatorWrapper) GetAuthenticatorType() AuthenticatorType {
	return AuthenticatorType(w.nativeAuth.GetType())
}

// NewNoAuthAuthenticator creates an authenticator that does no authentication.  Used only for testing,
// or for initial connection when discovering the available authenticators to select a default.
func NewNoAuthAuthenticator() Authenticator {
	return &authenticatorWrapper{
		nativeAuth: auth.NewNoAuthAuthenticator(),
	}
}

// NewDirectAuthenticator creates an authenticator which uses the supplied appName as the means of authentication
// with the parsec service
func NewDirectAuthenticator(appName string) Authenticator {
	return &authenticatorWrapper{
		nativeAuth: auth.NewDirectAuthenticator(appName),
	}
}

// NewUnixPeerAuthenticator creates a new authenticator which uses current logged in user id as authentication
// to the parsec service
func NewUnixPeerAuthenticator() Authenticator {
	return &authenticatorWrapper{
		nativeAuth: auth.NewUnixPeerAuthenticator(),
	}
}
