/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
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
 *
 * @author: Tingyu Zeng, Dell
 *******************************************************************************/

package pkg

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
)

const httpClientTimeoutDuration = 10 * time.Second

type HTTPSRequester interface {
	Insecure() Caller
	WithTLS(io.Reader, string) Caller
}

type fluentRequester struct {
	logger logger.LoggingClient
}

func NewRequester(logger logger.LoggingClient) HTTPSRequester {
	return &fluentRequester{logger}
}

func (r *fluentRequester) Insecure() Caller {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint: gosec
	}
	return &http.Client{Timeout: httpClientTimeoutDuration, Transport: tr}
}

func (r *fluentRequester) WithTLS(caReader io.Reader, serverName string) Caller {
	readCloser := fileioperformer.MakeReadCloser(caReader)
	caCert, err := io.ReadAll(readCloser)
	defer readCloser.Close()
	if err != nil {
		r.logger.Error("failed to load rootCA certificate.")
		return nil
	}
	r.logger.Info("successful loading the rootCA certificate.")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
			ServerName:         serverName,
			MinVersion:         tls.VersionTLS12,
		},
		TLSHandshakeTimeout: httpClientTimeoutDuration,
	}
	return &http.Client{Timeout: httpClientTimeoutDuration, Transport: tr}
}

type mockRequester struct {
}

func NewMockRequester() *mockRequester {
	return &mockRequester{}
}

func (r *mockRequester) Insecure() Caller {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint: gosec
	}
	return &http.Client{Timeout: httpClientTimeoutDuration, Transport: tr}
}
