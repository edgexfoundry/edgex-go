//
// Copyright (c) 2022 One Track Consulting
// Copyright (c) 2023 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

//go:build include_nats_messaging

package nats

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/nats-io/nats.go"
)

// ClientConfig contains all the configurations for the NATS client.
type ClientConfig struct {
	BrokerURL string
	ClientOptions
}

// ConnectionOptions contains the connection configurations for the NATS client.
//
// NOTE: The connection properties resides in its own struct in order to avoid the property being loaded in via
//
//	reflection during the load process.
type ConnectionOptions struct {
	BrokerURL string
}

// ClientOptions contains the client options which are loaded via reflection
type ClientOptions struct {
	// Client Identifiers
	Username             string
	Password             string
	ClientId             string
	Format               string
	RetryOnFailedConnect bool
	Durable              string
	Subject              string
	AutoProvision        bool
	ConnectTimeout       int // Seconds
	pkg.TlsConfigurationOptions
	QueueGroup              string
	Deliver                 string
	DefaultPubRetryAttempts int
	NKeySeedFile            string
	CredentialsFile         string
	ExactlyOnce             bool
}

// CreateClientConfiguration constructs a ClientConfig based on the provided MessageBusConfig.
func CreateClientConfiguration(messageBusConfig types.MessageBusConfig) (ClientConfig, error) {
	var brokerUrl string
	if !messageBusConfig.Broker.IsHostInfoEmpty() {
		brokerUrl = messageBusConfig.Broker.GetHostURL()
	} else {
		return ClientConfig{}, errors.New("broker information no specified")
	}

	_, err := url.Parse(brokerUrl)
	if err != nil {
		return ClientConfig{}, pkg.NewBrokerURLErr(fmt.Sprintf("Failed to parse broker: %v", err))
	}

	clientOptions := CreateClientOptionsWithDefaults()
	err = pkg.Load(messageBusConfig.Optional, &clientOptions)
	if err != nil {
		return ClientConfig{}, err
	}

	tlsConfig := pkg.TlsConfigurationOptions{}
	err = pkg.Load(messageBusConfig.Optional, &tlsConfig)
	if err != nil {
		return ClientConfig{}, err
	}

	clientOptions.TlsConfigurationOptions = tlsConfig

	return ClientConfig{
		BrokerURL:     brokerUrl,
		ClientOptions: clientOptions,
	}, nil
}

func (cc ClientConfig) ConnectOpt() ([]nats.Option, error) {
	connectTimeout := time.Second * 30

	if cc.ConnectTimeout != 0 {
		connectTimeout = time.Duration(cc.ConnectTimeout) * time.Second
	}

	opts := []nats.Option{
		nats.Timeout(connectTimeout),
		nats.RetryOnFailedConnect(cc.RetryOnFailedConnect),
	}

	if cc.ClientId != "" {
		opts = append(opts, nats.Name(cc.ClientId))
	}

	if cc.Username != "" {
		opts = append(opts, nats.UserInfo(cc.Username, cc.Password))
	}

	if tlsConfiguration, err := pkg.GenerateTLSForClientClientOptions(cc.BrokerURL, cc.TlsConfigurationOptions,
		tls.X509KeyPair, tls.LoadX509KeyPair, x509.ParseCertificate, os.ReadFile, pem.Decode); err == nil {
		if tlsConfiguration != nil {
			opts = append(opts, nats.Secure(tlsConfiguration))
		}
	} else {
		return nil, err
	}

	if cc.NKeySeedFile != "" {
		nkOpt, err := nats.NkeyOptionFromSeed(cc.NKeySeedFile)

		if err != nil {
			return nil, err
		}
		opts = append(opts, nkOpt)
	}

	if cc.CredentialsFile != "" {
		opts = append(opts, nats.UserCredentials(cc.CredentialsFile))
	}
	return opts, nil
}

// CreateClientOptionsWithDefaults constructs ClientOptions instance with defaults.
func CreateClientOptionsWithDefaults() ClientOptions {
	return ClientOptions{
		Username:                "",
		Password:                "",
		ConnectTimeout:          5, // 5 seconds
		RetryOnFailedConnect:    false,
		Durable:                 "",
		Subject:                 "",
		AutoProvision:           false, // AutoProvision JetStream streams - should maybe be true?
		TlsConfigurationOptions: pkg.CreateDefaultTlsConfigurationOptions(),
		DefaultPubRetryAttempts: nats.DefaultPubRetryAttempts,
		Format:                  "nats",
		NKeySeedFile:            "",
		CredentialsFile:         "",
		Deliver:                 "new",
		ExactlyOnce:             false, //could use QOS = 2 here
	}
}
