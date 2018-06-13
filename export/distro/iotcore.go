//
// Copyright (c) 2017
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"crypto/tls"
	"fmt"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"go.uber.org/zap"
)

const (
	tcpsPrefix    = "tcps"
	sslPrefix     = "ssl"
	tlsPrefix     = "tls"
	devicesPrefix = "/devices/"
)

type iotCoreSender struct {
	client MQTT.Client
	topic  string
}

// NewIoTCoreSender returns new Google IoT Core sender instance.
func NewIoTCoreSender(addr models.Addressable) Sender {
	protocol := strings.ToLower(addr.Protocol)
	broker := fmt.Sprintf("%s%s", addr.GetBaseURL(), addr.Path)
	deviceID := extractDeviceID(addr.Publisher)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	if validateProtocol(protocol) {
		cert, err := tls.LoadX509KeyPair(configuration.MQTTSCert, configuration.MQTTSKey)
		if err != nil {
			logger.Error("Failed loading x509 data")
			return nil
		}

		opts.SetTLSConfig(&tls.Config{
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		})
	}

	if addr.Topic == "" {
		addr.Topic = fmt.Sprintf("/devices/%s/events", deviceID)
	}

	return &mqttSender{
		client: MQTT.NewClient(opts),
		topic:  addr.Topic,
	}
}

func (sender *iotCoreSender) Send(data []byte) {
	if !sender.client.IsConnected() {
		logger.Info("Connecting to mqtt server")
		token := sender.client.Connect()
		token.Wait()
		if token.Error() != nil {
			logger.Warn("Could not connect to mqtt server, drop event", zap.Error(token.Error()))
			return
		}
	}

	token := sender.client.Publish(sender.topic, 0, false, data)
	token.Wait()
	if token.Error() != nil {
		logger.Warn("mqtt error: ", zap.Error(token.Error()))
		return
	}

	logger.Debug("Sent data: ", zap.ByteString("data", data))
}

func extractDeviceID(addr string) string {
	return addr[strings.Index(addr, devicesPrefix)+len(devicesPrefix):]
}

func validateProtocol(protocol string) bool {
	return protocol == tcpsPrefix || protocol == sslPrefix || protocol == tlsPrefix
}
