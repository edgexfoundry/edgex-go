/*******************************************************************************
 * Copyright 2023 Intel Corporation
 * Copyright 2023 IOTech Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"

	"github.com/labstack/echo/v4"
)

// VaultAuthenticationHandlerFunc prefixes an existing HandlerFunc
// with a Vault-based JWT authentication check.  Usage:
//
//	 authenticationHook := handlers.NilAuthenticationHandlerFunc()
//	 if secret.IsSecurityEnabled() {
//			lc := container.LoggingClientFrom(dic.Get)
//	     secretProvider := container.SecretProviderFrom(dic.Get)
//	     authenticationHook = handlers.VaultAuthenticationHandlerFunc(secretProvider, lc)
//	 }
//	 For optionally-authenticated requests
//	 r.HandleFunc("path", authenticationHook(handlerFunc)).Methods(http.MethodGet)
//
//	 For unauthenticated requests
//	 r.HandleFunc("path", handlerFunc).Methods(http.MethodGet)
//
// For typical usage, it is preferred to use AutoConfigAuthenticationFunc which
// will automatically select between a real and a fake JWT validation handler.
func VaultAuthenticationHandlerFunc(secretProvider interfaces.SecretProviderExt, lc logger.LoggingClient) echo.MiddlewareFunc {
	return func(inner echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			w := c.Response()
			authHeader := r.Header.Get("Authorization")
			lc.Debugf("Authorizing incoming call to '%s' via JWT (Authorization len=%d)", r.URL.Path, len(authHeader))
			authParts := strings.Split(authHeader, " ")
			if len(authParts) >= 2 && strings.EqualFold(authParts[0], "Bearer") {
				token := authParts[1]
				validToken, err := secretProvider.IsJWTValid(token)
				if err != nil {
					lc.Errorf("Error checking JWT validity: %v", err)
					// set Response.Committed to true in order to rewrite the status code
					w.Committed = false
					return echo.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				} else if !validToken {
					lc.Warnf("Request to '%s' UNAUTHORIZED", r.URL.Path)
					// set Response.Committed to true in order to rewrite the status code
					w.Committed = false
					return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}
				lc.Debugf("Request to '%s' authorized", r.URL.Path)
				return inner(c)
			}
			err := fmt.Errorf("unable to parse JWT for call to '%s'; unauthorized", r.URL.Path)
			lc.Errorf("%v", err)
			// set Response.Committed to true in order to rewrite the status code
			w.Committed = false
			return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}
	}
}

// NilAuthenticationHandlerFunc just invokes a nested handler
func NilAuthenticationHandlerFunc() echo.MiddlewareFunc {
	return func(inner echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return inner(c)
		}
	}
}

// AutoConfigAuthenticationFunc auto-selects between a HandlerFunc
// wrapper that does authentication and a HandlerFunc wrapper that does not.
// By default, JWT validation is enabled in secure mode
// (i.e. when using a real secrets provider instead of a no-op stub)
//
// Set EDGEX_DISABLE_JWT_VALIDATION to 1, t, T, TRUE, true, or True
// to disable JWT validation.  This might be wanted for an EdgeX
// adopter that wanted to only validate JWT's at the proxy layer,
// or as an escape hatch for a caller that cannot authenticate.
func AutoConfigAuthenticationFunc(secretProvider interfaces.SecretProviderExt, lc logger.LoggingClient) echo.MiddlewareFunc {
	// Golang standard library treats an error as false
	disableJWTValidation, _ := strconv.ParseBool(os.Getenv("EDGEX_DISABLE_JWT_VALIDATION"))
	authenticationHook := NilAuthenticationHandlerFunc()
	if secret.IsSecurityEnabled() && !disableJWTValidation {
		authenticationHook = VaultAuthenticationHandlerFunc(secretProvider, lc)
	}
	return authenticationHook
}
