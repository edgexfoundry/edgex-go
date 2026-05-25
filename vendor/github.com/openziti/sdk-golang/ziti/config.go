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

package ziti

import (
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/openziti/edge-api/rest_util"
	"github.com/openziti/identity"
	apis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
)

type Config struct {
	//ZtAPI should be in the form of https://<domain>[:<port>]/edge/client/v1. For backwards compatability with single controller identities
	ZtAPI string `json:"ztAPI"`

	//ZtAPIs is an array of ZtAPI values, supersedes `ZtAPI`. ZtAPIs is used to make an initial connection to a controller.
	ZtAPIs []string `json:"ztAPIs"`

	//ConfigTypes is an array of string configuration types that will be requested from the controller
	//for services.
	ConfigTypes []string `json:"configTypes"`

	//The ID field allows configurations is maintained for backwards compatability with previous SDK versions.
	//If set, it will be used to set the Credentials field.
	ID identity.Config `json:"id"`

	//The Credentials field is used to authenticate with the Edge Client API. If the ID field is set, it will be used
	//to populate this field with credentials.
	Credentials apis.Credentials `json:"-"`

	// EnableHa is now deprecated and will be removed in a future revision. The SDK will now query the controller
	// and use OIDC authentication if its enabled. This flag is now ignored.
	//
	// Deprecated: This flag is now ignored
	//
	EnableHa bool `json:"enableHa"`

	//Allows providing a function which controls how/where request to a controller are proxied.
	//See [http.Transport.Proxy] for more information
	//If this value is nil, [http.ProxyFromEnvironment] is used. If you never want a proxy to be used,
	//set a function which always returns nil.
	CtrlProxy func(*http.Request) (*url.URL, error) `json:"-"`

	//Allows providing a function which controls how/where connections to a router are proxied.
	RouterProxy func(addr string) *transport.ProxyConfiguration `json:"-"`

	// If set to a number greater than one, the sdk will attempt to create multiple connections to edge routers.
	// This configuration value should not be considered part of the stable API yet. It currently defaults to one,
	// but it may default to a larger number at some point in the future or be removed. If set to zero, it will
	// be reset to one.
	MaxDefaultConnections uint32 `json:"-"`

	// If set to a number greater than zero, the sdk will attempt to create one or more separate connection to
	// each edge routers for control plane data, such as dials. This configuration value should not be considered
	// part of the stable API yet. It currently defaults to zero, but it may default to 1 at some point in the future
	// or be removed.
	MaxControlConnections uint32 `json:"-"`
}

func (cfg *Config) SetMaxControlConnections(val uint32) {
	cfg.MaxControlConnections = val
}

func (cfg *Config) SetMaxDefaultConnections(val uint32) {
	cfg.MaxDefaultConnections = val
}

// NewConfig will create a new Config object from a provided Ziti Edge Client API URL and identity configuration.
// The Ziti Edge Client API is usually in the format of `https://host:port/edge/client/v1`.
func NewConfig(ztApi string, idConfig identity.Config) *Config {
	return &Config{
		ZtAPI: ztApi,
		ID:    idConfig,
	}
}

// NewConfigFromFile attempts to load a Config object from the provided path.
//
// The file that is indicated should be in the following format:
// ```
//
//	{
//	  "ztAPI": "https://ziti.controller.example.com/edge/client/v1",
//	  "configTypes": ["config1", "config2"],
//	  "id": { "cert": "...", "key": "..." },
//	}
//
// ```
func NewConfigFromFile(confFile string) (*Config, error) {
	conf, err := os.ReadFile(confFile)
	if err != nil {
		return nil, errors.Errorf("config file (%s) is not found ", confFile)
	}

	c := Config{}
	err = json.Unmarshal(conf, &c)

	if err != nil {
		return nil, errors.Errorf("failed to load ziti configuration (%s): %v", confFile, err)
	}

	return &c, nil
}

// GetControllerWellKnownCaPool will return a x509.CertPool. The target controller will not be verified via TLS and
// must be verified by some other means (i.e. enrollment JWT token).
//
// WARNING: This call is unauthenticated and should only be used for example purposes or explicitly when an unauthenticated
// request is required.
func GetControllerWellKnownCaPool(controllerAddr string) (*x509.CertPool, error) {
	return rest_util.GetControllerWellKnownCaPool(controllerAddr)
}
