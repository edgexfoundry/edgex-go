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
 * @author: Tingyu Zeng, Dell
 * @version: 1.1.0
 *******************************************************************************/
package proxy

import (
	"fmt"
	"net/url"

	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	Title         string
	KongURL       kongurl
	KongAuth      kongauth
	KongACL       KongACLPlugin
	SecretService secretservice
	EdgexServices map[string]service
}

type kongurl struct {
	Server             string
	AdminPort          string
	AdminPortSSL       string
	ApplicationPort    string
	ApplicationPortSSL string
}

type kongauth struct {
	Name       string
	TokenTTL   int
	Resource   string
	OutputPath string
}

type kongacl struct {
	Name      string
	WhiteList string
}

type secretservice struct {
	Server          string
	Port            string
	HealthcheckPath string
	CertPath        string
	TokenPath       string
	CACertPath      string
	SNIS            string
}

type service struct {
	Name     string
	Host     string
	Port     string
	Protocol string
}

func LoadTomlConfig(path string) (*tomlConfig, error) {
	config := tomlConfig{}
	_, err := toml.DecodeFile(path, &config)
	return &config, err
}

func (cfg *tomlConfig) GetCertPath() string {
	return cfg.SecretService.CertPath
}

func (cfg *tomlConfig) GetTokenPath() string {
	return cfg.SecretService.TokenPath
}

func (cfg *tomlConfig) GetProxyServerName() string {
	return cfg.KongURL.Server
}

func (cfg *tomlConfig) GetProxyServerPort() string {
	return cfg.KongURL.AdminPort
}

func (cfg *tomlConfig) GetProxyApplicationPortSSL() string {
	return cfg.KongURL.ApplicationPortSSL
}

func (cfg *tomlConfig) GetProxyAuthMethod() string {
	return cfg.KongAuth.Name
}

func (cfg *tomlConfig) GetProxyAuthTTL() int {
	return cfg.KongAuth.TokenTTL
}

func (cfg *tomlConfig) GetProxyAuthResource() string {
	return cfg.KongAuth.Resource
}

func (cfg *tomlConfig) GetProxyAuthOutputPath() string {
	return cfg.KongAuth.OutputPath
}

func (cfg *tomlConfig) GetProxyACLName() string {
	return cfg.KongACL.Name
}

func (cfg *tomlConfig) GetProxyACLWhiteList() string {
	return cfg.KongACL.WhiteList
}

func (cfg *tomlConfig) GetSecretSvcName() string {
	return cfg.SecretService.Server
}

func (cfg *tomlConfig) GetSecretSvcPort() string {
	return cfg.SecretService.Port
}

func (cfg *tomlConfig) GetSecretSvcSNIS() string {
	return cfg.SecretService.SNIS
}

func (cfg *tomlConfig) GetEdgeXSvcs() map[string]service {
	return cfg.EdgexServices
}

func (cfg *tomlConfig) GetProxyBaseURL() string {
	url := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", cfg.GetProxyServerName(), cfg.GetProxyServerPort()),
		Path:   "/",
	}
	return url.String()
}

func (cfg *tomlConfig) GetSecretSvcBaseURL() string {
	url := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%s", cfg.GetSecretSvcName(), cfg.GetSecretSvcPort()),
		Path:   "/",
	}
	return url.String()
}
