//
// Copyright (c) 2018 IOTech
//

package distro

import (
	"crypto/tls"
	"fmt"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/edgex-go/internal/export/interfaces"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"go.uber.org/zap"
)

const (
	awsMQTTPort         int    = 8883
	awsThingUpdateTopic string = "$aws/things/%s/shadow/update"
)

type awsiotSender struct {
	client MQTT.Client
	topic  string
}

func NewAWSIoTSender(addr models.Addressable) interfaces.Sender {

	cert, err := tls.LoadX509KeyPair(configuration.AWSCert, configuration.AWSKey)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to load x509 certificate or key: %s", err))
		return nil
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConfig.BuildNameToCertificate()

	serverURL := "tls://" + addr.Address + ":" + strconv.Itoa(awsMQTTPort)

	topic := fmt.Sprintf(awsThingUpdateTopic, addr.Topic)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(serverURL)
	opts.SetClientID(addr.Publisher).SetTLSConfig(tlsConfig)

	client := MQTT.NewClient(opts)

	sender := &awsiotSender{
		client: client,
		topic:  topic,
	}

	return sender
}

func (sender *awsiotSender) Send(data []byte, event *models.Event) bool {
	client := sender.client
	topic := sender.topic

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Error("Failed to connect to thing: %v", zap.Error(token.Error()))
		return false
	}

	if token := client.Publish(topic, 0, false, data); token.Wait() && token.Error() != nil {
		logger.Error("Failed to publish thing state to update topic: %v", zap.Error(token.Error()))
		return false
	}

	return true
}
