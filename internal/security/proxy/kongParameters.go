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

import jwt "github.com/dgrijalva/jwt-go"

type KongService struct {
	Name     string `url:"name,omitempty"`
	Host     string `url:"host,omitempty"`
	Port     string `url:"port,omitempty"`
	Protocol string `url:"protocol,omitempty"`
}

// KongServiceResponse is the response from Kong when creating a service
type KongServiceResponse struct {
	ID             string `json:"id,omitempty"`
	CreatedAt      uint64 `json:"created_at,omitempty"`
	UpdatedAt      uint64 `json:"updated_at,omitempty"`
	ConnectTimeout int64  `json:"connect_timeout,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	Host           string `json:"host,omitempty"`
	Port           uint64 `json:"port,omitempty"`
	Path           string `json:"path,omitempty"`
	Name           string `json:"name,omitempty"`
	Retries        int64  `json:"retries,omitempty"`
	ReadTimeout    int64  `json:"read_timeout,omitempty"`
	WriteTimeout   int64  `json:"write_timeout,omitempty"`
}

type KongRoute struct {
	Paths []string `json:"paths,omitempty"`
	Name  string   `json:"name,omitempty"`
}

type KongJWTPlugin struct {
	Name string `url:"name,omitempty"`
}

type KongOAuth2Plugin struct {
	Name                    string `url:"name"`
	Scope                   string `url:"config.scopes"`
	MandatoryScope          string `url:"config.mandatory_scope"`
	EnableClientCredentials string `url:"config.enable_client_credentials"`
	EnableGlobalCredentials string `url:"config.global_credentials"`
	TokenTTL                int    `url:"config.refresh_token_ttl"`
}

type KongConsumerOauth2 struct {
	Name         string `url:"name,omitempty"`
	ClientID     string `url:"client_id,omitempty"`
	ClientSecret string `url:"client_secret,omitempty"`
	RedirectURIS string `url:"redirect_uris,omitempty"`
}

type KongOuath2TokenRequest struct {
	ClientID     string `url:"client_id,omitempty"`
	ClientSecret string `url:"client_secret,omitempty"`
	GrantType    string `url:"grant_type,omitempty"`
	Scope        string `url:"scope,omitempty"`
}

type KongOauth2Token struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
}

type KongACLPlugin struct {
	Name      string `url:"name"`
	WhiteList string `url:"config.whitelist"`
}

type KongBasicAuthPlugin struct {
	Name            string `url:"name,omitempty"`
	HideCredentials string `url:"config.hide_credentials,omitempty"`
}

type CertInfo struct {
	Cert string   `json:"cert,omitempty"`
	Key  string   `json:"key,omitempty"`
	Snis []string `json:"snis,omitempty"`
}

type JWTCred struct {
	ConsumerID string `json:"consumer_id,omitempty"`
	CreatedAt  int    `json:"created_at,omitempty"`
	ID         string `json:"id,omitempty"`
	Key        string `json:"key,omitempty"`
	Secret     string `json:"secret,omitempty"`
}

type KongJWTClaims struct {
	ISS  string `json:"iss"`
	Acct string `json:"account"`
	jwt.StandardClaims
}

type Item struct {
	ID string `json:"id"`
}

type DataCollect struct {
	Section []Item `json:"data"`
}
