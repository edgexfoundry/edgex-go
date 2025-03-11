//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"strconv"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"

	"github.com/labstack/echo/v4"
)

const (
	Origin                        = "Origin"
	Vary                          = "Vary"
	AccessControlRequestMethod    = "Access-Control-Request-Method"
	AccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"
	AccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	AccessControlMaxAge           = "Access-Control-Max-Age"
)

// ProcessCORS is a middleware function that enables CORS responses and sets CORS headers.
func ProcessCORS(corsInfo config.CORSConfigurationInfo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			w := c.Response()
			if corsInfo.EnableCORS && r.Header.Get(Origin) != "" {
				// Set Access-Control-Expose-Headers only if it's not a preflight request
				// If the http method is OPTIONS with Access-Control-Request-Methods headers, it means a preflight request
				if !(r.Method == http.MethodOptions && r.Header.Get(AccessControlRequestMethod) != "") {
					if len(corsInfo.CORSExposeHeaders) > 0 {
						w.Header().Set(AccessControlExposeHeaders, corsInfo.CORSExposeHeaders)
					}
				}

				if len(corsInfo.CORSAllowedOrigin) > 0 {
					w.Header().Set(AccessControlAllowOrigin, corsInfo.CORSAllowedOrigin)
				}
				if corsInfo.CORSAllowCredentials {
					w.Header().Set(AccessControlAllowCredentials, "true")
				}

				w.Header().Set(Vary, Origin)
			}
			return next(c)
		}
	}
}

// HandlePreflight returns a http handler function that process CORS preflight responses and sets CORS headers.
func HandlePreflight(corsInfo config.CORSConfigurationInfo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			// skip the middleware if AccessControlRequestMethod header is not defined
			if r.Header.Get(AccessControlRequestMethod) == "" {
				return next(c)
			}

			w := c.Response()
			if corsInfo.EnableCORS && r.Header.Get(Origin) != "" {
				if len(corsInfo.CORSAllowedMethods) > 0 {
					w.Header().Set(AccessControlAllowMethods, corsInfo.CORSAllowedMethods)
				}
				if len(corsInfo.CORSAllowedHeaders) > 0 {
					w.Header().Set(AccessControlAllowHeaders, corsInfo.CORSAllowedHeaders)
				}
				if corsInfo.CORSMaxAge > 0 {
					w.Header().Set(AccessControlMaxAge, strconv.Itoa(corsInfo.CORSMaxAge))
				}
			}
			w.WriteHeader(http.StatusOK)
			return next(c)
		}
	}
}
