//
// Copyright (C) 2022-2025 IOTech Ltd
// Copyright (C) 2023 Intel Inc.
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

// Defaults for ExternalCommandQueue when configuration omits zero values.
const (
	defaultMaxConcurrentExternalCommands  = 32
	defaultMaxQueuedExternalCommands      = 64
	defaultOverloadPublishChannelCapacity = 16
	defaultShutdownTimeoutString          = "30s"
)

var defaultExternalMQTTShutdownTimeout time.Duration

func init() {
	var err error
	defaultExternalMQTTShutdownTimeout, err = time.ParseDuration(defaultShutdownTimeoutString)
	if err != nil {
		defaultExternalMQTTShutdownTimeout = 30 * time.Second
	}
}

type externalCommandLimits struct {
	maxWorkers     int
	maxQueued      int
	errPubCapacity int
	shutdown       time.Duration
}

func externalCommandLimitsFromConfig(q config.ExternalCommandQueue) externalCommandLimits {
	w := q.MaxConcurrentExternalCommands
	if w <= 0 {
		w = defaultMaxConcurrentExternalCommands
	}
	queueCap := q.MaxQueuedExternalCommands
	if queueCap <= 0 {
		queueCap = defaultMaxQueuedExternalCommands
	}
	e := q.OverloadPublishChannelCapacity
	if e <= 0 {
		e = defaultOverloadPublishChannelCapacity
	}
	st := q.ShutdownTimeout
	if st == "" {
		st = defaultShutdownTimeoutString
	}
	d, err := time.ParseDuration(st)
	if err != nil || d <= 0 {
		d = defaultExternalMQTTShutdownTimeout
	}
	return externalCommandLimits{
		maxWorkers:     w,
		maxQueued:      queueCap,
		errPubCapacity: e,
		shutdown:       d,
	}
}

// externalCommandWork is dispatched to workers; payload is not used after enqueue (envelope carries content).
type externalCommandWork struct {
	envelope              types.MessageEnvelope
	mqttTopic             string
	externalResponseTopic string
}

type overloadPublishWork struct {
	topic   string
	payload []byte
	qos     byte
	retain  bool
}

// Blocking enqueue on a full jobs channel was considered and rejected: it would only move head-of-line
// blocking from internal Request to queue admission and can stall other MQTT handlers on the same client
// (see design doc / plan §2). Only non-blocking send with overload response is supported.

// externalCommandProcessor owns the worker pool and overload publisher for external MQTT commands.
type externalCommandProcessor struct {
	ctx            context.Context
	requestTimeout time.Duration
	dic            *di.Container
	limits         externalCommandLimits

	clientMu sync.RWMutex
	client   mqtt.Client

	jobs       chan externalCommandWork
	errPubCh   chan overloadPublishWork
	enqueueMu  sync.Mutex
	jobsClosed bool

	workerWg      sync.WaitGroup
	overloadPubWg sync.WaitGroup
	startOnce     sync.Once
}

func newExternalCommandProcessor(ctx context.Context, requestTimeout time.Duration, dic *di.Container, limits externalCommandLimits) *externalCommandProcessor {
	return &externalCommandProcessor{
		ctx:            ctx,
		requestTimeout: requestTimeout,
		dic:            dic,
		limits:         limits,
	}
}

func (p *externalCommandProcessor) setMQTTClient(c mqtt.Client) {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()
	p.client = c
}

func (p *externalCommandProcessor) getMQTTClient() mqtt.Client {
	p.clientMu.RLock()
	defer p.clientMu.RUnlock()
	return p.client
}

func (p *externalCommandProcessor) ensureStarted() {
	p.startOnce.Do(func() {
		p.jobs = make(chan externalCommandWork, p.limits.maxQueued)
		p.errPubCh = make(chan overloadPublishWork, p.limits.errPubCapacity)
		p.workerWg.Add(p.limits.maxWorkers)
		for i := 0; i < p.limits.maxWorkers; i++ {
			go p.workerLoop()
		}
		p.overloadPubWg.Add(1)
		go p.overloadPublisherLoop()
		go p.runShutdownWatcher()
	})
}

func (p *externalCommandProcessor) runShutdownWatcher() {
	<-p.ctx.Done()
	done := make(chan struct{})
	go func() {
		p.shutdownWorkerPool()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(p.limits.shutdown):
		lc := bootstrapContainer.LoggingClientFrom(p.dic.Get)
		lc.Warnf("external MQTT command worker shutdown exceeded ShutdownTimeout (%v)", p.limits.shutdown)
	}
}

func (p *externalCommandProcessor) shutdownWorkerPool() {
	p.enqueueMu.Lock()
	if p.jobsClosed {
		p.enqueueMu.Unlock()
		return
	}
	p.jobsClosed = true
	close(p.jobs)
	p.enqueueMu.Unlock()

	p.workerWg.Wait()
	close(p.errPubCh)
	p.overloadPubWg.Wait()
}

func (p *externalCommandProcessor) tryEnqueue(w externalCommandWork) bool {
	p.enqueueMu.Lock()
	defer p.enqueueMu.Unlock()
	if p.jobsClosed {
		return false
	}
	// Non-blocking enqueue only: blocking send would reintroduce head-of-line blocking on the Paho callback
	// when the buffer is full (see package comment near commandRequestMQTTHandler).
	select {
	case p.jobs <- w:
		return true
	default:
		return false
	}
}

func (p *externalCommandProcessor) workerLoop() {
	defer p.workerWg.Done()
	lc := bootstrapContainer.LoggingClientFrom(p.dic.Get)
	for w := range p.jobs {
		func(work externalCommandWork) {
			defer func() {
				if r := recover(); r != nil {
					lc.Errorf("recovered panic in external MQTT command worker: %v", r)
				}
			}()
			p.processExternalMQTTCommand(work)
		}(w)
	}
}

func (p *externalCommandProcessor) overloadPublisherLoop() {
	defer p.overloadPubWg.Done()
	lc := bootstrapContainer.LoggingClientFrom(p.dic.Get)
	for item := range p.errPubCh {
		client := p.getMQTTClient()
		if client == nil {
			continue
		}
		token := client.Publish(item.topic, item.qos, item.retain, item.payload)
		if token.Wait() && token.Error() != nil {
			lc.Errorf("Could not publish overload response to external message broker on topic '%s': %s",
				item.topic, token.Error().Error())
		}
	}
}

func (p *externalCommandProcessor) enqueueOverloadPublish(externalResponseTopic string, qos byte, retain bool, envelope types.MessageEnvelope, lc logger.LoggingClient) {
	payload, err := json.Marshal(&envelope)
	if err != nil {
		lc.Errorf("Failed to marshal overload MessageEnvelope: %s", err.Error())
		return
	}
	item := overloadPublishWork{topic: externalResponseTopic, payload: payload, qos: qos, retain: retain}
	select {
	case p.errPubCh <- item:
	default:
		lc.Warn("overload response channel full; dropping overload notification for external MQTT command load shed")
	}
}

// OnConnectHandler returns an MQTT OnConnectHandler that subscribes to external command and query topics.
func OnConnectHandler(ctx context.Context, requestTimeout time.Duration, dic *di.Container) mqtt.OnConnectHandler {
	cfg := container.ConfigurationFrom(dic.Get)
	limits := externalCommandLimitsFromConfig(cfg.ExternalCommandQueue)
	proc := newExternalCommandProcessor(ctx, requestTimeout, dic, limits)

	return func(client mqtt.Client) {
		proc.setMQTTClient(client)
		proc.ensureStarted()

		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		config := container.ConfigurationFrom(dic.Get)
		externalTopics := config.ExternalMQTT.Topics
		qos := config.ExternalMQTT.QoS

		requestQueryTopic := externalTopics[common.CommandQueryRequestTopicKey]
		if token := client.Subscribe(requestQueryTopic, qos, commandQueryHandler(dic)); token.Wait() && token.Error() != nil {
			lc.Errorf("could not subscribe to topic '%s': %s", requestQueryTopic, token.Error().Error())
		} else {
			lc.Debugf("Subscribed to topic '%s' on external MQTT broker", requestQueryTopic)
		}

		requestCommandTopic := externalTopics[common.CommandRequestTopicKey]
		if token := client.Subscribe(requestCommandTopic, qos, proc.commandRequestMQTTHandler()); token.Wait() && token.Error() != nil {
			lc.Errorf("could not subscribe to topic '%s': %s", requestCommandTopic, token.Error().Error())
		} else {
			lc.Debugf("Subscribed to topic '%s' on external MQTT broker", requestCommandTopic)
		}
	}
}

func commandQueryHandler(dic *di.Container) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		lc.Debugf("Received command query request from external message broker on topic '%s' with %d bytes", message.Topic(), len(message.Payload()))

		requestEnvelope, err := types.NewMessageEnvelopeFromJSON(message.Payload())
		if err != nil {
			lc.Errorf("Failed to decode request MessageEnvelope: %s", err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		externalMQTTInfo := container.ConfigurationFrom(dic.Get).ExternalMQTT
		responseTopic := externalMQTTInfo.Topics[common.CommandQueryResponseTopicKey]
		if responseTopic == "" {
			lc.Error("QueryResponseTopic not provided in External.Topics")
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		// example topic scheme: edgex/commandquery/request/<device-name>
		// deviceName is expected to be at last topic level.
		topicLevels := strings.Split(message.Topic(), "/")
		deviceName, err := url.PathUnescape(topicLevels[len(topicLevels)-1])
		if err != nil {
			lc.Errorf("Failed to unescape device name '%s': %s", deviceName, err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}
		if strings.EqualFold(deviceName, common.All) {
			deviceName = common.All
		}

		responseEnvelope, err := getCommandQueryResponseEnvelope(requestEnvelope, deviceName, dic)
		if err != nil {
			responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		}

		qos := externalMQTTInfo.QoS
		retain := externalMQTTInfo.Retain
		responseEnvelope.ReceivedTopic = responseTopic
		publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
	}
}

func (p *externalCommandProcessor) commandRequestMQTTHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		lc := bootstrapContainer.LoggingClientFrom(p.dic.Get)
		cfg := container.ConfigurationFrom(p.dic.Get)
		lc.Debugf("Received command request from external message broker on topic '%s' with %d bytes", message.Topic(), len(message.Payload()))

		externalMQTTInfo := cfg.ExternalMQTT
		qos := externalMQTTInfo.QoS
		retain := externalMQTTInfo.Retain

		payload := append([]byte(nil), message.Payload()...)
		mqttTopic := strings.Clone(message.Topic())

		requestEnvelope, err := types.NewMessageEnvelopeFromJSON(payload)
		if err != nil {
			lc.Errorf("Failed to decode request MessageEnvelope: %s", err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		topicLevels := strings.Split(mqttTopic, "/")
		length := len(topicLevels)
		if length < 3 {
			lc.Error("Failed to parse and construct response topic scheme, expected request topic scheme: '#/<device-name>/<command-name>/<method>")
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		deviceName, err := url.PathUnescape(topicLevels[length-3])
		if err != nil {
			lc.Errorf("Failed to unescape device name from '%s': %s", topicLevels[length-3], err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}
		commandName, err := url.PathUnescape(topicLevels[length-2])
		if err != nil {
			lc.Errorf("Failed to unescape command name from '%s': %s", topicLevels[length-2], err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}
		method := topicLevels[length-1]
		if !strings.EqualFold(method, "get") && !strings.EqualFold(method, "set") {
			lc.Errorf("Unknown request method: %s, only 'get' or 'set' is allowed", method)
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		externalResponseTopic := common.BuildTopic(externalMQTTInfo.Topics[common.CommandResponseTopicPrefixKey], deviceName, commandName, method)

		_, err = retrieveServiceNameByDevice(deviceName, p.dic)
		if err != nil {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		err = validateGetCommandQueryParameters(requestEnvelope.QueryParams)
		if err != nil {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		select {
		case <-p.ctx.Done():
			return
		default:
		}

		work := externalCommandWork{
			envelope:              requestEnvelope,
			mqttTopic:             mqttTopic,
			externalResponseTopic: externalResponseTopic,
		}
		if p.tryEnqueue(work) {
			return
		}

		// Service busy — non-blocking policy only; blocking enqueue would reintroduce callback HOL (see tryEnqueue).
		busyMsg := "service busy: external command queue is full"
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, busyMsg)
		p.enqueueOverloadPublish(externalResponseTopic, qos, retain, responseEnvelope, lc)
	}
}

func (p *externalCommandProcessor) processExternalMQTTCommand(work externalCommandWork) {
	lc := bootstrapContainer.LoggingClientFrom(p.dic.Get)
	cfg := container.ConfigurationFrom(p.dic.Get)
	externalMQTTInfo := cfg.ExternalMQTT
	qos := externalMQTTInfo.QoS
	retain := externalMQTTInfo.Retain
	requestEnvelope := work.envelope
	externalResponseTopic := work.externalResponseTopic

	topicLevels := strings.Split(work.mqttTopic, "/")
	length := len(topicLevels)
	deviceName, err := url.PathUnescape(topicLevels[length-3])
	if err != nil {
		lc.Errorf("Failed to unescape device name from '%s': %s", topicLevels[length-3], err.Error())
		return
	}
	commandName, err := url.PathUnescape(topicLevels[length-2])
	if err != nil {
		lc.Errorf("Failed to unescape command name from '%s': %s", topicLevels[length-2], err.Error())
		return
	}
	method := topicLevels[length-1]

	internalBaseTopic := cfg.MessageBus.GetBaseTopicPrefix()
	topicPrefix := common.BuildTopic(internalBaseTopic, common.CoreCommandDeviceRequestPublishTopic)

	deviceServiceName, err := retrieveServiceNameByDevice(deviceName, p.dic)
	if err != nil {
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		publishMessage(p.getMQTTClient(), externalResponseTopic, qos, retain, responseEnvelope, lc)
		return
	}

	err = validateGetCommandQueryParameters(requestEnvelope.QueryParams)
	if err != nil {
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		publishMessage(p.getMQTTClient(), externalResponseTopic, qos, retain, responseEnvelope, lc)
		return
	}

	deviceRequestTopic := common.NewPathBuilder().EnableNameFieldEscape(cfg.Service.EnableNameFieldEscape).
		SetPath(topicPrefix).SetNameFieldPath(deviceServiceName).SetNameFieldPath(deviceName).SetNameFieldPath(commandName).SetPath(method).BuildPath()
	deviceResponseTopicPrefix := common.NewPathBuilder().EnableNameFieldEscape(cfg.Service.EnableNameFieldEscape).
		SetPath(internalBaseTopic).SetPath(common.ResponseTopic).SetNameFieldPath(deviceServiceName).BuildPath()

	lc.Debugf("Sending Command request to internal MessageBus. Topic: %s, Request-id: %s Correlation-id: %s", deviceRequestTopic, requestEnvelope.RequestID, requestEnvelope.CorrelationID)
	lc.Debugf("Expecting response on topic: %s/%s", deviceResponseTopicPrefix, requestEnvelope.RequestID)

	internalMessageBus := bootstrapContainer.MessagingClientFrom(p.dic.Get)
	// go-mod-messaging/v4 MessageClient.Request has no context parameter; cancellation is only via timeout.
	// Service shutdown closes the jobs channel so workers exit without relying on context cancellation here.
	response, err := internalMessageBus.Request(requestEnvelope, deviceRequestTopic, deviceResponseTopicPrefix, p.requestTimeout)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to send DeviceCommand request with internal MessageBus: %v", err)
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
		publishMessage(p.getMQTTClient(), externalResponseTopic, qos, retain, responseEnvelope, lc)
		return
	}

	lc.Debugf("Command response received from internal MessageBus. Topic: %s, Request-id: %s Correlation-id: %s", response.ReceivedTopic, response.RequestID, response.CorrelationID)

	// Copy before mutating ReceivedTopic: concurrent workers must not share the same *MessageEnvelope instance
	// if the messaging mock returns a reused pointer (tests) or the stack ever does.
	out := *response
	out.ReceivedTopic = externalResponseTopic
	publishMessage(p.getMQTTClient(), externalResponseTopic, qos, retain, out, lc)
}

func publishMessage(client mqtt.Client, responseTopic string, qos byte, retain bool, message types.MessageEnvelope, lc logger.LoggingClient) {
	if message.ErrorCode == 1 {
		lc.Errorf("%v", message.Payload)
	}

	envelopeBytes, _ := json.Marshal(&message)

	if token := client.Publish(responseTopic, qos, retain, envelopeBytes); token.Wait() && token.Error() != nil {
		lc.Errorf("Could not publish to external message broker on topic '%s': %s", responseTopic, token.Error())
	} else {
		lc.Debugf("Published response message to external message broker on topic '%s' with %d bytes", responseTopic, len(envelopeBytes))
	}
}
