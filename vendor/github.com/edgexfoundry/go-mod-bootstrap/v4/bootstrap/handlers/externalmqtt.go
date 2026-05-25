//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

type ExternalMQTT struct {
	onConnectHandler mqtt.OnConnectHandler
}

func NewExternalMQTT(onConnectHandler mqtt.OnConnectHandler) *ExternalMQTT {
	return &ExternalMQTT{onConnectHandler: onConnectHandler}
}

func (e *ExternalMQTT) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := container.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	brokerConfig := configuration.GetBootstrap().ExternalMQTT
	topics := brokerConfig.SubscribeTopics
	// TODO: remove first check and update error log in EdgeX 3.0
	if len(strings.TrimSpace(topics)) == 0 && len(brokerConfig.Topics) == 0 {
		lc.Errorf("missing SubscribeTopics and/or Topics for external MQTT connection. Must be present in [ExternalMqtt] section")
		return false
	}

	_, err := url.Parse(brokerConfig.Url)
	if err != nil {
		lc.Errorf("invalid MQTT Broker Url '%s': %s", brokerConfig.Url, err.Error())
		return false
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerConfig.Url)
	opts.SetClientID(brokerConfig.ClientId)
	opts.SetOnConnectHandler(e.onConnectHandler)
	opts.SetAutoReconnect(brokerConfig.AutoReconnect)
	opts.KeepAlive = brokerConfig.KeepAlive
	if len(brokerConfig.ConnectTimeout) > 0 {
		duration, err := time.ParseDuration(brokerConfig.ConnectTimeout)
		if err != nil {
			lc.Errorf("invalid MQTT ConnectTimeout '%s': %s", brokerConfig.ConnectTimeout, err.Error())
			return false
		}
		opts.SetConnectTimeout(duration)
	}

	secretProvider := container.SecretProviderFrom(dic.Get)
	authMode := brokerConfig.AuthMode
	if brokerConfig.AuthMode == "" {
		authMode = messaging.AuthModeNone
		lc.Warn("AuthMode not set, defaulting to \"" + messaging.AuthModeNone + "\"")
	}

	//get the secrets from the secret provider and populate the struct
	secretData, err := messaging.GetSecretData(authMode, brokerConfig.SecretName, secretProvider)
	if err != nil {
		lc.Errorf("Failed to retrieve secret data: %s", err.Error())
		return false
	}
	//ensure that the AuthMode selected has the required secret values
	if secretData != nil {
		err = messaging.ValidateSecretData(authMode, brokerConfig.SecretName, secretData)
		if err != nil {
			lc.Errorf("Invalid secret data: %s", err.Error())
			return false
		}

		// configure the mqtt client with the retrieved secret values
		tlsConfig := &tls.Config{
			// nolint: gosec
			InsecureSkipVerify: brokerConfig.SkipCertVerify,
			MinVersion:         tls.VersionTLS12,
		}
		switch authMode {
		case messaging.AuthModeUsernamePassword:
			opts.SetUsername(secretData.Username)
			opts.SetPassword(secretData.Password)
		case messaging.AuthModeCert:
			cert, err := tls.X509KeyPair(secretData.CertPemBlock, secretData.KeyPemBlock)
			if err != nil {
				lc.Errorf("Failed to parse public/private key pair: %s", err.Error())
				return false
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		case messaging.AuthModeCA:
			// Nothing to do here for this option
		case messaging.AuthModeNone:
			// Nothing to do here for this option
		}

		if len(secretData.CaPemBlock) > 0 {
			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM(secretData.CaPemBlock)
			if !ok {
				lc.Errorf("error parsing CA PEM block")
				return false
			}
			tlsConfig.RootCAs = caCertPool
		}

		opts.SetTLSConfig(tlsConfig)
	}

	var mqttClient mqtt.Client
	for startupTimer.HasNotElapsed() {
		select {
		case <-ctx.Done():
			return false
		default:
			mqttClient, err = createMqttClient(opts)
			if err != nil {
				lc.Warnf("Unable to create MQTT client: %s", err.Error())
				startupTimer.SleepForInterval()
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ctx.Done()
				mqttClient.Disconnect(0)
				lc.Info("Disconnected from external MQTT broker")
			}()

			dic.Update(di.ServiceConstructorMap{
				container.ExternalMQTTMessagingClientName: func(get di.Get) interface{} {
					return mqttClient
				},
			})

			lc.Infof(
				"Connected to external MQTT broker @ %s with AuthMode='%s'",
				brokerConfig.Url,
				brokerConfig.AuthMode)

			return true
		}
	}

	lc.Error("Connecting to external MQTT broker time out")
	return false
}

func createMqttClient(opts *mqtt.ClientOptions) (mqtt.Client, error) {
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("could not connect to broker: %s", token.Error().Error())
	}

	return mqttClient, nil
}
