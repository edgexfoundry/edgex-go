//go:build no_openziti

/*******************************************************************************
 * Copyright 2024-2025 IOTech Ltd
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
	"strings"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers/headers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// AuthenticationHandlerFunc prefixes an existing HandlerFunc,
// performing authentication checks based on OpenBao-issued JWTs or external JWTs by checking the Authorization header. Usage:
//
// authenticationHook := handlers.NilAuthenticationHandlerFunc()
//
//	if secret.IsSecurityEnabled() {
//		    authenticationHook = handlers.AuthenticationHandlerFunc(dic)
//		}
//		For optionally-authenticated requests
//		r.HandleFunc("path", authenticationHook(handlerFunc)).Methods(http.MethodGet)
//
//		For unauthenticated requests
//		r.HandleFunc("path", handlerFunc).Methods(http.MethodGet)
//
// For typical usage, it is preferred to use AutoConfigAuthenticationFunc which
// will automatically select between a real and a fake JWT validation handler.
func AuthenticationHandlerFunc(dic *di.Container) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lc := container.LoggingClientFrom(dic.Get)
			secretProvider := container.SecretProviderExtFrom(dic.Get)
			r := c.Request()
			w := c.Response()
			authHeader := r.Header.Get("Authorization")
			lc.Debugf("Authorizing incoming call to '%s' via JWT (Authorization len=%d), %v", r.URL.Path, len(authHeader), secretProvider.IsZeroTrustEnabled())

			if secretProvider.IsZeroTrustEnabled() {
				// this implementation will be pick up in the build when build tag no_openziti is specified, where
				// OpenZiti packages are not included and the Zero Trust feature is not available.
				lc.Info("zero trust was enabled, but service is built with no_openziti flag. falling back to token-based auth")
			}

			authParts := strings.Split(authHeader, " ")
			if len(authParts) >= 2 && strings.EqualFold(authParts[0], "Bearer") {
				token := authParts[1]

				parser := jwt.NewParser()
				parsedToken, _, jwtErr := parser.ParseUnverified(token, &jwt.MapClaims{})
				if jwtErr != nil {
					w.Committed = false
					return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}
				issuer, jwtErr := parsedToken.Claims.GetIssuer()
				if jwtErr != nil {
					w.Committed = false
					return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}

				var err edgexErr.EdgeX
				if issuer == openBaoIssuer {
					err = SecretStoreAuthenticationHandlerFunc(secretProvider, lc, token, c)
				} else {
					// Verify the JWT by invoking security-proxy-auth http client
					err = headers.VerifyJWT(token, issuer, parsedToken.Method.Alg(), dic, r.Context())
				}
				if err != nil {
					errResp := dtoCommon.NewBaseResponse("", err.Error(), err.Code())
					return c.JSON(err.Code(), errResp)
				} else {
					return next(c)
				}
			}
			err := fmt.Errorf("unable to parse JWT for call to '%s'; unauthorized", r.URL.Path)
			lc.Errorf("%v", err)
			// set Response.Committed to true in order to rewrite the status code
			w.Committed = false
			return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}
	}
}
