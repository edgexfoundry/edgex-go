//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	zmq "github.com/pebbe/zmq4"

	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// Sender abstracts the notification sending via specified channel
type Sender interface {
	Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX)
}

// RESTSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via REST
type RESTSender struct {
	dic            *di.Container
	secretProvider bootstrapInterfaces.SecretProviderExt
}

// NewRESTSender creates the RESTSender instance
func NewRESTSender(dic *di.Container, secretProvider bootstrapInterfaces.SecretProviderExt) Sender {
	return &RESTSender{dic: dic, secretProvider: secretProvider}
}

// Send sends the REST request to the specified address
func (sender *RESTSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	lc := container.LoggingClientFrom(sender.dic.Get)

	restAddress, ok := address.(models.RESTAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to RESTAddress", nil)
	}

	var injector interfaces.AuthenticationInjector
	if restAddress.InjectEdgeXAuth {
		injector = secret.NewJWTSecretProvider(sender.secretProvider)
	}

	return utils.SendRequestWithRESTAddress(lc, notification.Content, notification.ContentType, restAddress, injector)
}

// EmailSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via email
type EmailSender struct {
	dic *di.Container
}

// NewEmailSender creates the EmailSender instance
func NewEmailSender(dic *di.Container) Sender {
	return &EmailSender{dic: dic}
}

// Send sends the email to the specified address
func (sender *EmailSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	smtpInfo := notificationContainer.ConfigurationFrom(sender.dic.Get).Smtp

	emailAddress, ok := address.(models.EmailAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to EmailAddress", nil)
	}

	msg := buildSmtpMessage(notification.Sender, smtpInfo.Subject, emailAddress.Recipients, notification.ContentType, notification.Content)
	auth, err := deduceAuth(sender.dic, smtpInfo)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	err = sendEmail(smtpInfo, auth, emailAddress.Recipients, msg)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	return "", nil
}

// MQTTSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via MQTT broker
type MQTTSender struct {
	dic   *di.Container
	ctx   context.Context
	wg    *sync.WaitGroup
	mutex sync.RWMutex
	// clientCache stores the MQTT client for reusing
	clientCache map[string]mqtt.Client
}

// NewMQTTSender creates the MQTTSender instance
func NewMQTTSender(ctx context.Context, wg *sync.WaitGroup, dic *di.Container) Sender {
	sender := &MQTTSender{ctx: ctx, wg: wg, dic: dic, clientCache: make(map[string]mqtt.Client)}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		lc := container.LoggingClientFrom(dic.Get)
		for key, mqttClient := range sender.clientCache {
			if mqttClient.IsConnected() {
				mqttClient.Disconnect(uint(WaitDuration))
				lc.Infof("disconnected from the MQTT broker(%s)", key)
			}
		}
	}()
	return sender
}

// Send sends the message to the MQTT broker
func (sender *MQTTSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	mqttAddress, ok := address.(models.MQTTPubAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to MQTTPubAddress", nil)
	}

	client, err := sender.prepareMqttClient(mqttAddress)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	payload, _ := json.Marshal(notification)
	token := client.Publish(mqttAddress.Topic, byte(mqttAddress.QoS), mqttAddress.Retained, payload)
	if token.WaitTimeout(WaitDuration) && token.Error() != nil {
		return "", errors.NewCommonEdgeXWrapper(token.Error())
	}
	return "", nil
}

// ZeroMQSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via ZeroMQ
type ZeroMQSender struct {
	dic   *di.Container
	ctx   context.Context
	wg    *sync.WaitGroup
	mutex sync.RWMutex
	// clientCache stores the ZeroMQ client for reusing
	clientCache map[string]*zmq.Socket
}

// NewZeroMQSender creates the ZeroMQSender instance
func NewZeroMQSender(ctx context.Context, wg *sync.WaitGroup, dic *di.Container) Sender {
	sender := &ZeroMQSender{ctx: ctx, wg: wg, dic: dic, clientCache: make(map[string]*zmq.Socket)}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		lc := container.LoggingClientFrom(dic.Get)
		for key, socket := range sender.clientCache {
			err := socket.SetLinger(time.Duration(0))
			if err != nil {
				lc.Errorf("fail to set linger period for ZeroMQ socket shutdown, %v", err)
			}
			err = socket.Close()
			if err != nil {
				lc.Errorf("unable to close ZeroMQ socket, %v", err)
			}
			lc.Infof("disconnected from the ZeroMQ socket (port: %d)", key)
		}
	}()
	return sender
}

// Send sends the message to the ZeroMQ
func (sender *ZeroMQSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	zeroMQAddress, ok := address.(models.ZeroMQAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to ZeroMQAddress", nil)
	}

	socket, err := sender.prepareZeroMQTClient(zeroMQAddress)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	_, zmqErr := socket.SendMessage(zeroMQAddress.Topic, notification.Content)
	if zmqErr != nil {
		return "", errors.NewCommonEdgeXWrapper(zmqErr)
	}

	return "", nil
}

// RemoveClientFromCache removes the client from the Sender's clientCache
func RemoveClientFromCache(dic *di.Container, addresses []models.Address) errors.EdgeX {
	mqttSender := MQTTSenderFrom(dic.Get).(*MQTTSender)
	zmqSender := ZeroMQSenderFrom(dic.Get).(*ZeroMQSender)

	for _, address := range addresses {
		switch address.GetBaseAddress().Type {
		case common.MQTT:
			mqttAddress, ok := address.(models.MQTTPubAddress)
			if !ok {
				return errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to MQTTPubAddress", nil)
			}
			mqttSender.removeClientFromCache(mqttAddress)
		case common.ZeroMQ:
			zeroMQAddress, ok := address.(models.ZeroMQAddress)
			if !ok {
				return errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to ZeroMQAddress", nil)
			}
			zmqSender.removeClientFromCache(zeroMQAddress)
		}
	}
	return nil
}

func (sender *MQTTSender) removeClientFromCache(address models.MQTTPubAddress) {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	key := sender.cacheKey(address.Publisher, address.Host, address.Port)
	mqttClient, ok := sender.clientCache[key]
	if ok {
		if mqttClient.IsConnected() {
			mqttClient.Disconnect(uint(WaitDuration))
			delete(sender.clientCache, key)
		}
	}
}

func (sender *ZeroMQSender) removeClientFromCache(address models.ZeroMQAddress) {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	lc := container.LoggingClientFrom(sender.dic.Get)
	key := strconv.Itoa(address.Port)
	socket, ok := sender.clientCache[key]
	if ok {
		err := socket.SetLinger(time.Duration(0))
		if err != nil {
			lc.Errorf("fail to set linger period for ZeroMQ socket shutdown, %v", err)
		}
		err = socket.Close()
		if err != nil {
			lc.Errorf("unable to close ZeroMQ socket, %v", err)
		}
		delete(sender.clientCache, key)
	}
}
