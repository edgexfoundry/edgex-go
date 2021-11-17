//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"strconv"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
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
func ProcessCORS(corsInfo config.CORSConfigurationInfo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			next.ServeHTTP(w, r)
		})
	}
}

// HandlePreflight returns a http handler function that process CORS preflight responses and sets CORS headers.
func HandlePreflight(corsInfo config.CORSConfigurationInfo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
