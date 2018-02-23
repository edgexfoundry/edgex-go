//
// Copyright (c) 2017
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"crypto/tls"
	"strconv"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"go.uber.org/zap"
)

type mqttSender struct {
	client MQTT.Client
	topic  string
}

// NewMqttSender - create new mqtt sender
func NewMqttSender(addr models.Addressable) Sender {

	protocol := strings.ToLower(addr.Protocol)

	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	if protocol == "tcps" ||
		protocol == "ssl" ||
		protocol == "tls" {

		cert, err := tls.LoadX509KeyPair(cfg.MQTTSCert, cfg.MQTTSKey)

		if err != nil {
			logger.Error("Failed loading x509 data")
			return nil
		}

		tlsConfig := &tls.Config{
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		}

		opts.SetTLSConfig(tlsConfig)

	}

	sender := &mqttSender{
		client: MQTT.NewClient(opts),
		topic:  addr.Topic,
	}

	return sender
}

func (sender *mqttSender) Send(data []byte) {
	if !sender.client.IsConnected() {
		logger.Info("Connecting to mqtt server")
		if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
			logger.Warn("Could not connect to mqtt server, drop event", zap.Error(token.Error()))
			return
		}
	}

	token := sender.client.Publish(sender.topic, 0, false, data)
	// FIXME: could be removed? set of tokens?
	token.Wait()
	if token.Error() != nil {
		logger.Warn("mqtt error: ", zap.Error(token.Error()))
	} else {
		logger.Debug("Sent data: ", zap.ByteString("data", data))
	}
}
