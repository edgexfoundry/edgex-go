/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

// Package ziti provides methods for loading Contexts which interact with an OpenZiti Controller via the Edge Client
// API to bind (host) services or dial (connect) to services.
//
// Each context is required to authenticate with the Edge Client API via Credentials instance. Credentials come in the
// form of identity files, username/password, JWTs, and more.
//
// Identity files specified in `ZITI_IDENTITIES` environment variable (semicolon separates) are loaded automatically
// at startup to populate the DefaultCollection. This behavior is deprecated, and explicit usage of an CtxCollection
// is suggested. This behavior can be replicated via NewSdkCollectionFromEnv().
package ziti

import (
	"github.com/kataras/go-events"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_model"
	edge_apis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/sdk-golang/ziti/edge/posture"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strconv"
)

var idCount = 0

// NewId will return a unique string id suitable for ziti.Context Id functionality.
func NewId() string {
	idCount = idCount + 1

	return strconv.Itoa(idCount)
}

// NewContextFromFile attempts to load a new Config from the provided path and then uses that
// config to instantiate a new Context. See NewConfigFromFile() for configuration file details.
func NewContextFromFile(path string) (Context, error) {
	return NewContextFromFileWithOpts(path, nil)
}

// NewContextFromFileWithOpts does the same as NewContextFromFile but allow Options to be supplied.
func NewContextFromFileWithOpts(path string, options *Options) (Context, error) {
	cfg, err := NewConfigFromFile(path)

	if err != nil {
		return nil, err
	}

	return NewContextWithOpts(cfg, options)
}

// NewContext creates a Context from the supplied Config with the default options. See NewContextWithOpts().
func NewContext(cfg *Config) (Context, error) {
	return NewContextWithOpts(cfg, nil)
}

// NewContextWithOpts creates a Context from the supplied Config and Options. The configuration requires
// either the `ID` field or the `Credentials` field to be populated. If both are supplied, the `ID` field is used.
func NewContextWithOpts(cfg *Config, options *Options) (Context, error) {
	if cfg == nil {
		return nil, errors.New("a config is required")
	}

	if options == nil {
		options = DefaultOptions
	}

	newContext := &ContextImpl{
		Id:                NewId(),
		routerConnections: cmap.New[edge.RouterConn](),
		options:           options,
		authQueryHandlers: map[string]func(query *rest_model.AuthQueryDetail, response MfaCodeResponse) error{},
		closeNotify:       make(chan struct{}),
		EventEmmiter:      events.New(),
		routerProxy:       cfg.RouterProxy,
	}

	if cfg.ID.Cert != "" && cfg.ID.Key != "" {
		idCredentials := edge_apis.NewIdentityCredentialsFromConfig(cfg.ID)
		idCredentials.ConfigTypes = cfg.ConfigTypes
		cfg.Credentials = idCredentials
	} else if cfg.Credentials == nil {
		return nil, errors.New("either cfg.ID or cfg.Credentials must be provided")
	}

	var apiStrs []string
	if len(cfg.ZtAPIs) > 0 {
		apiStrs = cfg.ZtAPIs
	} else {
		apiStrs = []string{cfg.ZtAPI}
	}

	var apiUrls []*url.URL
	for _, apiStr := range apiStrs {
		apiUrl, err := url.Parse(cfg.ZtAPI)

		if err != nil {
			return nil, errors.Wrapf(err, "could not parse ZtAPI from configuration as URI: %s", apiStr)
		}

		apiUrls = append(apiUrls, apiUrl)
	}

	apiClientConfig := &edge_apis.ApiClientConfig{
		ApiUrls: apiUrls,
		CaPool:  cfg.Credentials.GetCaPool(),
		TotpCallback: func(codeCh chan string) {
			provider := rest_model.MfaProvidersZiti

			authQuery := &rest_model.AuthQueryDetail{
				Provider: &provider,
			}

			newContext.Emit(EventAuthQuery, authQuery)

			if *authQuery.Provider == rest_model.MfaProvidersZiti {
				handler := newContext.authQueryHandlers[string(rest_model.MfaProvidersZiti)]

				newContext.Emit(EventMfaTotpCode, authQuery, MfaCodeResponse(newContext.authenticateMfa))

				if handler == nil {
					pfxlog.Logger().Debugf("no callback handler registered for provider: %v, event will still be emitted", *authQuery.Provider)
					return
				}

				_ = handler(authQuery, func(code string) error {
					codeCh <- code
					return nil
				})
			}
		},
		Proxy: cfg.CtrlProxy,
	}

	if apiClientConfig.Proxy == nil {
		apiClientConfig.Proxy = http.ProxyFromEnvironment
	}

	newContext.CtrlClt = &CtrlClient{
		ClientApiClient: edge_apis.NewClientApiClientWithConfig(apiClientConfig),
		Credentials:     cfg.Credentials,
		ConfigTypes:     cfg.ConfigTypes,
	}

	newContext.CtrlClt.ClientApiClient.SetAllowOidcDynamicallyEnabled(cfg.EnableHa)
	newContext.CtrlClt.PostureCache = posture.NewCache(newContext.CtrlClt, newContext.closeNotify)

	newContext.CtrlClt.AddOnControllerUpdateListeners(func(urls []*url.URL) {
		newContext.Emit(EventControllerUrlsUpdated, urls)
	})

	return newContext, nil
}
