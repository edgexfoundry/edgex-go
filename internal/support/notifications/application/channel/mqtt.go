// Copyright (C) 2024-2025 IOTech Ltd

package channel

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"strings"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	WaitDuration = 3 * time.Second
)

// prepareMqttClient creates a new client or load the exist client from cache
func (sender *MQTTSender) prepareMqttClient(address models.MQTTPubAddress) (mqtt.Client, errors.EdgeX) {
	client := sender.loadClient(address)
	if client != nil {
		return client, nil
	}

	client, err := sender.createClient(address)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	return client, nil
}

func (sender *MQTTSender) cacheKey(publisher string, host string, port int) string {
	return fmt.Sprintf("%s:%s:%d", publisher, host, port)
}

func (sender *MQTTSender) loadClient(address models.MQTTPubAddress) mqtt.Client {
	sender.mutex.RLock()
	defer sender.mutex.RUnlock()
	key := sender.cacheKey(address.Publisher, address.Host, address.Port)
	mqttClient, ok := sender.clientCache[key]
	if ok {
		return mqttClient
	}
	return nil
}

// createMqttClient creates a new MQTT client
// The implementation can refer to https://github.com/edgexfoundry/app-functions-sdk-go/blob/1bc0c5a6f3d13f883f4b71f940f0cb2168d0daab/pkg/secure/mqttfactory.go#L58
func (sender *MQTTSender) createClient(address models.MQTTPubAddress) (mqtt.Client, errors.EdgeX) {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	// Check the cache before creating new MQTT client
	key := sender.cacheKey(address.Publisher, address.Host, address.Port)
	mqttClient, ok := sender.clientCache[key]
	if ok {
		return mqttClient, nil
	}

	scheme := common.TCP
	if address.Scheme != "" {
		scheme = address.Scheme
	}
	brokerUrl := &url.URL{
		Scheme: strings.ToLower(scheme),
		Host:   fmt.Sprintf("%s:%d", address.Host, address.Port),
	}
	opts := mqtt.NewClientOptions()
	opts.SetAutoReconnect(true)
	opts.SetClientID(address.Publisher)
	opts.Servers = []*url.URL{brokerUrl}

	secretProvider := bootstrapContainer.SecretProviderFrom(sender.dic.Get)

	//get the secrets from the secret provider and populate the struct
	secretData, err := messaging.GetSecretData(address.AuthMode, address.SecretPath, secretProvider)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	//ensure that the authmode selected has the required secret values
	if secretData != nil {
		err = messaging.ValidateSecretData(address.AuthMode, address.SecretPath, secretData)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		// configure the mqtt client with the retrieved secret values
		err = configureMQTTClientForAuth(address, opts, secretData)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(WaitDuration) && token.Error() != nil {
		return client, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to connect the MQTT broker, %v", token.Error()), nil)
	}

	sender.clientCache[key] = client

	return client, nil
}

func configureMQTTClientForAuth(address models.MQTTPubAddress, options *mqtt.ClientOptions, secretData *messaging.SecretData) errors.EdgeX {
	caCertPool := x509.NewCertPool()
	tlsConfig := &tls.Config{
		// nolint: gosec
		InsecureSkipVerify: address.SkipCertVerify,
	}

	switch address.AuthMode {
	case messaging.AuthModeUsernamePassword:
		options.SetUsername(secretData.Username)
		options.SetPassword(secretData.Password)
	case messaging.AuthModeCert:
		// Expect user set require_certificate and use_identity_as_username to true, which is assumed that only authenticated clients have valid certificates
		// This authentication usage can refer to https://mosquitto.org/man/mosquitto-conf-5.html
		cert, err := tls.X509KeyPair(secretData.CertPemBlock, secretData.KeyPemBlock)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	case messaging.AuthModeCA:
		// Nothing to do here for this option
	case messaging.AuthModeNone:
		return nil
	}

	if len(secretData.CaPemBlock) > 0 {
		ok := caCertPool.AppendCertsFromPEM(secretData.CaPemBlock)
		if !ok {
			return errors.NewCommonEdgeX(errors.KindServerError, "Error parsing CA PEM block", nil)
		}
		tlsConfig.RootCAs = caCertPool
	}

	options.SetTLSConfig(tlsConfig)

	return nil
}
