//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package interfaces

const (
	// JwtTokenType is passed on command line to select JWT auth tokens
	JwtTokenType string = "jwt"
	// OAuth2TokenType is passed on command line to select OAuth2 auth tokens
	OAuth2TokenType string = "oauth2"
	// RS256 JWT Alg RS256
	RS256 string = "RS256"
	// ES256 JWT Alt ES256
	ES256 string = "ES256"
)
