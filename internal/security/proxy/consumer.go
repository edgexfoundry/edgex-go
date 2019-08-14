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
 *******************************************************************************/
package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

type Consumer struct {
	name   string
	client internal.HttpCaller
}

func NewConsumer(name string, r internal.HttpCaller) Consumer {
	return Consumer{
		name:   name,
		client: r,
	}
}

func (c *Consumer) Delete() error {
	resource := NewResource(c.name, c.client)
	return resource.Remove(ConsumersPath)
}

func (c *Consumer) Create(service string) error {
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), ConsumersPath, c.name}
	req, err := http.NewRequest(http.MethodPut, strings.Join(tokens, "/"), nil)
	if err != nil {
		e := fmt.Sprintf("failed to create consumer %s for %s service with error %s", c.name, service, err.Error())
		LoggingClient.Error(e)
		return errors.New(e)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to create consumer %s for %s service with error %s", c.name, service, err.Error())
		LoggingClient.Error(e)
		return errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info(fmt.Sprintf("successful to create consumer %s for %s service", c.name, service))
		break
	default:
		e := fmt.Sprintf("failed to create consumer %s for %s service", c.name, service)
		LoggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (c *Consumer) AssociateWithGroup(g string) error {
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), ConsumersPath, c.name, "acls"}
	formVals := url.Values{
		"group": {g},
	}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to create group association request. consumer %s, group %s -- %s", c.name, g, err.Error())
		LoggingClient.Error(e)
		return errors.New(e)
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to associate consumer %s for with group %s with error %s", c.name, g, err.Error())
		LoggingClient.Error(e)
		return errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info(fmt.Sprintf("successful to associate consumer %s with group %s", c.name, g))
		break
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		e := fmt.Sprintf("failed to associate consumer %s with group %s with error %s,%s", c.name, g, resp.Status, string(b))
		LoggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (c *Consumer) CreateToken() (string, error) {
	switch Configuration.KongAuth.Name {
	case "jwt":
		LoggingClient.Info("autheticate the user with jwt authentication.")
		return c.createJWTToken()
	case "oauth2":
		LoggingClient.Info("authenticate the user with oauth2 authentication.")
		return c.createOAuth2Token()
	default:
		e := fmt.Sprintf("unknown authentication method provided: %s", Configuration.KongAuth.Name)
		LoggingClient.Error(e)
		return "", errors.New(e)
	}
}

func (c *Consumer) createJWTToken() (string, error) {
	jwtCred := JWTCred{}
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), ConsumersPath, c.name, "jwt"}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), nil)
	if err != nil {
		e := fmt.Sprintf("error creating JWT token request -- %s", err.Error())
		LoggingClient.Error(e)
		return "", errors.New(e)
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to create jwt token for consumer %s with error %s", c.name, err.Error())
		return "", errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		if err = json.NewDecoder(resp.Body).Decode(&jwtCred); err != nil {
			return "", err
		}
		LoggingClient.Info(fmt.Sprintf("successful on retrieving JWT credential for consumer %s", c.name))

		// Create the Claims
		claims := KongJWTClaims{
			jwtCred.Key,
			c.name,
			jwt.StandardClaims{
				Issuer: EdgeXKong,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(jwtCred.Secret))
	default:
		e := fmt.Sprintf("failed to create JWT for consumer %s with errorCode %d", c.name, resp.StatusCode)
		LoggingClient.Error(e)
		return "", errors.New(e)
	}
}

//createOAuth2Token implements the two curl command below. Using these two curl commands below against KONG to request OAuth2 token.
//curl -X POST "http://localhost:8001/consumers/user123/oauth2" -d "name=www.edgexfoundry.org" --data "client_id=user123" -d "client_secret=user123"  -d "redirect_uri=http://www.edgexfoundry.org/"
//curl -k -v https://localhost:8443/{service}/oauth2/token -d "client_id=user123" -d "grant_type=client_credentials" -d "client_secret=user123" -d "scope=email"
func (c *Consumer) createOAuth2Token() (string, error) {
	ko := &KongConsumerOauth2{
		Name:         EdgeXKong,
		ClientID:     c.name,
		ClientSecret: c.name,
		RedirectURIS: "http://" + EdgeXKong,
	}

	formVals := url.Values{
		"name":          {ko.Name},
		"client_id":     {ko.ClientID},
		"client_secret": {ko.ClientSecret},
		"redirect_uris": {ko.RedirectURIS},
	}

	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), ConsumersPath, c.name, "oauth2"}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to construct http POST form request: %s %s", c.name, err.Error())
		LoggingClient.Error(e)
		return "", err
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to enable oauth2 authentication for consumer %s with error %s", c.name, err.Error())
		LoggingClient.Error(e)
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info(fmt.Sprintf("successful on enabling oauth2 for consumer %s", c.name))

		// obtain token
		tokenreq := &KongOuath2TokenRequest{
			ClientID:     c.name,
			ClientSecret: c.name,
			GrantType:    OAuth2GrantType,
			Scope:        OAuth2Scopes,
		}

		formVals := url.Values{
			"client_id":     {tokenreq.ClientID},
			"client_secret": {tokenreq.ClientSecret},
			"grant_type":    {tokenreq.GrantType},
			"scope":         {tokenreq.Scope},
		}
		tokens := []string{Configuration.KongURL.GetSecureURL(), Configuration.KongAuth.Resource, "oauth2/token"}
		LoggingClient.Info(fmt.Sprintf("creating token on the endpoint of %s", strings.Join(tokens, "/")))

		req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("failed to create oauth2 token for client_id %s with error %s", c.name, err.Error()))
			return "", err
		}
		req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")
		tresp, err := c.client.Do(req)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("failed to create oauth2 token for client_id %s with error %s", c.name, err.Error()))
			return "", err
		}
		defer tresp.Body.Close()

		token := KongOauth2Token{}
		switch tresp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			if err = json.NewDecoder(tresp.Body).Decode(&token); err != nil {
				return "", err
			}
			LoggingClient.Info(fmt.Sprintf("successful on retrieving bearer credential for consumer %s", c.name))
			return token.AccessToken, nil
		default:
			b, err := ioutil.ReadAll(tresp.Body)
			if err != nil {
				return "", err
			}
			e := fmt.Sprintf("failed to create bearer token for oauth authentication at endpoint oauth2/token with error %s,%s", tresp.Status, string(b))
			LoggingClient.Error(e)
			return "", errors.New(e)
		}
	default:
		e := fmt.Sprintf("failed to enable oauth2 for consumer %s with error code %d", c.name, resp.StatusCode)
		LoggingClient.Error(e)
		return "", errors.New(e)
	}
}
