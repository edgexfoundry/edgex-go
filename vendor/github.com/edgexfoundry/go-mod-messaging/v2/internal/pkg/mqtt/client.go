/********************************************************************************
 *  Copyright 2019 Dell Inc.
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
 *******************************************************************************/

package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
)

// ClientCreator defines the function signature for creating an MQTT client.
type ClientCreator func(config types.MessageBusConfig, handler pahoMqtt.OnConnectHandler) (pahoMqtt.Client, error)

// MessageMarshaller defines the function signature for marshaling structs into []byte.
type MessageMarshaller func(v interface{}) ([]byte, error)

// MessageUnmarshaller defines the function signature for unmarshaling []byte into structs.
type MessageUnmarshaller func(data []byte, v interface{}) error

// Client facilitates communication to an MQTT server and provides functionality needed to send and receive MQTT
// messages.
type Client struct {
	creator             ClientCreator
	configuration       types.MessageBusConfig
	mqttClient          pahoMqtt.Client
	marshaller          MessageMarshaller
	unmarshaller        MessageUnmarshaller
	activeSubscriptions []activeSubscription
}

type activeSubscription struct {
	topic   string
	qos     byte
	handler pahoMqtt.MessageHandler
	errors  chan error
}

// NewMQTTClient constructs a new MQTT client based on the options provided.
func NewMQTTClient(config types.MessageBusConfig) (*Client, error) {
	client := &Client{
		creator:       DefaultClientCreator(),
		configuration: config,
		marshaller:    json.Marshal,
		unmarshaller:  json.Unmarshal,
	}

	return client, nil
}

// NewMQTTClientWithCreator constructs a new MQTT client based on the options and ClientCreator provided.
func NewMQTTClientWithCreator(
	config types.MessageBusConfig,
	marshaller MessageMarshaller,
	unmarshaller MessageUnmarshaller,
	creator ClientCreator) (*Client, error) {

	client := &Client{
		creator:       creator,
		configuration: config,
		marshaller:    marshaller,
		unmarshaller:  unmarshaller,
	}

	return client, nil
}

// Connect establishes a connection to a MQTT server.
// This must be called before any other functionality provided by the Client.
func (mc *Client) Connect() error {
	if mc.mqttClient == nil {
		// Move created MQTT Client here since we need to set the onConnectHandler which needs to have access to
		// the Client's activeSubscriptions. This was not possible from the factory method.
		mqttClient, err := mc.creator(mc.configuration, mc.onConnectHandler)
		if err != nil {
			return err
		}
		mc.mqttClient = mqttClient
	}

	// Avoid reconnecting if already connected.
	if mc.mqttClient.IsConnected() {
		return nil
	}

	optionsReader := mc.mqttClient.OptionsReader()

	return getTokenError(
		mc.mqttClient.Connect(),
		optionsReader.ConnectTimeout(),
		ConnectOperation,
		"Unable to connect")
}

func (mc *Client) onConnectHandler(_ pahoMqtt.Client) {
	optionsReader := mc.mqttClient.OptionsReader()

	// activeSubscriptions will be empty on the first connection.
	// On a re-connect is when the subscriptions must be re-created.
	for _, subscription := range mc.activeSubscriptions {
		token := mc.mqttClient.Subscribe(subscription.topic, subscription.qos, subscription.handler)
		message := fmt.Sprintf("Failed to re-create subscription for topic=%s", subscription.topic)
		err := getTokenError(token, optionsReader.ConnectTimeout(), SubscribeOperation, message)
		if err != nil {
			subscription.errors <- err
		}
	}
}

// Publish sends a message to the connected MQTT server.
func (mc *Client) Publish(message types.MessageEnvelope, topic string) error {
	marshaledMessage, err := mc.marshaller(message)
	if err != nil {
		return NewOperationErr(PublishOperation, err.Error())
	}

	optionsReader := mc.mqttClient.OptionsReader()

	return getTokenError(
		mc.mqttClient.Publish(
			topic,
			optionsReader.WillQos(),
			optionsReader.WillRetained(),
			marshaledMessage),
		optionsReader.ConnectTimeout(),
		PublishOperation,
		"Unable to publish message")

}

// Subscribe creates a subscription for the specified topics.
func (mc *Client) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	optionsReader := mc.mqttClient.OptionsReader()

	for _, topic := range topics {
		handler := newMessageHandler(mc.unmarshaller, topic.Messages, messageErrors)
		qos := optionsReader.WillQos()

		token := mc.mqttClient.Subscribe(topic.Topic, qos, handler)
		err := getTokenError(token, optionsReader.ConnectTimeout(), SubscribeOperation, "Failed to create subscription")
		if err != nil {
			return err
		}

		mc.activeSubscriptions = append(mc.activeSubscriptions, activeSubscription{
			topic:   topic.Topic,
			qos:     qos,
			handler: handler,
			errors:  messageErrors,
		})

	}

	return nil
}

// Disconnect closes the connection to the connected MQTT server.
func (mc *Client) Disconnect() error {
	// Specify a wait time equal to the write timeout so that we allow other any queued processing to complete before
	// disconnecting.
	optionsReader := mc.mqttClient.OptionsReader()
	mc.mqttClient.Disconnect(uint(optionsReader.ConnectTimeout() * time.Millisecond))

	return nil
}

// DefaultClientCreator returns a default function for creating MQTT clients.
func DefaultClientCreator() ClientCreator {
	return func(config types.MessageBusConfig, handler pahoMqtt.OnConnectHandler) (pahoMqtt.Client, error) {
		clientConfiguration, err := CreateMQTTClientConfiguration(config)
		if err != nil {
			return nil, err
		}

		clientOptions, err := createClientOptions(clientConfiguration, tls.X509KeyPair, tls.LoadX509KeyPair)
		if err != nil {
			return nil, err
		}

		clientOptions.OnConnect = handler
		return pahoMqtt.NewClient(clientOptions), nil
	}
}

// ClientCreatorWithCertLoader creates a ClientCreator which leverages the specified cert creator and loader when
// creating an MQTT client.
func ClientCreatorWithCertLoader(certCreator pkg.X509KeyPairCreator, certLoader pkg.X509KeyLoader) ClientCreator {
	return func(options types.MessageBusConfig, handler pahoMqtt.OnConnectHandler) (pahoMqtt.Client, error) {
		clientConfiguration, err := CreateMQTTClientConfiguration(options)
		if err != nil {
			return nil, err
		}

		clientOptions, err := createClientOptions(clientConfiguration, certCreator, certLoader)
		if err != nil {
			return nil, err
		}

		clientOptions.OnConnect = handler
		return pahoMqtt.NewClient(clientOptions), nil
	}
}

// newMessageHandler creates a function which meets the criteria for a MessageHandler and propagates the received
// messages to the proper channel.
func newMessageHandler(
	unmarshaler MessageUnmarshaller,
	messageChannel chan<- types.MessageEnvelope,
	errorChannel chan<- error) pahoMqtt.MessageHandler {

	return func(client pahoMqtt.Client, message pahoMqtt.Message) {
		var messageEnvelope types.MessageEnvelope
		payload := message.Payload()
		err := unmarshaler(payload, &messageEnvelope)
		if err != nil {
			errorChannel <- err
		}

		messageEnvelope.ReceivedTopic = message.Topic()

		messageChannel <- messageEnvelope
	}
}

// getTokenError determines if a Token is in an errored state and if so returns the proper error message. Otherwise,
// nil.
//
// NOTE the paho.pahoMqtt.golang's recommended way for handling errors do not cover all cases. During manual verification
// with an MQTT server, it was observed that the Token.Error() was sometimes nil even when a token.WaitTimeout(...)
// returned false(indicating the operation has timed-out). Therefore, there are some additional checks that need to
// take place to ensure the error message is returned if it is present. One example scenario, if you attempt to connect
// without providing a ClientID.
func getTokenError(token pahoMqtt.Token, timeout time.Duration, operation string, defaultTimeoutMessage string) error {
	hasTimedOut := !token.WaitTimeout(timeout)

	if hasTimedOut && token.Error() != nil {
		return NewTimeoutError(operation, token.Error().Error())
	}

	if hasTimedOut && token.Error() == nil {
		return NewTimeoutError(operation, defaultTimeoutMessage)
	}

	if token.Error() != nil {
		return NewOperationErr(operation, token.Error().Error())
	}

	return nil
}

// createClientOptions constructs mqtt.Client options from an MQTTClientConfig.
func createClientOptions(
	clientConfiguration MQTTClientConfig,
	certCreator pkg.X509KeyPairCreator,
	certLoader pkg.X509KeyLoader) (*pahoMqtt.ClientOptions, error) {

	clientOptions := pahoMqtt.NewClientOptions()
	clientOptions.AddBroker(clientConfiguration.BrokerURL)
	clientOptions.SetUsername(clientConfiguration.Username)
	clientOptions.SetPassword(clientConfiguration.Password)
	clientOptions.SetClientID(clientConfiguration.ClientId)
	clientOptions.SetKeepAlive(time.Duration(clientConfiguration.KeepAlive) * time.Second)
	clientOptions.WillQos = byte(clientConfiguration.Qos)
	clientOptions.WillRetained = clientConfiguration.Retained
	clientOptions.CleanSession = clientConfiguration.CleanSession
	clientOptions.SetAutoReconnect(clientConfiguration.AutoReconnect)
	clientOptions.SetConnectTimeout(time.Duration(clientConfiguration.ConnectTimeout) * time.Second)
	tlsConfiguration, err := pkg.GenerateTLSForClientClientOptions(
		clientConfiguration.BrokerURL,
		clientConfiguration.TlsConfigurationOptions,
		certCreator,
		certLoader)

	if err != nil {
		return clientOptions, err
	}

	clientOptions.SetTLSConfig(tlsConfiguration)

	return clientOptions, nil
}
