/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
 *******************************************************************************/

package proxy

import (
	"fmt"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
)

type ConfigurationStruct struct {
	Writable      WritableInfo
	Logging       config.LoggingInfo
	KongURL       KongUrlInfo
	KongAuth      KongAuthInfo
	KongACL       KongAclInfo
	SecretService SecretServiceInfo
	Clients       map[string]config.ClientInfo
}

type WritableInfo struct {
	LogLevel       string
	RequestTimeout int
}

type KongUrlInfo struct {
	Server             string
	AdminPort          int
	AdminPortSSL       int
	ApplicationPort    int
	ApplicationPortSSL int
}

func (k KongUrlInfo) GetProxyBaseURL() string {
	url := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%v", k.Server, k.AdminPort),
	}
	return url.String()
}

func (k KongUrlInfo) GetSecureURL() string {
	url := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%v", k.Server, k.ApplicationPortSSL),
	}
	return url.String()
}

type KongAuthInfo struct {
	Name       string
	TokenTTL   int
	Resource   string
	OutputPath string
}

type KongAclInfo struct {
	Name      string
	WhiteList string
}

type SecretServiceInfo struct {
	Server          string
	Port            int
	HealthcheckPath string
	CertPath        string
	TokenPath       string
	CACertPath      string
	SNIS            []string
}

func (s SecretServiceInfo) GetSecretSvcBaseURL() string {
	url := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%v", s.Server, s.Port),
	}
	return url.String()
}
