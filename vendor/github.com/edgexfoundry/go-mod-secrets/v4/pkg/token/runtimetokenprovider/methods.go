//go:build !non_delayedstart
// +build !non_delayedstart

//
// Copyright (c) 2022 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package runtimetokenprovider

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type runtimetokenprovider struct {
	ctx                  context.Context
	lc                   logger.LoggingClient
	runtimeTokenProvider types.RuntimeTokenProviderInfo
	getTLSConfig         func(context.Context, logger.LoggingClient) (*tls.Config, error)
	source               *workloadapi.X509Source
}

// NewRuntimeTokenProvider creates a new runtime token provider
func NewRuntimeTokenProvider(ctx context.Context, lc logger.LoggingClient,
	runtimeTokenProvider types.RuntimeTokenProviderInfo) RuntimeTokenProvider {
	instance := &runtimetokenprovider{
		ctx:                  ctx,
		lc:                   lc,
		runtimeTokenProvider: runtimeTokenProvider,
	}
	instance.getTLSConfig = instance.defaultGetTLSConfig
	return instance
}

// GetRawToken implements the RuntimeTokenProvider interface for getting the raw secret token for service
// with serviceKey and returns token and error if any
func (p *runtimetokenprovider) GetRawToken(serviceKey string) (string, error) {
	if !p.runtimeTokenProvider.Enabled {
		p.lc.Infof("runtime token provider not enabled for service %s", serviceKey)
		return "", nil
	}

	tlsConf, err := p.getTLSConfig(p.ctx, p.lc)
	if err != nil {
		return "", fmt.Errorf("failed to get TLS Config %s", err.Error())
	}

	if p.source != nil {
		defer p.source.Close()
	}

	tlsConf.MinVersion = tls.VersionTLS13

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConf,
		},
	}

	// obtain token from spiffe token provider with API call

	providerURL, err := p.runtimeTokenProvider.BuildProviderURL(pkg.SpiffeTokenProviderGetTokenAPI)
	if err != nil {
		return "", fmt.Errorf("failed to build token provider URL: %w", err)
	}

	// requiredSecrets are runtime created and is provided from the configuration
	requiredSecrets := strings.Split(p.runtimeTokenProvider.RequiredSecrets, ",")

	form := url.Values{
		"service_key":        []string{serviceKey},
		"known_secret_names": requiredSecrets,
		"raw_token":          []string{"true"},
	}
	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, providerURL, strings.NewReader(formVal))
	if err != nil {
		return "", fmt.Errorf("failed to prepare spiffe-token-provider gettoken request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send spiffe-token-provider gettoken request: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var rawToken string
	var gotTokenErr error
	switch resp.StatusCode {
	case http.StatusOK:
		rawToken = string(responseBody)
		p.lc.Info("successfully got token from spiffe-token-provider!")
	default:
		gotTokenErr = fmt.Errorf("failed to get spiffe-token-provider gettoken api call, status code = %d ", resp.StatusCode)
	}

	return rawToken, gotTokenErr
}

func (p *runtimetokenprovider) SetTLSConfigFunc(tlsConfFunc func(context.Context, logger.LoggingClient) (*tls.Config, error)) {
	p.getTLSConfig = tlsConfFunc
}

func (p *runtimetokenprovider) defaultGetTLSConfig(ctx context.Context, lc logger.LoggingClient) (*tls.Config, error) {
	udsSocket := p.runtimeTokenProvider.EndpointSocket
	if !strings.HasPrefix(p.runtimeTokenProvider.EndpointSocket, "unix://") {
		udsSocket = "unix://" + p.runtimeTokenProvider.EndpointSocket
	}

	lc.Infof("using Unix Domain Socket at %s", udsSocket)

	// command-line equivalent: spire-agent api fetch x509 -socketPath xxxx
	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket path
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(udsSocket)))
	if err != nil {
		return nil, fmt.Errorf("Unable to create X509Source: %v", err)
	}

	// need to cache the source so that we can close it later after https TLS calls
	p.source = source

	lc.Info("workload got X509 source")

	td, err := spiffeid.TrustDomainFromString(p.runtimeTokenProvider.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("could not get SPIFFE trust domain from string '%s': %v", p.runtimeTokenProvider.TrustDomain, err)
	}

	tlsConf := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeMemberOf(td))

	return tlsConf, nil
}
