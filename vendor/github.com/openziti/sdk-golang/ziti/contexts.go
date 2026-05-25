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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"

	"github.com/kataras/go-events"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_model"
	edgeApis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/sdk-golang/ziti/edge/posture"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var idCount atomic.Int64

// NewId will return a unique string id suitable for ziti.Context Id functionality.
func NewId() string {
	id := idCount.Add(1)
	return strconv.Itoa(int(id))
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
		Id:                    NewId(),
		routerConnections:     cmap.New[edge.RouterConn](),
		options:               options,
		closeNotify:           make(chan struct{}),
		EventEmmiter:          events.New(),
		routerProxy:           cfg.RouterProxy,
		maxDefaultConnections: int(cfg.MaxDefaultConnections),
		maxControlConnections: int(cfg.MaxControlConnections),
		services:              cmap.New[*rest_model.ServiceDetail](),
		sessions:              cmap.New[*rest_model.SessionDetail](),
		intercepts:            cmap.New[*edge.InterceptV1Config](),
		activeBinds:           cmap.New[*rest_model.ServiceDetail](),
		activeDials:           cmap.New[*rest_model.ServiceDetail](),
		listenerManagers:      cmap.New[*listenerManager](),
	}

	if newContext.maxDefaultConnections < 1 {
		newContext.maxDefaultConnections = 1
	}

	if cfg.Credentials == nil {
		idCredentials := edgeApis.NewIdentityCredentialsFromConfig(cfg.ID)
		idCredentials.ConfigTypes = cfg.ConfigTypes
		cfg.Credentials = idCredentials
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
			return nil, fmt.Errorf("could not parse ZtAPI from configuration as URI: %s: %w", apiStr, err)
		}

		apiUrls = append(apiUrls, apiUrl)
	}

	apiClientConfig := &edgeApis.ApiClientConfig{
		ApiUrls: apiUrls,
		CaPool:  cfg.Credentials.GetCaPool(),
		TotpCodeProvider: edgeApis.NewTotpCodeProviderFromChStringFunc(func(codeCh chan string) {
			provider := rest_model.MfaProvidersZiti

			authQuery := &rest_model.AuthQueryDetail{
				TypeID:     rest_model.AuthQueryTypeMFA,
				Format:     rest_model.MfaFormatsAlphaNumeric,
				HTTPMethod: http.MethodPost,
				HTTPURL:    EdgeClientTotpAuthEndpoint,
				MaxLength:  MaxTotpCodeLength,
				MinLength:  MinTotpCodeLength,
				Provider:   &provider,
			}

			newContext.Emit(EventAuthQuery, authQuery)

			if *authQuery.Provider == rest_model.MfaProvidersZiti {
				newContext.Emit(EventMfaTotpCode, authQuery, MfaCodeResponse(func(code string) error {
					codeCh <- code
					return nil
				}))

				if newContext.Events().ListenerCount(EventMfaTotpCode) == 0 {
					pfxlog.Logger().Debugf("no callback handler registered for provider: %v, event will still be emitted", *authQuery.Provider)
					return

				}
			}
		}),
		TotpEnrollmentProvider: edgeApis.TotpEnrollmentProviderFunc(func(provisioningUrl string) <-chan edgeApis.TotpEnrollmentResult {
			resultCh := make(chan edgeApis.TotpEnrollmentResult, 1)

			if newContext.Events().ListenerCount(EventMfaTotpEnrollment) == 0 {
				resultCh <- edgeApis.TotpEnrollmentResult{
					Err: errors.New("totp enrollment is required but no enrollment provider has been added via zitiContext.Events().AddMfaTotpEnrollmentListener()"),
				}
				return resultCh
			}

			newContext.Emit(EventMfaTotpEnrollment, provisioningUrl, MfaTotpEnrollmentResponse(func(code string, err error) {
				resultCh <- edgeApis.TotpEnrollmentResult{
					Code: code,
					Err:  err,
				}
			}))

			return resultCh
		}),
		Proxy: cfg.CtrlProxy,
	}

	if apiClientConfig.Proxy == nil {
		apiClientConfig.Proxy = http.ProxyFromEnvironment
	}

	newContext.CtrlClt = &CtrlClient{
		ClientApiClient: edgeApis.NewClientApiClientWithConfig(apiClientConfig),
		Credentials:     cfg.Credentials,
		ConfigTypes:     cfg.ConfigTypes,
	}

	newContext.CtrlClt.SetAllowOidcDynamicallyEnabled(true)

	multiSubmitter := posture.NewMultiSubmitter(newContext.CtrlClt, newContext.CtrlClt, newContext)
	totpTokenProvider := edgeApis.NewSingularTokenRequestor(newContext, newContext.CtrlClt)
	newContext.CtrlClt.PostureCache = posture.NewCache(newContext, multiSubmitter, totpTokenProvider, newContext.closeNotify)

	newContext.CtrlClt.AddOnControllerUpdateListeners(func(urls []*url.URL) {
		newContext.Emit(EventControllerUrlsUpdated, urls)
	})

	return newContext, nil
}
