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
	"github.com/edgexfoundry/go-mod-core-contracts/models"
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

// newIoTCoreSender returns new Google IoT Core sender instance.
func newIoTCoreSender(addr models.Addressable) sender {
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
		c := Configuration.Certificates["MQTTS"]
		cert, err := tls.LoadX509KeyPair(c.Cert, c.Key)
		if err != nil {
			LoggingClient.Error("Failed loading x509 data")
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

func (sender *iotCoreSender) Send(data []byte) bool {
	if !sender.client.IsConnected() {
		LoggingClient.Info("Connecting to mqtt server")
		token := sender.client.Connect()
		token.Wait()
		if token.Error() != nil {
			LoggingClient.Error(fmt.Sprintf("Could not connect to mqtt server, drop event. Error: %s", token.Error().Error()))
			return false
		}
	}

	token := sender.client.Publish(sender.topic, 0, false, data)
	token.Wait()
	if token.Error() != nil {
		LoggingClient.Error(token.Error().Error())
		return false
	}

	LoggingClient.Debug(fmt.Sprintf("Sent data: %X", data))
	return true
}

func extractDeviceID(addr string) string {
	return addr[strings.Index(addr, devicesPrefix)+len(devicesPrefix):]
}

func validateProtocol(protocol string) bool {
	return protocol == tcpsPrefix || protocol == sslPrefix || protocol == tlsPrefix
}
